package pool

import (
	"errors"

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
func (cont *PoolContract) setFarm(cc *types.ContractContext, To common.Address) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagFarm}, To[:])
	return nil
}
func (cont *PoolContract) setGov(cc *types.ContractContext, To common.Address) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagGov}, To[:])
	return nil
}
func (cont *PoolContract) setWant(cc *types.ContractContext, To common.Address) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}

	cc.SetContractData([]byte{tagWant}, To[:])
	return nil
}

func (cont *PoolContract) setFeeFundAddress(cc *types.ContractContext, val common.Address) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagFeeFundAddress}, val[:])
	return nil
}

func (cont *PoolContract) setRewardsAddress(cc *types.ContractContext, val common.Address) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagRewardsAddress}, val[:])
	return nil
}

func (cont *PoolContract) setDepositFeeFactor(cc *types.ContractContext, val uint16) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagDepositFeeFactor}, bin.Uint16Bytes(val))
	return nil
}

func (cont *PoolContract) setWithdrawFeeFactor(cc *types.ContractContext, val uint16) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagWithdrawFeeFactor}, bin.Uint16Bytes(val))
	return nil
}

func (cont *PoolContract) setRewardFeeFactor(cc *types.ContractContext, val uint16) error {
	if cc.From() != cont.master && !cont.isGov(cc) {
		return errors.New("ownable: caller is not the master")
	}
	cc.SetContractData([]byte{tagRewardFeeFactor}, bin.Uint16Bytes(val))
	return nil
}

func (cont *PoolContract) SetHoldShares(cc *types.ContractContext, height uint32) error {
	if !cont.isFarm(cc) {
		return errors.New("ownable: withdraw are only possible using owner contract")
	}
	cc.SetContractData([]byte{tagHoldShares}, []byte{1})
	cc.SetContractData([]byte{tagHoldSharesHeight}, bin.Uint32Bytes(height))
	return nil
}

func (cont *PoolContract) UnsetHoldShares(cc *types.ContractContext) error {
	if !cont.isFarm(cc) {
		return errors.New("ownable: withdraw are only possible using owner contract")
	}
	cc.SetContractData([]byte{tagHoldShares}, nil)
	return nil
}

func (cont *PoolContract) isHoldShares(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagHoldShares})
	if len(bs) == 0 {
		return false
	}
	bsHeight := cc.ContractData([]byte{tagHoldSharesHeight})
	height := bin.Uint32(bsHeight)
	return cc.TargetHeight() > height
}
