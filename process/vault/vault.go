package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/apiserver"
)

// Vault manages balance of accounts of the chain
type Vault struct {
	*types.ProcessBase
	pid uint8
	pm  types.ProcessManager
	cn  types.Provider
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
	reg.RegisterTransaction(3, &Withdraw{})
	reg.RegisterTransaction(4, &CreateAccount{})
	reg.RegisterTransaction(5, &CreateMultiAccount{})
	reg.RegisterTransaction(6, &Assign{})
	reg.RegisterTransaction(7, &Deposit{})
	reg.RegisterTransaction(8, &OpenAccount{})

	if vs, err := pm.ServiceByName("fleta.apiserver"); err != nil {
		//ignore when not loaded
	} else if v, is := vs.(*apiserver.APIServer); !is {
		//ignore when not loaded
	} else {
		s, err := v.JRPC("vault")
		if err != nil {
			return err
		}
		s.Set("balance", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
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
			ctw := cn.NewContextWrapper(p.ID())
			return p.Balance(ctw, addr), nil
		})
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Vault) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Vault) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Vault) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	keys, err := ctw.ProcessDataKeys(toLockedBalancePrefix(b.Header.Height))
	if err != nil {
		return err
	}
	for _, k := range keys {
		if addr, is := fromLockedBalancePrefix(k); is {
			am := p.LockedBalance(ctw, addr, b.Header.Height)
			if err := p.AddBalance(ctw, addr, am); err != nil {
				return err
			}
			ctw.SetProcessData(k, nil)
			ctw.SetProcessData(toLockedBalanceSumKey(addr), p.LockedBalanceTotal(ctw, addr).Sub(am).Bytes())
		}
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Vault) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
