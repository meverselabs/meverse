package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Process is a interface of the chain Process
type Process interface {
	Version() string
	Name() string
	Init(reg *Register, cn *Chain) error
	OnLoadChain(loader types.LoaderProcess) error
	BeforeExecuteTransactions(ctp *types.ContextProcess) error
	AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error
	OnSaveData(b *types.Block, ctp *types.ContextProcess) error
}

// ProcessBase is a base handler of the chain Process
type ProcessBase struct{}

// OnLoadChain called when the chain loaded
func (p *ProcessBase) OnLoadChain(loader types.LoaderProcess) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *ProcessBase) BeforeExecuteTransactions(ctp *types.ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *ProcessBase) AfterExecuteTransactions(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *ProcessBase) OnSaveData(b *types.Block, ctp *types.ContextProcess) error {
	return nil
}
