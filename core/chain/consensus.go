package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// Consensus defines chain consensus functions
type Consensus interface {
	Init(cn *Chain, ct Committer) error
	InitGenesis(ctw *types.ContextWrapper) error
	OnLoadChain(loader types.LoaderWrapper) error
	ValidateSignature(bh *types.Header, sigs []common.Signature) error
	OnSaveData(b *types.Block, ctw *types.ContextWrapper) error
}

// ConsensusBase is a base handler of the chain consensus
type ConsensusBase struct{}

// InitGenesis initializes genesis data
func (cs *ConsensusBase) InitGenesis(ctw *types.ContextWrapper) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *ConsensusBase) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// ValidateSignature called when required to validate signatures
func (cs *ConsensusBase) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *ConsensusBase) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
