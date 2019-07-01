package pof

import (
	"bytes"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Consensus implements the proof of formulation algorithm
type Consensus struct {
	sync.Mutex
	*chain.ConsensusBase
	cn                     *chain.Chain
	ct                     chain.Committer
	rt                     *RankTable
	MaxBlocksPerFormulator uint32
	blocksBySameFormulator uint32
	keyMap                 map[common.PublicHash]bool
	ObserverKeyMap         *types.PublicHashBoolMap
}

// NewConsensus returns a Consensus
func NewConsensus(ObserverKeyMap *types.PublicHashBoolMap, MaxBlocksPerFormulator uint32) *Consensus {
	KeyMap := map[common.PublicHash]bool{}
	ObserverKeyMap.EachAll(func(pubhash common.PublicHash, value bool) bool {
		KeyMap[pubhash] = true
		return true
	})
	cs := &Consensus{
		rt:                     NewRankTable(),
		MaxBlocksPerFormulator: MaxBlocksPerFormulator,
		keyMap:                 KeyMap,
		ObserverKeyMap:         ObserverKeyMap,
	}
	return cs
}

// Init initializes the consensus
func (cs *Consensus) Init(reg *chain.Register, cn *chain.Chain, ct chain.Committer) error {
	cs.cn = cn
	cs.ct = ct
	reg.RegisterAccount(1, &FormulationAccount{})
	return nil
}

// InitGenesis initializes genesis data
func (cs *Consensus) InitGenesis(ctp *chain.ContextProcess) error {
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
	SaveData, err := cs.buildSaveData()
	if err != nil {
		return err
	}
	ctp.SetProcessData([]byte("state"), SaveData)
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *Consensus) OnLoadChain(loader chain.LoaderProcess) error {
	if err := cs.loadFromSaveData(loader.ProcessData([]byte("state"))); err != nil {
		return err
	}
	return nil
}

// ValidateHeader called when required to validate the header
func (cs *Consensus) ValidateHeader(bh *types.Header, sigs []common.Signature) error {
	dec := encoding.NewDecoder(bytes.NewReader(bh.ConsensusData))
	TimeoutCount, err := dec.DecodeUint16()
	if err != nil {
		return err
	}
	Top, err := cs.rt.TopRank(int(TimeoutCount))
	if err != nil {
		return err
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
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (cs *Consensus) BeforeExecuteTransactions(b *types.Block, ctp *chain.ContextProcess) error {
	if len(b.Signatures) != cs.ObserverKeyMap.Len()/2+2 {
		return ErrInvalidSignatureCount
	}
	s := &ObserverSigned{
		BlockSign: types.BlockSign{
			HeaderHash:         encoding.Hash(b.Header),
			GeneratorSignature: b.Signatures[0],
		},
		ObserverSignatures: b.Signatures[1:],
	}
	if err := common.ValidateSignaturesMajority(encoding.Hash(s.BlockSign), s.ObserverSignatures, cs.keyMap); err != nil {
		return err
	}
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (cs *Consensus) AfterExecuteTransactions(b *types.Block, ctp *chain.ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *Consensus) OnSaveData(b *types.Block, ctp *chain.ContextProcess) error {
	cs.Lock()
	defer cs.Unlock()

	HeaderHash := encoding.Hash(b.Header)

	dec := encoding.NewDecoder(bytes.NewReader(b.Header.ConsensusData))
	TimeoutCount, err := dec.DecodeUint16()
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

	SaveData, err := cs.buildSaveData()
	if err != nil {
		return err
	}
	ctp.SetProcessData([]byte("state"), SaveData)
	return nil
}

// ProcessReward called when required to process reward to the context
func (cs *Consensus) ProcessReward(b *types.Block, ctp *chain.ContextProcess) error {
	return nil
}

func (cs *Consensus) buildSaveData() ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(cs.MaxBlocksPerFormulator); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint32(cs.blocksBySameFormulator); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.rt); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.ObserverKeyMap); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (cs *Consensus) loadFromSaveData(SaveData []byte) error {
	cs.Lock()
	defer cs.Unlock()

	dec := encoding.NewDecoder(bytes.NewReader(SaveData))
	if v, err := dec.DecodeUint32(); err != nil {
		return err
	} else {
		if cs.MaxBlocksPerFormulator != v {
			return ErrInvalidMaxBlocksPerFormulator
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
	ObserverKeyMap := types.NewPublicHashBoolMap()
	if err := dec.Decode(&ObserverKeyMap); err != nil {
		return err
	} else {
		if ObserverKeyMap.Len() != cs.ObserverKeyMap.Len() {
			return ErrInvalidObserverKey
		}
		var inErr error
		ObserverKeyMap.EachAll(func(pubhash common.PublicHash, value bool) bool {
			if !cs.ObserverKeyMap.Has(pubhash) {
				inErr = ErrInvalidObserverKey
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	}
	return nil
}
