package pof

import (
	"bytes"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/vault"
)

// Consensus implements the proof of formulation algorithm
type Consensus struct {
	sync.Mutex
	*chain.ConsensusBase
	vault                  *vault.Vault
	cn                     *chain.Chain
	ct                     chain.Committer
	MaxBlocksPerFormulator uint32
	blocksBySameFormulator uint32
	observerKeyMap         *types.PublicHashBoolMap
	rt                     *RankTable
	genesisPolicy          *ConsensusPolicy
}

// NewConsensus returns a Consensus
func NewConsensus(genesisPolicy *ConsensusPolicy, MaxBlocksPerFormulator uint32, ObserverKeys []common.PublicHash) *Consensus {
	ObserverKeyMap := types.NewPublicHashBoolMap()
	for _, pubhash := range ObserverKeys {
		ObserverKeyMap.Put(pubhash.Clone(), true)
	}
	cs := &Consensus{
		MaxBlocksPerFormulator: MaxBlocksPerFormulator,
		observerKeyMap:         ObserverKeyMap,
		rt:                     NewRankTable(),
		genesisPolicy:          genesisPolicy,
	}
	return cs
}

// Init initializes the consensus
func (cs *Consensus) Init(reg *chain.Register, cn *chain.Chain, ct chain.Committer) error {
	cs.cn = cn
	cs.ct = ct
	reg.RegisterAccount(1, &FormulationAccount{})
	p, err := cn.ProcessByName("fleta.vault")
	if err != nil {
		return ErrNotExistVault
	}
	if v, is := p.(*vault.Vault); !is {
		return ErrNotExistVault
	} else {
		cs.vault = v
	}
	return nil
}

