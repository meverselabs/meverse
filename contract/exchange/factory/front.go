package factory

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *FactoryContract) Front() interface{} {
	return &FactoryFront{
		cont: cont,
	}
}

type FactoryFront struct {
	cont *FactoryContract
}

//////////////////////////////////////////////////
// Factory Reader Functions
//////////////////////////////////////////////////
func (f *FactoryFront) Owner(cc types.ContractLoader) common.Address {
	return f.cont.owner(cc)
}
func (f *FactoryFront) GetPair(cc types.ContractLoader, token0, token1 common.Address) common.Address {
	return f.cont.getPair(cc, token0, token1)
}
func (f *FactoryFront) AllPairs(cc types.ContractLoader) []common.Address {
	return f.cont.allPairs(cc)
}
func (f *FactoryFront) AllPairsLength(cc types.ContractLoader) uint16 {
	return f.cont.allPairsLength(cc)
}

//////////////////////////////////////////////////
// Factory Writer Functions
//////////////////////////////////////////////////
func (f *FactoryFront) CreatePairUni(cc *types.ContractContext, tokenA, tokenB, payToken common.Address, name, symbol string, owner, winner common.Address, fee, adminFee, winnerFee uint64, whiteList common.Address, groupId hash.Hash256, classID uint64) (common.Address, error) {
	return f.cont.createPairUni(cc, tokenA, tokenB, payToken, name, symbol, owner, winner, fee, adminFee, winnerFee, whiteList, groupId, classID)
}
func (f *FactoryFront) CreatePairStable(cc *types.ContractContext, tokenA, tokenB, payToken common.Address, name, symbol string, owner, winner common.Address, fee, adminFee, winnerFee uint64, whiteList common.Address, groupId hash.Hash256, amp, classID uint64) (common.Address, error) {
	return f.cont.createPairStable(cc, tokenA, tokenB, payToken, name, symbol, owner, winner, fee, adminFee, winnerFee, whiteList, groupId, amp, classID)
}
func (f *FactoryFront) SetOwner(cc *types.ContractContext, _owner common.Address) error {
	return f.cont.setOwner(cc, _owner)
}
