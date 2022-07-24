package farm

import (
	"errors"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

// Update reward variables for all pools. Be careful of gas spending!
func (cont *FarmContract) MassUpdatePools(cc *types.ContractContext) error {
	length := uint64(cont.PoolLength(cc))
	var pid uint64 = 0
	for ; pid < length; pid++ {
		err := cont.UpdatePool(cc, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update reward variables of the given pool to be up-to-date.
func (cont *FarmContract) UpdatePool(cc *types.ContractContext, _pid uint64) error {
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return err
	}
	if cc.TargetHeight() <= pool.LastRewardBlock {
		return nil
	}

	var total *amount.Amount
	if !cont.isHoldShares2(cc) {
		sharesTotal, err := cont._sharesTotal(cc, pool)
		if err != nil {
			return err
		}
		total = sharesTotal
	} else {
		wantLockedTotal, err := cont.callContAmountValue(cc, pool.Strat, "WantLockedTotal")
		if err != nil {
			return err
		}
		total = wantLockedTotal
	}
	if total.IsZero() {
		pool.LastRewardBlock = cc.TargetHeight()
		return nil
	}

	multiplier, err := cont.GetMultiplier(cc, pool.LastRewardBlock, cc.TargetHeight())
	if err != nil {
		return err
	}
	tokenPerBlock := cont.TokenPerBlock(cc)
	totalAllocPoint := cont.TotalAllocPoint(cc)
	var CherryReward *amount.Amount
	if totalAllocPoint == 0 {
		CherryReward = amount.NewAmount(0, 0)
	} else {
		CherryReward = tokenPerBlock.MulC(int64(multiplier)).MulC(int64(pool.AllocPoint)).DivC(int64(totalAllocPoint))
		ownerFactor := cont.OwnerReward(cc)
		ownerReward := CherryReward.MulC(int64(ownerFactor)).DivC(1000)

		owner := cont.Owner(cc)

		FarmToken := cont.FarmToken(cc)
		if _, err := cc.Exec(cc, FarmToken, "Mint", []interface{}{owner, ownerReward}); err != nil {
			return err
		}
		// CherryToken(CherryAddr).mint(address(this), CherryReward);
		if _, err := cc.Exec(cc, FarmToken, "Mint", []interface{}{cont.addr, CherryReward}); err != nil {
			return err
		}
	}
	pool.AccTokenPerShare = pool.AccTokenPerShare.Add(CherryReward.Div(total))
	pool.LastRewardBlock = cc.TargetHeight()
	err = cont.setPoolInfo(cc, _pid, pool)
	if err != nil {
		return err
	}
	return nil
}

// Want tokens moved from user -> FletaFinance (Cherry allocation) -> Strat (compounding)
func (cont *FarmContract) Deposit(cc *types.ContractContext, _pid uint64, _wantAmt *amount.Amount) error {
	if err := cont.UpdatePool(cc, _pid); err != nil {
		return err
	}
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return err
	}
	user, err := cont._userInfo(cc, _pid, cc.From())
	if err != nil {
		return err
	}

	if !user.Shares.IsZero() {
		var pending *amount.Amount
		if cont.isHoldShares3(cc) && _pid == 0 {
			pending = user.Shares.Mul(pool.AccTokenPerShare).Sub(user.RewardDebt).DivC(2)
		} else {
			pending = user.Shares.Mul(pool.AccTokenPerShare).Sub(user.RewardDebt)
		}
		if pending.IsPlus() {
			if err := cont.safeFarmTokenTransfer(cc, cc.From(), pending); err != nil {
				return err
			}
		}
	}
	if _wantAmt.IsPlus() {
		if _, err := cc.Exec(cc, pool.Want, "TransferFrom", []interface{}{cc.From(), cont.addr, _wantAmt}); err != nil {
			return err
		}
		// pool.want.safeIncreaseAllowance(pool.strat, _wantAmt)
		if err := cont.safeIncreaseAllowance(cc, pool.Want, pool.Strat, _wantAmt); err != nil {
			return err
		}

		// sharesAdded = IStrategy(poolInfo[_pid].strat).deposit(msg.sender, _wantAmt)
		if ins, err := cc.Exec(cc, pool.Strat, "Deposit", []interface{}{cc.From(), _wantAmt}); err != nil {
			return err
		} else if len(ins) == 0 {
			return errors.New("invalid Strat Depoist")
		} else if amt, ok := ins[0].(*amount.Amount); !ok {
			return errors.New("invalid Strat Depoist amount")
		} else {
			if cont.isHoldShares2(cc) && _pid == 0 {
				user.Shares = user.Shares.Add(amt.MulC(2))
			} else {
				user.Shares = user.Shares.Add(amt)
			}
		}
	}
	user.RewardDebt = user.Shares.Mul(pool.AccTokenPerShare)
	cont.setUserInfo(cc, _pid, cc.From(), user)
	// emit Deposit(msg.sender, _pid, _wantAmt);
	return nil
}

// Withdraw LP tokens from MasterPicker.
func (cont *FarmContract) Withdraw(cc *types.ContractContext, _pid uint64, _wantAmt *amount.Amount) error {
	if err := cont.UpdatePool(cc, _pid); err != nil {
		return err
	}
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return err
	}
	user, err := cont._userInfo(cc, _pid, cc.From())
	if err != nil {
		return err
	}

	wantLockedTotal, err := cont.callContAmountValue(cc, pool.Strat, "WantLockedTotal")
	if err != nil {
		return err
	}
	var sharesTotal *amount.Amount
	if !cont.isHoldShares(cc) {
		sharesTotal, err = cont.callContAmountValue(cc, pool.Strat, "SharesTotal")
		if err != nil {
			return err
		}
	}
	// require(user.shares > 0, "user.shares is 0");
	if !user.Shares.IsPlus() {
		return errors.New("user.shares is 0")
	}
	// require(sharesTotal > 0, "sharesTotal is 0");
	if !cont.isHoldShares(cc) {
		if !sharesTotal.IsPlus() {
			return errors.New("sharesTotal is 0")
		}
	}

	// Withdraw pending Cherry
	// uint256 pending =
	// 	user.shares.mul(pool.accCherryPerShare).div(1e12).sub(
	// 		user.rewardDebt
	// 	);
	// if (pending > 0) {
	// 	safeCherryTransfer(msg.sender, pending);
	// }
	var pending *amount.Amount
	if cont.isHoldShares3(cc) && _pid == 0 {
		pending = user.Shares.Mul(pool.AccTokenPerShare).Sub(user.RewardDebt).DivC(2)
	} else {
		pending = user.Shares.Mul(pool.AccTokenPerShare).Sub(user.RewardDebt)
	}

	if pending.IsPlus() {
		if err := cont.safeFarmTokenTransfer(cc, cc.From(), pending); err != nil {
			return err
		}
	}

	// Withdraw want tokens
	var amt *amount.Amount
	if !cont.isHoldShares(cc) {
		amt = user.Shares.Mul(wantLockedTotal).Div(sharesTotal)
	} else {
		amt = user.Shares
	}
	// if (_wantAmt > amount) {
	// 	_wantAmt = amount;
	// }
	if _wantAmt.Cmp(amt.Int) > 0 {
		_wantAmt = amt
	}
	if _wantAmt.IsPlus() {
		// uint256 sharesRemoved = IStrategy(poolInfo[_pid].strat).withdraw(msg.sender, _wantAmt);
		sharesRemoved, err := cont.callContAmountValue(cc, pool.Strat, "Withdraw", cc.From(), _wantAmt)
		if err != nil {
			return err
		}

		if sharesRemoved.Cmp(user.Shares.Int) > 0 {
			user.Shares = amount.ZeroCoin
		} else {
			if cont.isHoldShares2(cc) && _pid == 0 {
				user.Shares = user.Shares.Sub(sharesRemoved.MulC(2))
			} else {
				user.Shares = user.Shares.Sub(sharesRemoved)
			}
		}

		// uint256 wantBal = IERC20(pool.want).balanceOf(address(this));
		wantBal, err := cont.callContAmountValue(cc, pool.Want, "BalanceOf", cont.addr)
		if err != nil {
			return err
		}
		// if (wantBal < _wantAmt) {
		// 	_wantAmt = wantBal;
		// }
		if wantBal.Cmp(_wantAmt.Int) < 0 {
			_wantAmt = wantBal
		}
		// pool.want.safeTransfer(address(msg.sender), _wantAmt)
		_, err = cc.Exec(cc, pool.Want, "Transfer", []interface{}{cc.From(), _wantAmt})
		if err != nil {
			return err
		}
	}

	user.RewardDebt = user.Shares.Mul(pool.AccTokenPerShare)
	cont.setUserInfo(cc, _pid, cc.From(), user)
	// emit Withdraw(msg.sender, _pid, _wantAmt);
	return nil
}

func (cont *FarmContract) WithdrawAll(cc *types.ContractContext, _pid uint64) error {
	if amt, err := cont.StakedWantTokens(cc, _pid, cc.From()); err != nil {
		return err
	} else {
		return cont.Withdraw(cc, _pid, amt)
	}
}

// Withdraw without caring about rewards. EMERGENCY ONLY.
func (cont *FarmContract) EmergencyWithdraw(cc *types.ContractContext, _pid uint64) error {
	if err := cont.UpdatePool(cc, _pid); err != nil {
		return err
	}
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return err
	}
	user, err := cont._userInfo(cc, _pid, cc.From())
	if err != nil {
		return err
	}

	wantLockedTotal, err := cont.callContAmountValue(cc, pool.Strat, "WantLockedTotal")
	if err != nil {
		return err
	}

	var sharesTotal *amount.Amount
	if !cont.isHoldShares(cc) {
		sharesTotal, err = cont.callContAmountValue(cc, pool.Strat, "SharesTotal")
		if err != nil {
			return err
		}
	}

	var amt *amount.Amount
	if !cont.isHoldShares(cc) {
		amt = user.Shares.Mul(wantLockedTotal).Div(sharesTotal)
	} else {
		amt = user.Shares
	}

	// IStrategy(poolInfo[_pid].strat).withdraw(msg.sender, amount);
	_, err = cont.callContAmountValue(cc, pool.Strat, "Withdraw", cc.From(), amt)
	if err != nil {
		return err
	}
	_, err = cc.Exec(cc, pool.Want, "Transfer", []interface{}{cc.From(), amt})
	if err != nil {
		return err
	}
	user.Shares = amount.ZeroCoin
	user.RewardDebt = amount.ZeroCoin
	cont.setUserInfo(cc, _pid, cc.From(), user)
	return nil
}
