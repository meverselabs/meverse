package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/service/apiserver"
)

// Vault manages balance of accounts of the chain
type Vault struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	admin *admin.Admin
}

// NewVault returns a Vault
func NewVault(pid uint8) *Vault {
	p := &Vault{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Vault) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Vault) Name() string {
	return "fleta.vault"
}

// Version returns the version of the process
func (p *Vault) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Vault) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn
	reg.RegisterAccount(1, &SingleAccount{})
	reg.RegisterAccount(2, &MultiAccount{})
	reg.RegisterTransaction(1, &Transfer{})
	reg.RegisterTransaction(2, &Burn{})
	reg.RegisterTransaction(4, &CreateAccount{})
	reg.RegisterTransaction(5, &CreateMultiAccount{})
	reg.RegisterTransaction(9, &IssueAccount{})
	reg.RegisterTransaction(10, &UpdatePolicy{})

	if vp, err := pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if v, is := vp.(*admin.Admin); !is {
		return types.ErrInvalidProcess
	} else {
		p.admin = v
	}
	if vs, err := pm.ServiceByName("fleta.apiserver"); err != nil {
		//ignore when not loaded
	} else if v, is := vs.(*apiserver.APIServer); !is {
		//ignore when not loaded
	} else {
		s, err := v.JRPC("vault")
		if err != nil {
			return err
		}
		s.Set("getBalance", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 1 {
				return nil, apiserver.ErrInvalidArgument
			}
			arg0, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(arg0)
			if err != nil {
				return nil, err
			}
			loader := cn.NewLoaderWrapper(p.ID())
			return p.Balance(loader, addr), nil
		})
	}
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Vault) InitPolicy(ctw *types.ContextWrapper, policy *Policy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs, err := encoding.Marshal(policy); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagPolicy, bs)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Vault) OnLoadChain(loader types.LoaderWrapper) error {
	p.admin.AdminAddress(loader, p.Name())
	if bs := loader.ProcessData(tagPolicy); len(bs) == 0 {
		return ErrPolicyShouldBeSetupInApplication
	}
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Vault) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Vault) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	LockedBalanceMap, err := p.flushLockedBalanceMap(ctw, b.Header.Height)
	if err != nil {
		return err
	}

	for addr, am := range LockedBalanceMap {
		if has, err := ctw.HasAccount(addr); err != nil {
			if err == types.ErrDeletedAccount {
				ctw.SetAccountData(addr, tagLockedBalanceSum, nil)
			} else {
				return err
			}
		} else if !has {
			ctw.SetAccountData(addr, tagLockedBalanceSum, nil)
		} else {
			if err := p.AddBalance(ctw, addr, am); err != nil {
				return err
			}
			sum := p.TotalLockedBalanceByAddress(ctw, addr).Sub(am)
			if !sum.IsZero() {
				ctw.SetAccountData(addr, tagLockedBalanceSum, sum.Bytes())
			} else {
				ctw.SetAccountData(addr, tagLockedBalanceSum, nil)
			}
		}
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Vault) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
