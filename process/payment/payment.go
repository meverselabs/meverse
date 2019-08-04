package payment

import (
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/vault"
)

// Payment manages balance of accounts of the chain
type Payment struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	vault *vault.Vault
	admin *admin.Admin
}

// NewPayment returns a Payment
func NewPayment(pid uint8) *Payment {
	p := &Payment{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Payment) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Payment) Name() string {
	return "fleta.payment"
}

// Version returns the version of the process
func (p *Payment) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Payment) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("fleta.vault"); err != nil {
		return err
	} else if v, is := vp.(*vault.Vault); !is {
		return types.ErrInvalidProcess
	} else {
		p.vault = v
	}
	if vp, err := pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if v, is := vp.(*admin.Admin); !is {
		return types.ErrInvalidProcess
	} else {
		p.admin = v
	}

	reg.RegisterTransaction(1, &RequestPayment{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Payment) OnLoadChain(loader types.LoaderWrapper) error {
	p.admin.AdminAddress(loader, p.Name())
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Payment) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Payment) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Payment) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
