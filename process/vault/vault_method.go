package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// Balance returns balance of the account of the address
func (p *Vault) Balance(lw types.LoaderWrapper, addr common.Address) *amount.Amount {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	var total *amount.Amount
	if bs := lw.AccountData(addr, tagBalance); len(bs) > 0 {
		total = amount.NewAmountFromBytes(bs)
	} else {
		total = amount.NewCoinAmount(0, 0)
	}
	return total
}

// AddBalance adds balance to the account of the address
func (p *Vault) AddBalance(ctw *types.ContextWrapper, addr common.Address, am *amount.Amount) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

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
	ctw = types.SwitchContextWrapper(p.pid, ctw)

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
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	zero := amount.NewCoinAmount(0, 0)
	if am.Less(zero) {
		return ErrMinusInput
	}
	ctw.SetProcessData(toLockedBalanceKey(UnlockedHeight, addr), p.LockedBalance(ctw, addr, UnlockedHeight).Add(am).Bytes())
	ctw.SetProcessData(toLockedBalanceSumKey(addr), p.LockedBalanceTotal(ctw, addr).Add(am).Bytes())
	return nil
}

// LockedBalance returns locked balance of the account of the address
func (p *Vault) LockedBalance(lw types.LoaderWrapper, addr common.Address, UnlockedHeight uint32) *amount.Amount {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.ProcessData(toLockedBalanceKey(UnlockedHeight, addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

// LockedBalanceTotal returns all locked balance of the account of the address
func (p *Vault) LockedBalanceTotal(lw types.LoaderWrapper, addr common.Address) *amount.Amount {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.ProcessData(toLockedBalanceSumKey(addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

// CheckFeePayable returns tx fee can be paid or not
func (p *Vault) CheckFeePayable(lw types.LoaderWrapper, tx FeeTransaction) error {
	return p.CheckFeePayableWith(lw, tx, nil)
}

// CheckFeePayableWith returns tx fee and amount can be paid or not
func (p *Vault) CheckFeePayableWith(lw types.LoaderWrapper, tx FeeTransaction, am *amount.Amount) error {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if has, err := lw.HasAccount(tx.From()); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}

	fee := tx.Fee(lw)
	if am != nil {
		am = am.Add(fee)
	} else {
		am = fee
	}

	b := p.Balance(lw, tx.From())
	if b.Less(am) {
		return ErrInsufficientFee
	}
	return nil
}

// WithFee processes function after withdraw fee
func (p *Vault) WithFee(ctw *types.ContextWrapper, tx FeeTransaction, fn func() error) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	fee := tx.Fee(ctw)
	if err := p.SubBalance(ctw, tx.From(), fee); err != nil {
		return err
	}
	ctw.SetProcessData(tagCollectedFee, p.CollectedFee(ctw).Add(fee).Bytes())

	sn := ctw.Snapshot()
	if err := fn(); err != nil {
		ctw.Revert(sn)
		return err
	}
	ctw.Commit(sn)
	return nil
}

// CollectedFee returns a total collected fee
func (p *Vault) CollectedFee(lw types.LoaderWrapper) *amount.Amount {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	var total *amount.Amount
	if bs := lw.ProcessData(tagCollectedFee); len(bs) > 0 {
		total = amount.NewAmountFromBytes(bs)
	} else {
		total = amount.NewCoinAmount(0, 0)
	}
	return total
}

// AddCollectedFee adds collected fee to the account of the address
func (p *Vault) AddCollectedFee(ctw *types.ContextWrapper, am *amount.Amount) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	zero := amount.NewCoinAmount(0, 0)
	if am.Less(zero) {
		return ErrMinusInput
	}
	//log.Println("AddCollectedFee", ctw.TargetHeight(), am.String(), p.CollectedFee(ctw).Add(am).String())
	ctw.SetProcessData(tagCollectedFee, p.CollectedFee(ctw).Add(am).Bytes())
	return nil
}

// SubCollectedFee subtracts collected fee
func (p *Vault) SubCollectedFee(ctw *types.ContextWrapper, am *amount.Amount) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	total := p.CollectedFee(ctw)
	if total.Less(am) {
		return ErrMinusCollectedFee
	}
	//log.Println("SubCollectedFee", ctw.TargetHeight(), am.String(), p.CollectedFee(ctw).Sub(am).String())

	total = total.Sub(am)
	if total.IsZero() {
		ctw.SetProcessData(tagCollectedFee, nil)
	} else {
		ctw.SetProcessData(tagCollectedFee, total.Bytes())
	}
	return nil
}
