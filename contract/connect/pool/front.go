package pool

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *PoolContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *PoolContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////
// func (f *front) SetOwner(cc *types.ContractContext, To common.Address) error {
// 	return f.cont.setOwner(cc, To)
// }
func (f *front) SetFarm(cc *types.ContractContext, To common.Address) error {
	return f.cont.setFarm(cc, To)
}
func (f *front) SetGov(cc *types.ContractContext, To common.Address) error {
	return f.cont.setGov(cc, To)
}
func (f *front) SetWant(cc *types.ContractContext, To common.Address) error {
	return f.cont.setWant(cc, To)
}

func (f *front) SetFeeFundAddress(cc *types.ContractContext, val common.Address) error {
	return f.cont.setFeeFundAddress(cc, val)
}

func (f *front) SetRewardsAddress(cc *types.ContractContext, val common.Address) error {
	return f.cont.setRewardsAddress(cc, val)
}

func (f *front) SetDepositFeeFactor(cc *types.ContractContext, val uint16) error {
	return f.cont.setDepositFeeFactor(cc, val)
}

func (f *front) SetWithdrawFeeFactor(cc *types.ContractContext, val uint16) error {
	return f.cont.setWithdrawFeeFactor(cc, val)
}

func (f *front) SetRewardFeeFactor(cc *types.ContractContext, val uint16) error {
	return f.cont.setRewardFeeFactor(cc, val)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
func (f *front) Withdraw(cc *types.ContractContext, _userAddress common.Address, _wantAmt *amount.Amount) (*amount.Amount, error) {
	return f.cont.Withdraw(cc, _userAddress, _wantAmt)
}
func (f *front) Deposit(cc *types.ContractContext, _userAddress common.Address, _wantAmt *amount.Amount) (*amount.Amount, error) {
	return f.cont.Deposit(cc, _userAddress, _wantAmt)
}

// View function to see staked Want tokens on frontend.
func (f *front) SetHoldShares(cc *types.ContractContext, height uint32) error {
	return f.cont.SetHoldShares(cc, height)
}

// View function to see staked Want tokens on frontend.
func (f *front) UnsetHoldShares(cc *types.ContractContext) error {
	return f.cont.UnsetHoldShares(cc)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////
// func (f *front) Owner(cc *types.ContractContext) common.Address {
// 	return f.cont.Owner(cc)
// }

func (f *front) Gov(cc *types.ContractContext) common.Address {
	return f.cont.Gov(cc)
}

func (f *front) Farm(cc *types.ContractContext) common.Address {
	return f.cont.Farm(cc)
}

func (f *front) Want(cc *types.ContractContext) common.Address {
	return f.cont.Want(cc)
}

func (f *front) FeeFundAddress(cc *types.ContractContext) common.Address {
	return f.cont.FeeFundAddress(cc)
}

func (f *front) RewardsAddress(cc *types.ContractContext) common.Address {
	return f.cont.RewardsAddress(cc)
}

func (f *front) DepositFeeFactor(cc *types.ContractContext) uint16 {
	return f.cont.DepositFeeFactor(cc)
}

func (f *front) WithdrawFeeFactor(cc *types.ContractContext) uint16 {
	return f.cont.WithdrawFeeFactor(cc)
}

func (f *front) RewardFeeFactor(cc *types.ContractContext) uint16 {
	return f.cont.RewardFeeFactor(cc)
}

func (f *front) WantLockedTotal(cc *types.ContractContext) *amount.Amount {
	return f.cont.wantLockedTotal(cc)
}
func (f *front) SharesTotal(cc *types.ContractContext) *amount.Amount {
	return f.cont.sharesTotal(cc)
}
