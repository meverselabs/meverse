package mapppool

import (
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *PoolContract) setWantLockedTotal(cc *types.ContractContext, val *amount.Amount) {
	cc.SetContractData([]byte{tagWantLockedTotal}, val.Bytes())
}
