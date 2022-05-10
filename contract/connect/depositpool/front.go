package depositpool

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *DepositPoolContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *DepositPoolContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////
func (f *front) Deposit(cc *types.ContractContext) error {
	return f.cont.Deposit(cc)
}

func (f *front) Withdraw(cc *types.ContractContext) error {
	return f.cont.Withdraw(cc)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////

func (f *front) LockDeposit(cc *types.ContractContext) error {
	return f.cont.LockDeposit(cc)
}

func (f *front) UnlockWithdraw(cc *types.ContractContext) error {
	return f.cont.UnlockWithdraw(cc)
}

func (f *front) ReclaimToken(cc *types.ContractContext, token common.Address, amt *amount.Amount) error {
	return f.cont.ReclaimToken(cc, token, amt)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) Holder(cc *types.ContractContext) *big.Int {
	return f.cont.Holder(cc)
}

func (f *front) Holders(cc *types.ContractContext) []common.Address {
	return f.cont.Holders(cc)
}

func (f *front) IsHolder(cc *types.ContractContext, addr common.Address) bool {
	return f.cont.IsHolder(cc, addr)
}
