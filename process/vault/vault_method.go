package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

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
	ctw.SetAccountData(addr, tagBalance, total.Bytes())
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
		ctw.SetAccountData(addr, tagBalance, nil)
	} else {
		ctw.SetAccountData(addr, tagBalance, total.Bytes())
	}
	return nil
}

// AddLockedBalance adds locked balance to the account of the address
func (p *Vault) AddLockedBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount, height uint32) error {
	ctw = ctw.Switch(p.pid)

	panic("TODO")
}
