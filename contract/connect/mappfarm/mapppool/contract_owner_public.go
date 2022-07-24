package mapppool

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
// func (cont *PoolContract) setOwner(cc *types.ContractContext, To common.Address) error {
// 	if cc.From() != cont.master && !cont.isGov(cc) {
// 		return errors.New("ownable: caller is not the master")
// 	}
// 	cc.SetContractData([]byte{tagOwner}, To[:])
// 	return nil
// }
// func (cont *PoolContract) setGov(cc *types.ContractContext, To common.Address) error {
// 	cc.SetContractData([]byte{tagGov}, To[:])
// 	return nil
// }
func (cont *PoolContract) SetWant(cc *types.ContractContext, To common.Address) error {
	cc.SetContractData([]byte{tagWant}, To[:])
	return nil
}

func (cont *PoolContract) SetFeeFundAddress(cc *types.ContractContext, val common.Address) error {
	cc.SetContractData([]byte{tagFeeFundAddress}, val[:])
	return nil
}

func (cont *PoolContract) SetRewardsAddress(cc *types.ContractContext, val common.Address) error {
	cc.SetContractData([]byte{tagRewardsAddress}, val[:])
	return nil
}

func (cont *PoolContract) SetDepositFeeFactor(cc *types.ContractContext, val uint16) error {
	cc.SetContractData([]byte{tagDepositFeeFactor}, bin.Uint16Bytes(val))
	return nil
}

func (cont *PoolContract) SetWithdrawFeeFactor(cc *types.ContractContext, val uint16) error {
	cc.SetContractData([]byte{tagWithdrawFeeFactor}, bin.Uint16Bytes(val))
	return nil
}

func (cont *PoolContract) SetRewardFeeFactor(cc *types.ContractContext, val uint16) error {
	cc.SetContractData([]byte{tagRewardFeeFactor}, bin.Uint16Bytes(val))
	return nil
}
