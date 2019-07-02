package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/vault"
)

// Formulator manages balance of accounts of the chain
type Formulator struct {
	*chain.ProcessBase
	cn            *chain.Chain
	vault         *vault.Vault
	genesisPolicy *FormulatorPolicy
}

// NewFormulator returns a Formulator
func NewFormulator(genesisPolicy *FormulatorPolicy) *Formulator {
	p := &Formulator{
		genesisPolicy: genesisPolicy,
	}
	return p
}

// Name returns the name of the process
func (p *Formulator) Name() string {
	return "fleta.formulator"
}

// Version returns the version of the process
func (p *Formulator) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Formulator) Init(reg *chain.Register, cn *chain.Chain) error {
	p.cn = cn
	reg.RegisterAccount(1, &FormulatorAccount{})
	if vp, err := cn.ProcessByName("fleta.vault"); err != nil {
		return ErrNotExistVault
	} else if v, is := vp.(*vault.Vault); !is {
		return ErrNotExistVault
	} else {
		p.vault = v
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Formulator) OnLoadChain(loader types.LoaderProcess) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Formulator) BeforeExecuteTransactions(ctp *types.ContextProcess) error {
	if ctp.TargetHeight() == 1 {
		if data, err := encoding.Marshal(&p.genesisPolicy); err != nil {
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
	}
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Formulator) AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error {
	policy := &FormulatorPolicy{}
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
		acc, err := ctp.Account(b.Header.Generator)
		if err != nil {
			return err
		}

		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return ErrInvalidAccountType
		}
		switch frAcc.FormulatorType {
		case AlphaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.AlphaEfficiency1000)).DivC(1000))
		case SigmaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.SigmaEfficiency1000)).DivC(1000))
		case OmegaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.OmegaEfficiency1000)).DivC(1000))
		case HyperFormulatorType:
			PowerSum := frAcc.Amount.MulC(int64(policy.HyperEfficiency1000)).DivC(1000)

			keys, err := ctp.AccountDataKeys(b.Header.Generator, tagStaking)
			if err != nil {
				return err
			}
			for _, k := range keys {
				if StakingAddress, is := fromStakingKey(k); is {
					bs := ctp.AccountData(b.Header.Generator, k)
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

						if bs := ctp.AccountData(b.Header.Generator, toAutoStakingKey(StakingAddress)); len(bs) > 0 && bs[0] == 1 {
							rd.addStakingPower(b.Header.Generator, StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(StakingPower)
						} else {
							rd.addRewardPower(StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(ComissionPower)
						}
					}
				}
			}
			rd.addRewardPower(b.Header.Generator, PowerSum)
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
				frAcc := acc.(*FormulatorAccount)
				if err := p.cn.SwitchProcess(ctp, p.vault, func(stp *types.ContextProcess) error {
					if err := p.vault.AddBalance(stp, frAcc.Address(), PowerSum.Mul(Ratio).Div(amount.COIN)); err != nil {
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
func (p *Formulator) OnSaveData(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}
