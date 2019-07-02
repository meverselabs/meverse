package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
)

// Vault manages balance of accounts of the chain
type Vault struct {
	*chain.ProcessBase
	cn *chain.Chain
}

// NewVault returns a Vault
func NewVault() *Vault {
	p := &Vault{}
	return p
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
func (p *Vault) Init(reg *chain.Register, cn *chain.Chain) error {
	p.cn = cn
	return nil
}

// Balance returns balance of the account of the address
func (p *Vault) Balance(ctp *chain.ContextProcess, addr common.Address) *amount.Amount {
	var total *amount.Amount
	if bs := ctp.ProcessData(addr[:]); len(bs) > 0 {
		total = amount.NewAmountFromBytes(bs)
	} else {
		total = amount.NewCoinAmount(0, 0)
	}
	return total
}

// AddBalance adds balance to the account of the address
func (p *Vault) AddBalance(ctp *chain.ContextProcess, addr common.Address, am *amount.Amount) error {
	zero := amount.NewCoinAmount(0, 0)
	if am.Less(zero) {
		return ErrMinusInput
	}
	total := p.Balance(ctp, addr)
	total = total.Add(am)
	ctp.SetProcessData(addr[:], total.Bytes())
	return nil
}

// SubBalance subtracts balance to the account of the address
func (p *Vault) SubBalance(ctp *chain.ContextProcess, addr common.Address, am *amount.Amount) error {
	total := p.Balance(ctp, addr)
	if total.Less(am) {
		return ErrMinusBalance
	}
	total = total.Sub(am)
	if total.IsZero() {
		ctp.SetProcessData(addr[:], nil)
	} else {
		ctp.SetProcessData(addr[:], total.Bytes())
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Vault) OnLoadChain(loader chain.LoaderProcess) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Vault) BeforeExecuteTransactions(ctp *chain.ContextProcess) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Vault) AfterExecuteTransactions(b *types.Block, ctp *chain.ContextProcess) error {
	return nil
}

// ProcessReward called when required to process reward to the context
func (p *Vault) ProcessReward(b *types.Block, ctp *chain.ContextProcess) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Vault) OnSaveData(b *types.Block, ctp *chain.ContextProcess) error {
	return nil
}
