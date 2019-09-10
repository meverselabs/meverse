package poa

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
	cn           *chain.Chain
	ct           chain.Committer
	authorityKey common.PublicHash
}

// NewConsensus returns a Consensus
func NewConsensus(AuthorityKey common.PublicHash) *Consensus {
	cs := &Consensus{
		authorityKey: AuthorityKey,
	}
	return cs
}

// Init initializes the consensus
func (cs *Consensus) Init(cn *chain.Chain, ct chain.Committer) error {
	cs.cn = cn
	cs.ct = ct

	/*
		if vs, err := cn.ServiceByName("fleta.apiserver"); err != nil {
			//ignore when not loaded
		} else if v, is := vs.(*apiserver.APIServer); !is {
			//ignore when not loaded
		} else {
		}
	*/
	return nil
}

// InitGenesis initializes genesis data
func (cs *Consensus) InitGenesis(ctw *types.ContextWrapper) error {
	cs.Lock()
	defer cs.Unlock()

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
	var AuthorityKey common.PublicHash
	if err := dec.Decode(&AuthorityKey); err != nil {
		return err
	} else if cs.authorityKey != AuthorityKey {
		return ErrInvalidAuthorityKey
	}
	return nil
}

// ValidateSignature called when required to validate signatures
func (cs *Consensus) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
	if len(sigs) != 1 {
		return ErrInvalidSignatureCount
	}
	GeneratorSignature := sigs[0]
	pubkey, err := common.RecoverPubkey(encoding.Hash(bh), GeneratorSignature)
	if err != nil {
		return err
	}
	pubhash := common.NewPublicHash(pubkey)
	if cs.authorityKey != pubhash {
		return ErrInvalidAuthorityKey
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *Consensus) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	cs.Lock()
	defer cs.Unlock()

	if data, err := cs.buildSaveData(); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagState, data)
	}
	return nil
}
