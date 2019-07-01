package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Process is a interface of the chain Process
type Process interface {
	Version() string
	Name() string
	Init(reg *Register, cn *Chain) error
	InitGenesis(ctp *types.ContextProcess) error
	ValidateHeader(bh *types.Header) error
	BeforeExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error
	ProcessReward(b *types.Block, ctp *types.ContextProcess) error
}

// ProcessBase is a base handler of the chain Process
type ProcessBase struct{}

// InitGenesis initializes genesis data
func (p *ProcessBase) InitGenesis(ctp *types.ContextProcess) error {
	return nil
}

// ValidateHeader called when required to validate the header
func (p *ProcessBase) ValidateHeader(bh *types.Header) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *ProcessBase) BeforeExecuteTransactions(cd *types.Block, ctp *types.ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *ProcessBase) AfterExecuteTransactions(cd *types.Block, ctp *types.ContextProcess) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (p *ProcessBase) ProcessReward(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}
