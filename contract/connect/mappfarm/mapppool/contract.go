package mapppool

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

type PoolContract struct {
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////
// func (cont *PoolContract) Owner(cc *types.ContractContext) common.Address {
// 	bs := cc.ContractData([]byte{tagOwner})
// 	var owner common.Address
// 	copy(owner[:], bs)
// 	return owner
// }

func (cont PoolContract) Farm(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagFarm})
	var Farm common.Address
	copy(Farm[:], bs)
	return Farm
}

func (cont PoolContract) Want(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagWant})
	var want common.Address
	copy(want[:], bs)
	return want
}

func (cont PoolContract) WantLockedTotal(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagWantLockedTotal})
	return amount.NewAmountFromBytes(bs)
}
func (cont PoolContract) SharesTotal(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagSharesTotal})
	return amount.NewAmountFromBytes(bs)
}