// InitGenesis initializes genesis data
func (cs *Consensus) InitGenesis(ctp *types.ContextProcess) error {
	cs.Lock()
	defer cs.Unlock()

	var inErr error
	phase := cs.rt.largestPhase() + 1
	ctp.Top().AccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if a.Address().Coordinate().Height == ctp.TargetHeight() {
			if acc, is := a.(*FormulationAccount); is {
				addr := acc.Address()
				if err := cs.rt.addRank(NewRank(addr, acc.KeyHash, phase, hash.DoubleHash(addr[:]))); err != nil {
					inErr = err
					return false
				}
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctp.Top().DeletedAccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if acc, is := a.(*FormulationAccount); is {
			cs.rt.removeRank(acc.Address())
		}
		return true
	})
	if data, err := cs.buildSaveData(); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("state"), data)
	}

	if data, err := encoding.Marshal(&cs.genesisPolicy); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("policy"), data)
	}

	rd := newRewardData()
	if data, err := encoding.Marshal(&rd); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("reward"), data)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *Consensus) OnLoadChain(loader types.LoaderProcess) error {
	cs.Lock()
	defer cs.Unlock()

	dec := encoding.NewDecoder(bytes.NewReader(loader.ProcessData([]byte("state"))))
	if v, err := dec.DecodeUint32(); err != nil {
		return err
	} else {
		if cs.MaxBlocksPerFormulator != v {
			return ErrInvalidMaxBlocksPerFormulator
		}
	}
	ObserverKeyMap := types.NewPublicHashBoolMap()
	if err := dec.Decode(&ObserverKeyMap); err != nil {
		return err
	} else {
		if ObserverKeyMap.Len() != cs.observerKeyMap.Len() {
			return ErrInvalidObserverKey
		}
		var inErr error
		ObserverKeyMap.EachAll(func(pubhash common.PublicHash, value bool) bool {
			if !cs.observerKeyMap.Has(pubhash) {
				inErr = ErrInvalidObserverKey
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	}
	if v, err := dec.DecodeUint32(); err != nil {
		return err
	} else {
		cs.blocksBySameFormulator = v
	}
	if err := dec.Decode(&cs.rt); err != nil {
		return err
	}
	return nil
}

// ValidateSignature called when required to validate signatures
func (cs *Consensus) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
	TimeoutCount, Formulator, err := cs.decodeConsensusData(bh.ConsensusData)
	if err != nil {
		return err
	}

	Top, err := cs.rt.TopRank(int(TimeoutCount))
	if err != nil {
		return err
	}
	if Top.Address != Formulator {
		return ErrInvalidTopAddress
	}

	GeneratorSignature := sigs[0]
	pubkey, err := common.RecoverPubkey(encoding.Hash(bh), GeneratorSignature)
	if err != nil {
		return err
	}
	pubhash := common.NewPublicHash(pubkey)
	if Top.PublicHash != pubhash {
		return ErrInvalidTopSignature
	}

	if len(sigs) != cs.observerKeyMap.Len()/2+2 {
		return ErrInvalidSignatureCount
	}
	s := &ObserverSigned{
		BlockSign: types.BlockSign{
			HeaderHash:         encoding.Hash(bh),
			GeneratorSignature: sigs[0],
		},
		ObserverSignatures: sigs[1:],
	}

	KeyMap := map[common.PublicHash]bool{}
	cs.observerKeyMap.EachAll(func(pubhash common.PublicHash, value bool) bool {
		KeyMap[pubhash] = true
		return true
	})
	if err := common.ValidateSignaturesMajority(encoding.Hash(s.BlockSign), s.ObserverSignatures, KeyMap); err != nil {
		return err
	}
	return nil
}

// ProcessReward called when required to process reward to the context
func (cs *Consensus) ProcessReward(b *types.Block, ctp *types.ContextProcess) error {
	_, Formulator, err := cs.decodeConsensusData(b.Header.ConsensusData)
	if err != nil {
		return err
	}

	policy := &ConsensusPolicy{}
	if bs := ctp.ProcessData([]byte("policy")); len(bs) == 0 {
		return ErrInvalidRewardData
	} else if err := encoding.Unmarshal(bs, &policy); err != nil {
		return err
	}

	rd := newRewardData()
	if bs := ctp.ProcessData([]byte("reward")); len(bs) == 0 {
		return ErrInvalidRewardData
	} else if err := encoding.Unmarshal(bs, &rd); err != nil {
		return err
	}

	if true {
		acc, err := ctp.Account(Formulator)
		if err != nil {
			return err
		}

		frAcc, is := acc.(*FormulationAccount)
		if !is {
			return ErrInvalidAccountType
		}
		switch frAcc.FormulationType {
		case AlphaFormulatorType:
			rd.addRewardPower(Formulator, frAcc.Amount.MulC(int64(policy.AlphaEfficiency1000)).DivC(1000))
		case SigmaFormulatorType:
			rd.addRewardPower(Formulator, frAcc.Amount.MulC(int64(policy.SigmaEfficiency1000)).DivC(1000))
		case OmegaFormulatorType:
			rd.addRewardPower(Formulator, frAcc.Amount.MulC(int64(policy.OmegaEfficiency1000)).DivC(1000))
		case HyperFormulatorType:
			PowerSum := frAcc.Amount.MulC(int64(policy.HyperEfficiency1000)).DivC(1000)

			keys, err := ctp.AccountDataKeys(Formulator, tagStaking)
			if err != nil {
				return err
			}
			for _, k := range keys {
				if StakingAddress, is := fromStakingKey(k); is {
					bs := ctp.AccountData(Formulator, k)
					if len(bs) == 0 {
						return ErrInvalidStakingAddress
					}
					StakingAmount := amount.NewAmountFromBytes(bs)

					if _, err := ctp.Account(StakingAddress); err != nil {
						if err != types.ErrNotExistAccount {
							return err
						}
						rd.removeRewardPower(StakingAddress)
					} else {
						StakingPower := StakingAmount.MulC(int64(policy.StakingEfficiency1000)).DivC(1000)
						ComissionPower := StakingPower.MulC(int64(frAcc.Policy.CommissionRatio1000)).DivC(1000)

						if bs := ctp.AccountData(Formulator, toAutoStakingKey(StakingAddress)); len(bs) > 0 && bs[0] == 1 {
							rd.addStakingPower(Formulator, StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(StakingPower)
						} else {
							rd.addRewardPower(StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(ComissionPower)
						}
					}
				}
			}
			rd.addRewardPower(Formulator, PowerSum)
		default:
			return ErrInvalidAccountType
		}
	}

	if ctp.TargetHeight() >= rd.lastPaidHeight+policy.PayRewardEveryBlocks {
		TotalPower := amount.NewCoinAmount(0, 0)
		rd.powerMap.EachAll(func(addr common.Address, PowerSum *amount.Amount) bool {
			TotalPower = TotalPower.Add(PowerSum)
			return true
		})
		TotalReward := policy.RewardPerBlock.MulC(int64(ctp.TargetHeight() - rd.lastPaidHeight))
		Ratio := TotalReward.Mul(amount.COIN).Div(TotalPower)
		var inErr error
		rd.powerMap.EachAll(func(RewardAddress common.Address, PowerSum *amount.Amount) bool {
			acc, err := ctp.Account(RewardAddress)
			if err != nil {
				if err != types.ErrNotExistAccount {
					inErr = err
					return false
				}
			} else {
				frAcc := acc.(*FormulationAccount)
				if err := cs.cn.SwitchProcess(ctp, cs.vault, func(stp *types.ContextProcess) error {
					if err := cs.vault.AddBalance(stp, frAcc.Address(), PowerSum.Mul(Ratio).Div(amount.COIN)); err != nil {
						return err
					}
					return nil
				}); err != nil {
					inErr = err
					return false
				}
			}
			rd.removeRewardPower(RewardAddress)
			return true
		})
		if inErr != nil {
			return inErr
		}

		rd.stakingPowerMap.EachAll(func(HyperAddress common.Address, PowerMap *types.AddressAmountMap) bool {
			PowerMap.EachAll(func(StakingAddress common.Address, PowerSum *amount.Amount) bool {
				bs := ctp.AccountData(HyperAddress, toStakingKey(StakingAddress))
				if len(bs) == 0 {
					inErr = ErrInvalidStakingAddress
					return false
				}
				StakingAmount := amount.NewAmountFromBytes(bs)
				ctp.SetAccountData(HyperAddress, toStakingKey(StakingAddress), StakingAmount.Add(PowerSum.Mul(Ratio).Div(amount.COIN)).Bytes())
				return true
			})
			if inErr != nil {
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
		rd.stakingPowerMap = types.NewAddressAddressAmountMap()

		//log.Println("Paid at", ctp.TargetHeight())
		rd.lastPaidHeight = ctp.TargetHeight()
	}

	if data, err := encoding.Marshal(&policy); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("policy"), data)
	}

	if data, err := encoding.Marshal(&rd); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("reward"), data)
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *Consensus) OnSaveData(b *types.Block, ctp *types.ContextProcess) error {
	cs.Lock()
	defer cs.Unlock()

	HeaderHash := encoding.Hash(b.Header)

	TimeoutCount, _, err := cs.decodeConsensusData(b.Header.ConsensusData)
	if err != nil {
		return err
	}
	if TimeoutCount > 0 {
		if err := cs.rt.forwardCandidates(int(TimeoutCount)); err != nil {
			return err
		}
		cs.blocksBySameFormulator = 0
	}
	cs.blocksBySameFormulator++
	if cs.blocksBySameFormulator >= cs.MaxBlocksPerFormulator {
		cs.rt.forwardTop(HeaderHash)
		cs.blocksBySameFormulator = 0
	}

	var inErr error
	phase := cs.rt.largestPhase() + 1
	ctp.Top().AccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if a.Address().Coordinate().Height == ctp.TargetHeight() {
			if acc, is := a.(*FormulationAccount); is {
				addr := acc.Address()
				if err := cs.rt.addRank(NewRank(addr, acc.KeyHash, phase, hash.DoubleHash(addr[:]))); err != nil {
					inErr = err
					return false
				}
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctp.Top().DeletedAccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if acc, is := a.(*FormulationAccount); is {
			cs.rt.removeRank(acc.Address())
		}
		return true
	})

	if data, err := cs.buildSaveData(); err != nil {
		return err
	} else {
		ctp.SetProcessData([]byte("state"), data)
	}
	return nil
}

func (cs *Consensus) decodeConsensusData(ConsensusData []byte) (uint32, common.Address, error) {
	dec := encoding.NewDecoder(bytes.NewReader(ConsensusData))
	TimeoutCount, err := dec.DecodeUint32()
	if err != nil {
		return 0, common.Address{}, err
	}
	var Formulator common.Address
	if err := dec.Decode(&Formulator); err != nil {
		return 0, common.Address{}, err
	}
	return TimeoutCount, Formulator, nil
}

func (cs *Consensus) encodeConsensusData(TimeoutCount uint32, Formulator common.Address) ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(TimeoutCount); err != nil {
		return nil, err
	}
	if err := enc.Encode(Formulator); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (cs *Consensus) buildSaveData() ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(cs.MaxBlocksPerFormulator); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.observerKeyMap); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint32(cs.blocksBySameFormulator); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.rt); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
