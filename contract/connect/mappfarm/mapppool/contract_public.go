package mapppool

import (
	"errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

const (
	factorMax uint16 = 10000
)

// Receives new deposits from user
func (cont PoolContract) Deposit(cc *types.ContractContext, _userAddress common.Address, _wantAmt *amount.Amount) (*amount.Amount, error) {
	if !_wantAmt.IsPlus() {
		return nil, errors.New("wantAmt must plus")
	}
	// if !cont.isFarm(cc) {
	// 	return nil, errors.New("ownable: Deposit are only possible using owner contract")
	// }
	want := cont.Want(cc)

	// IERC20(wantAddress).safeTransferFrom(address(msg.sender),address(this),_wantAmt);
	if _, err := cc.Exec(cc, want, "TransferFrom", []interface{}{cc.From(), cont.Farm(cc), _wantAmt}); err != nil {
		return nil, err
	}

	wantLockedTotal := cont.WantLockedTotal(cc).Add(_wantAmt)
	cont.setWantLockedTotal(cc, wantLockedTotal)
	return _wantAmt, nil
}

func (cont *PoolContract) Withdraw(cc *types.ContractContext, _userAddress common.Address, _wantAmt *amount.Amount) (*amount.Amount, error) {
	// if !cont.isFarm(cc) {
	// 	return nil, errors.New("ownable: Withdraw are only possible using owner contract.")
	// }
	if !_wantAmt.IsPlus() {
		return nil, errors.New("_wantAmt <= 0")
	}

	wantLockedTotal := cont.WantLockedTotal(cc)
	want := cont.Want(cc)

	// wantAmt := IERC20(wantAddress).balanceOf(address(this))
	if ins, err := cc.Exec(cc, want, "BalanceOf", []interface{}{cont.Farm(cc)}); err != nil {
		return nil, err
	} else if len(ins) == 0 {
		return nil, errors.New("invalid Want BalanceOf")
	} else if wantAmt, ok := ins[0].(*amount.Amount); !ok {
		return nil, errors.New("invalid Want BalanceOf !amount")
	} else if _wantAmt.Cmp(wantAmt.Int) > 0 {
		_wantAmt = wantAmt
	}

	if wantLockedTotal.Cmp(_wantAmt.Int) < 0 {
		_wantAmt = wantLockedTotal
	}

	wantLockedTotal = wantLockedTotal.Sub(_wantAmt)
	cont.setWantLockedTotal(cc, wantLockedTotal)

	// IERC20(wantAddress).safeTransfer(cherryFarmAddress, _wantAmt)
	// farm := cont.Farm(cc)
	// if _, err := cc.Exec(cc, want, "Transfer", []interface{}{farm, _wantAmt}); err != nil {
	// 	return nil, err
	// }
	return _wantAmt, nil
}
