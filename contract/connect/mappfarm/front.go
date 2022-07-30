package mappfarm

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *FarmContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *FarmContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (f *front) MassUpdatePools(cc *types.ContractContext) error {
	return f.cont.MassUpdatePools(cc)
}
func (f *front) UpdatePool(cc *types.ContractContext, _pid uint64) error {
	return f.cont.UpdatePool(cc, _pid)
}
func (f *front) Deposit(cc *types.ContractContext, _pid uint64, _wantAmt *amount.Amount) error {
	return f.cont.Deposit(cc, _pid, _wantAmt)
}
func (f *front) Withdraw(cc *types.ContractContext, _pid uint64, _wantAmt *amount.Amount) error {
	return f.cont.Withdraw(cc, _pid, _wantAmt)
}
func (f *front) WithdrawAll(cc *types.ContractContext, _pid uint64) error {
	return f.cont.WithdrawAll(cc, _pid)
}
func (f *front) EmergencyWithdraw(cc *types.ContractContext, _pid uint64) error {
	return f.cont.EmergencyWithdraw(cc, _pid)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
func (f *front) SetOwner(cc *types.ContractContext, To common.Address) error {
	return f.cont.setOwner(cc, To)
}
func (f *front) SetOwnerReward(cc *types.ContractContext, OwnerReward uint16) error {
	return f.cont.setOwnerReward(cc, OwnerReward)
}
func (f *front) SetTokenPerBlock(cc *types.ContractContext, TokenPerBlock *amount.Amount) error {
	return f.cont.setTokenPerBlock(cc, TokenPerBlock)
}
func (f *front) SetStartBlock(cc *types.ContractContext, StartBlock uint32) error {
	return f.cont.setStartBlock(cc, StartBlock)
}

// func (f *front) Add(cc *types.ContractContext, _allocPoint uint32, _want common.Address, _withUpdate bool, _strat common.Address) error {
// 	return f.cont.Add(cc, _allocPoint, _want, _withUpdate, _strat)
// }
// func (f *front) Set(cc *types.ContractContext, _pid uint64, _allocPoint uint32, _withUpdate bool) error {
// 	return f.cont.Set(cc, _pid, _allocPoint, _withUpdate)
// }
func (f *front) InCaseTokensGetStuck(cc *types.ContractContext, _token common.Address, _amount *amount.Amount) error {
	return f.cont.InCaseTokensGetStuck(cc, _token, _amount)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////
func (f *front) Owner(cc *types.ContractContext) common.Address {
	return f.cont.Owner(cc)
}

func (f *front) OwnerReward(cc *types.ContractContext) uint16 {
	return f.cont.OwnerReward(cc)
}

func (f *front) FarmToken(cc *types.ContractContext) common.Address {
	return f.cont.FarmToken(cc)
}

func (f *front) TokenPerBlock(cc *types.ContractContext) *amount.Amount {
	return f.cont.TokenPerBlock(cc)
}

func (f *front) StartBlock(cc *types.ContractContext) uint32 {
	return f.cont.StartBlock(cc)
}

func (f *front) PoolLength(cc *types.ContractContext) uint64 {
	return 1
}

func (f *front) PoolInfo(cc *types.ContractContext, pid uint64) (common.Address, uint32, uint32, *amount.Amount, common.Address, error) {
	return f.cont.PoolInfo(cc, pid)
}

func (f *front) UserInfo(cc *types.ContractContext, pid uint64, user common.Address) (*amount.Amount, *amount.Amount, error) {
	return f.cont.UserInfo(cc, pid, user)
}

func (f *front) TotalAllocPoint(cc *types.ContractContext) uint32 {
	return 1
}

// Return reward multiplier over the given _from to _to block.
func (f *front) GetMultiplier(cc *types.ContractContext, _from uint32, _to uint32) (uint32, error) {
	return f.cont.GetMultiplier(cc, _from, _to)
}

// View function to see pending Cherry on frontend.
func (f *front) PendingReward(cc *types.ContractContext, _pid uint64, _user common.Address) (*amount.Amount, error) {
	return f.cont.PendingReward(cc, _pid, _user)
}

// View function to see staked Want tokens on frontend.
func (f *front) StakedWantTokens(cc *types.ContractContext, _pid uint64, _user common.Address) (*amount.Amount, error) {
	return f.cont.StakedWantTokens(cc, _pid, _user)
}

func (f *front) Want(cc *types.ContractContext) common.Address {
	return f.cont.pool.Want(cc)
}

func (f *front) WantLockedTotal(cc *types.ContractContext) *amount.Amount {
	return f.cont.pool.WantLockedTotal(cc)
}
func (f *front) SharesTotal(cc *types.ContractContext) *amount.Amount {
	return f.cont.pool.SharesTotal(cc)
}
