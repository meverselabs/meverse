package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Consensus is a interface of the chain Consensus
type Consensus interface {
	Init(cn *Chain, ct Committer) error
	InitGenesis(ctp *types.ContextProcess) error
	ValidateHeader(bh *types.Header) error
	ProcessReward(b *types.Block, ctp *types.ContextProcess) error
}

// ConsensusBase is a base handler of the chain Consensus
type ConsensusBase struct{}

// InitGenesis initializes genesis data
func (cs *ConsensusBase) InitGenesis(ctp *types.ContextProcess) error {
	return nil
}

// ValidateHeader called when required to validate the header
func (cs *ConsensusBase) ValidateHeader(bh *types.Header) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (cs *ConsensusBase) ProcessReward(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}
