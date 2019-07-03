package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
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
	reg.RegisterTransaction(1, &Transfer{})
	return nil
}

// Balance returns balance of the account of the address
func (p *Vault) Balance(ctw *types.ContextWrapper, addr common.Address) *amount.Amount {
	ctw = ctw.Switch(p.pid)

	var total *amount.Amount
	if bs := ctw.ProcessData(addr[:]); len(bs) > 0 {
		total = amount.NewAmountFromBytes(bs)
	} else {
		total = amount.NewCoinAmount(0, 0)
	}
	return total
}

// AddBalance adds balance to the account of the address
func (p *Vault) AddBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount) error {
	ctw = ctw.Switch(p.pid)

	zero := amount.NewCoinAmount(0, 0)
	if am.Less(zero) {
		return ErrMinusInput
	}
	total := p.Balance(ctw, addr)
	total = total.Add(am)
	ctw.SetProcessData(addr[:], total.Bytes())
	return nil
}

// SubBalance subtracts balance to the account of the address
func (p *Vault) SubBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount) error {
	ctw = ctw.Switch(p.pid)

	total := p.Balance(ctw, addr)
	if total.Less(am) {
		return ErrMinusBalance
	}
	total = total.Sub(am)
	if total.IsZero() {
		ctw.SetProcessData(addr[:], nil)
	} else {
		ctw.SetProcessData(addr[:], total.Bytes())
	}
	return nil
}

// AddLockedBalance adds locked balance to the account of the address
func (p *Vault) AddLockedBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount, height uint32) error {
	ctw = ctw.Switch(p.pid)

	panic("TODO")
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
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Vault) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
