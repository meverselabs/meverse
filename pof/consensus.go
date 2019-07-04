package pof

import (
	"bytes"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Consensus implements the proof of formulator algorithm
type Consensus struct {
	sync.Mutex
	*chain.ConsensusBase
	cn                     *chain.Chain
	ct                     chain.Committer
	MaxBlocksPerFormulator uint32
	blocksBySameFormulator uint32
	observerKeyMap         *types.PublicHashBoolMap
	rt                     *RankTable
}

// NewConsensus returns a Consensus
func NewConsensus(MaxBlocksPerFormulator uint32, ObserverKeys []common.PublicHash) *Consensus {
	ObserverKeyMap := types.NewPublicHashBoolMap()
	for _, pubhash := range ObserverKeys {
		ObserverKeyMap.Put(pubhash.Clone(), true)
	}
	cs := &Consensus{
		MaxBlocksPerFormulator: MaxBlocksPerFormulator,
		observerKeyMap:         ObserverKeyMap,
		rt:                     NewRankTable(),
	}
	return cs
}

// Init initializes the consensus
func (cs *Consensus) Init(cn *chain.Chain, ct chain.Committer) error {
	cs.cn = cn
	cs.ct = ct
	return nil
}

// InitGenesis initializes genesis data
func (cs *Consensus) InitGenesis(ctw *types.ContextWrapper) error {
	cs.Lock()
	defer cs.Unlock()

	if err := cs.updateFormulatorList(ctw); err != nil {
		return err
	}
	if data, err := cs.buildSaveData(); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagState, data)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *Consensus) OnLoadChain(loader types.LoaderWrapper) error {
	cs.Lock()
	defer cs.Unlock()

	dec := encoding.NewDecoder(bytes.NewReader(loader.ProcessData(tagState)))
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
	TimeoutCount, err := cs.decodeConsensusData(bh.ConsensusData)
	if err != nil {
		return err
	}

	Top, err := cs.rt.TopRank(int(TimeoutCount))
	if err != nil {
		return err
	}
	if Top.Address != bh.Generator {
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
	KeyMap := map[common.PublicHash]bool{}
	cs.observerKeyMap.EachAll(func(pubhash common.PublicHash, value bool) bool {
		KeyMap[pubhash] = true
		return true
	})
	bs := types.BlockSign{
		HeaderHash:         encoding.Hash(bh),
		GeneratorSignature: sigs[0],
	}
	ObserverSignatures := sigs[1:]
	if err := common.ValidateSignaturesMajority(encoding.Hash(bs), ObserverSignatures, KeyMap); err != nil {
		return err
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *Consensus) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	cs.Lock()
	defer cs.Unlock()

	HeaderHash := encoding.Hash(b.Header)

	TimeoutCount, err := cs.decodeConsensusData(b.Header.ConsensusData)
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

	if err := cs.updateFormulatorList(ctw); err != nil {
		return err
	}
	if data, err := cs.buildSaveData(); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagState, data)
	}
	return nil
}
