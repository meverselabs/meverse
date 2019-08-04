package gateway

import (
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/vault"
)

// Gateway manages balance of accounts of the chain
type Gateway struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	vault *vault.Vault
	admin *admin.Admin
}

// NewGateway returns a Gateway
func NewGateway(pid uint8) *Gateway {
	p := &Gateway{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Gateway) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Gateway) Name() string {
	return "fleta.gateway"
}

// Version returns the version of the process
func (p *Gateway) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Gateway) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
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

	reg.RegisterTransaction(1, &TokenIn{})
	reg.RegisterTransaction(2, &TokenOut{})
	reg.RegisterTransaction(3, &TokenLeave{})
	reg.RegisterTransaction(4, &UpdatePolicy{})
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Gateway) InitPolicy(ctw *types.ContextWrapper, policy *Policy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs, err := encoding.Marshal(policy); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagPolicy, bs)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Gateway) OnLoadChain(loader types.LoaderWrapper) error {
	p.admin.AdminAddress(loader, p.Name())
	if bs := loader.ProcessData(tagPolicy); len(bs) == 0 {
		return ErrPolicyShouldBeSetupInApplication
	}
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Gateway) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Gateway) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Gateway) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
