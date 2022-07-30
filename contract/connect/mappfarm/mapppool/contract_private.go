package mapppool

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *PoolContract) setWantLockedTotal(cc *types.ContractContext, val *amount.Amount) {
	cc.SetContractData([]byte{tagWantLockedTotal}, val.Bytes())
}

func (cont *PoolContract) SetFarm(cc *types.ContractContext, Farm common.Address) error {
	cc.SetContractData([]byte{tagFarm}, Farm[:])
	return nil
}

func (cont *PoolContract) SetWant(cc *types.ContractContext, Want common.Address) error {
	cc.SetContractData([]byte{tagWant}, Want[:])
	return nil
}
