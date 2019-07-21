package admin

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// Admin manages balance of accounts of the chain
type Admin struct {
	*types.ProcessBase
	pid uint8
	pm  types.ProcessManager
	cn  types.Provider
}

// NewAdmin returns a Admin
func NewAdmin(pid uint8) *Admin {
	p := &Admin{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Admin) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Admin) Name() string {
	return "fleta.admin"
}

// Version returns the version of the process
func (p *Admin) Version() string {
	return "0.0.1"
}

// InitAdmin called at OnInitGenesis of an application
func (p *Admin) InitAdmin(ctw *types.ContextWrapper, addrMap map[string]common.Address) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	for name, adminAddr := range addrMap {
		ctw.SetProcessData(toAdminAddressKey(name), adminAddr[:])
	}
	return nil
}

// Init initializes the process
func (p *Admin) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Admin) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Admin) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Admin) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Admin) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
