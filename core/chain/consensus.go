package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// Consensus is a interface of the chain Consensus
type Consensus interface {
	Init(cn *Chain, ct Committer) error
	InitGenesis(ctw *types.ContextWrapper) error
	OnLoadChain(loader types.LoaderWrapper) error
	ValidateSignature(bh *types.Header, sigs []common.Signature) error
	BeforeExecuteTransactions(ctw *types.ContextWrapper) error
	AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error
	OnSaveData(b *types.Block, ctw *types.ContextWrapper) error
}

// ConsensusBase is a base handler of the chain Consensus
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

// BeforeExecuteTransactions called before processes transactions of the block
func (cs *ConsensusBase) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (cs *ConsensusBase) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *ConsensusBase) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
