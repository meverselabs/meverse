package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Process is a interface of the chain Process
type Process interface {
	Version() string
	Name() string
	Init(reg *Register, cn *Chain) error
	OnLoadChain(loader LoaderProcess) error
	BeforeExecuteTransactions(ctp *ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *ContextProcess) error
	ProcessReward(b *types.Block, ctp *ContextProcess) error
	OnSaveData(b *types.Block, ctp *ContextProcess) error
}

// ProcessBase is a base handler of the chain Process
type ProcessBase struct{}

// OnLoadChain called when the chain loaded
func (p *ProcessBase) OnLoadChain(loader LoaderProcess) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *ProcessBase) BeforeExecuteTransactions(ctp *ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *ProcessBase) AfterExecuteTransactions(b *types.Block, ctp *ContextProcess) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (p *ProcessBase) ProcessReward(b *types.Block, ctp *ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *ProcessBase) OnSaveData(b *types.Block, ctp *ContextProcess) error {
	return nil
}
