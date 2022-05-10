package pool

import (
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *PoolContract) isGov(cc *types.ContractContext) bool {
	return cont.Gov(cc) == cc.From()
}

// func (cont *PoolContract) isOwner(cc *types.ContractContext) bool {
// 	return cont.Owner(cc) == cc.From()
// }
func (cont *PoolContract) isFarm(cc *types.ContractContext) bool {
	return cont.Farm(cc) == cc.From()
}

func (cont *PoolContract) lastEarnBlock(cc *types.ContractContext) uint32 {
	bs := cc.ContractData([]byte{tagLastEarnBlock})
	if len(bs) == 4 {
		return bin.Uint32(bs)
	}
	return 0
}
func (cont *PoolContract) setLastEarnBlock(cc *types.ContractContext, val uint32) {
	cc.SetContractData([]byte{tagLastEarnBlock}, bin.Uint32Bytes(val))
}
func (cont *PoolContract) setWantLockedTotal(cc *types.ContractContext, val *amount.Amount) {
	cc.SetContractData([]byte{tagWantLockedTotal}, val.Bytes())
}
func (cont *PoolContract) setSharesTotal(cc *types.ContractContext, val *amount.Amount) {
	cc.SetContractData([]byte{tagSharesTotal}, val.Bytes())
}
