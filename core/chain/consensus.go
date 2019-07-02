package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// Consensus is a interface of the chain Consensus
type Consensus interface {
	Init(reg *Register, cn *Chain, ct Committer) error
	InitGenesis(ctp *ContextProcess) error
	OnLoadChain(loader LoaderProcess) error
	ValidateHeader(bh *types.Header, sigs []common.Signature) error
	BeforeExecuteTransactions(ctp *ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *ContextProcess) error
	ProcessReward(b *types.Block, ctp *ContextProcess) error
	OnSaveData(b *types.Block, ctp *ContextProcess) error
}

// ConsensusBase is a base handler of the chain Consensus
type ConsensusBase struct{}

// InitGenesis initializes genesis data
func (cs *ConsensusBase) InitGenesis(ctp *ContextProcess) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (cs *ConsensusBase) OnLoadChain(loader LoaderProcess) error {
	return nil
}

// ValidateHeader called when required to validate the header
func (cs *ConsensusBase) ValidateHeader(bh *types.Header, sigs []common.Signature) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (cs *ConsensusBase) BeforeExecuteTransactions(ctp *ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (cs *ConsensusBase) AfterExecuteTransactions(b *types.Block, ctp *ContextProcess) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (cs *ConsensusBase) ProcessReward(b *types.Block, ctp *ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (cs *ConsensusBase) OnSaveData(b *types.Block, ctp *ContextProcess) error {
	return nil
}
