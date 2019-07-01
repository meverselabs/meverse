package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Consensus is a interface of the chain Consensus
type Consensus interface {
	Init(cn *Chain, ct Committer, loader LoaderProcess) error
	InitGenesis(ctp *ContextProcess) error
	ValidateHeader(bh *types.Header) error
	BeforeExecuteTransactions(b *types.Block, ctp *ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *ContextProcess) error
	ProcessReward(b *types.Block, ctp *ContextProcess) error
}

// ConsensusBase is a base handler of the chain Consensus
type ConsensusBase struct{}

// InitGenesis initializes genesis data
func (cs *ConsensusBase) InitGenesis(ctp *ContextProcess) error {
	return nil
}

// ValidateHeader called when required to validate the header
func (cs *ConsensusBase) ValidateHeader(bh *types.Header) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (cs *ConsensusBase) BeforeExecuteTransactions(cd *types.Block, ctp *ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (cs *ConsensusBase) AfterExecuteTransactions(cd *types.Block, ctp *ContextProcess) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (cs *ConsensusBase) ProcessReward(b *types.Block, ctp *ContextProcess) error {
	return nil
}
