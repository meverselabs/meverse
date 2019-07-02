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
	ValidateSignature(bh *types.Header, sigs []common.Signature) error
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

// ValidateSignature called when required to validate signatures
func (cs *ConsensusBase) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
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
