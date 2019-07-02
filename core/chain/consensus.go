package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// Consensus is a interface of the chain Consensus
type Consensus interface {
	Init(reg *Register, cn *Chain, ct Committer) error
	InitGenesis(ctp *types.ContextProcess) error
	OnLoadChain(loader types.LoaderProcess) error
	ValidateSignature(bh *types.Header, sigs []common.Signature) error
	BeforeExecuteTransactions(ctp *types.ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error
	OnSaveData(b *types.Block, ctp *types.ContextProcess) error
}

// ConsensusBase is a base handler of the chain Consensus
type ConsensusBase struct{}

// InitGenesis initializes genesis data
func (cs *ConsensusBase) InitGenesis(ctp *types.ContextProcess) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *ConsensusBase) OnLoadChain(loader types.LoaderProcess) error {
	return nil
}

// ValidateSignature called when required to validate signatures
func (cs *ConsensusBase) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (cs *ConsensusBase) BeforeExecuteTransactions(bctp *types.ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (cs *ConsensusBase) AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *ConsensusBase) OnSaveData(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}
