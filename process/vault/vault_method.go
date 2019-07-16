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
	if bs := ctw.AccountData(addr, tagBalance); len(bs) > 0 {
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
	//log.Println("AddBalance", ctw.TargetHeight(), addr.String(), am.String(), p.Balance(ctw, addr).Add(am).String())
	ctw.SetAccountData(addr, tagBalance, p.Balance(ctw, addr).Add(am).Bytes())
	return nil
}

// SubBalance subtracts balance to the account of the address
func (p *Vault) SubBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount) error {
	ctw = ctw.Switch(p.pid)

	total := p.Balance(ctw, addr)
	if total.Less(am) {
		return ErrMinusBalance
	}
	//log.Println("SubBalance", ctw.TargetHeight(), addr.String(), am.String(), p.Balance(ctw, addr).Sub(am).String())

	total = total.Sub(am)
	if total.IsZero() {
		ctw.SetAccountData(addr, tagBalance, nil)
	} else {
		ctw.SetAccountData(addr, tagBalance, total.Bytes())
	}
	return nil
}

// AddLockedBalance adds locked balance to the account of the address
func (p *Vault) AddLockedBalance(ctw *types.ContextWrapper, addr common.Address, UnlockedHeight uint32, am *amount.Amount) error {
	ctw = ctw.Switch(p.pid)

	zero := amount.NewCoinAmount(0, 0)
	if am.Less(zero) {
		return ErrMinusInput
	}
	ctw.SetProcessData(toLockedBalanceKey(UnlockedHeight, addr), p.LockedBalance(ctw, addr, UnlockedHeight).Add(am).Bytes())
	ctw.SetProcessData(toLockedBalanceSumKey(addr), p.LockedBalanceTotal(ctw, addr).Add(am).Bytes())
	return nil
}

// LockedBalance returns locked balance of the account of the address
func (p *Vault) LockedBalance(ctw *types.ContextWrapper, addr common.Address, UnlockedHeight uint32) *amount.Amount {
	ctw = ctw.Switch(p.pid)

	if bs := ctw.ProcessData(toLockedBalanceKey(UnlockedHeight, addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

// LockedBalanceTotal returns all locked balance of the account of the address
func (p *Vault) LockedBalanceTotal(ctw *types.ContextWrapper, addr common.Address) *amount.Amount {
	ctw = ctw.Switch(p.pid)

	if bs := ctw.ProcessData(toLockedBalanceSumKey(addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}
