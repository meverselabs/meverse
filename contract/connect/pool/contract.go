package pool

import (
	"bytes"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

type PoolContract struct {
	addr   common.Address
	master common.Address
}

func (cont *PoolContract) Name() string {
	return "PoolContract"
}

func (cont *PoolContract) Address() common.Address {
	return cont.addr
}

func (cont *PoolContract) Master() common.Address {
	return cont.master
}

func (cont *PoolContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *PoolContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &PoolContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	// cc.SetContractData([]byte{tagOwner}, data.Owner[:])
	cc.SetContractData([]byte{tagFarm}, data.Farm[:])
	cc.SetContractData([]byte{tagGov}, data.Gov[:])
	cc.SetContractData([]byte{tagWant}, data.Want[:])

	cc.SetContractData([]byte{tagFeeFundAddress}, data.FeeFundAddress[:])
	cc.SetContractData([]byte{tagRewardsAddress}, data.RewardsAddress[:])

	cc.SetContractData([]byte{tagDepositFeeFactor}, bin.Uint16Bytes(data.DepositFeeFactor))

	cc.SetContractData([]byte{tagWithdrawFeeFactor}, bin.Uint16Bytes(data.WithdrawFeeFactor))
	cc.SetContractData([]byte{tagRewardFeeFactor}, bin.Uint16Bytes(data.RewardFeeFactor))

	return nil
}

func (cont *PoolContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
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

func (cont *PoolContract) Gov(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagGov})
	var gov common.Address
	copy(gov[:], bs)
	return gov
}

func (cont *PoolContract) Farm(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagFarm})
	var Farm common.Address
	copy(Farm[:], bs)
	return Farm
}

func (cont *PoolContract) Want(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagWant})
	var want common.Address
	copy(want[:], bs)
	return want
}

func (cont *PoolContract) FeeFundAddress(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagFeeFundAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *PoolContract) RewardsAddress(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagRewardsAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *PoolContract) DepositFeeFactor(cc *types.ContractContext) uint16 {
	bs := cc.ContractData([]byte{tagDepositFeeFactor})
	if len(bs) == 2 {
		return bin.Uint16(bs)
	}
	return 0
}

func (cont *PoolContract) WithdrawFeeFactor(cc *types.ContractContext) uint16 {
	bs := cc.ContractData([]byte{tagWithdrawFeeFactor})
	if len(bs) == 2 {
		return bin.Uint16(bs)
	}
	return 0
}

func (cont *PoolContract) RewardFeeFactor(cc *types.ContractContext) uint16 {
	bs := cc.ContractData([]byte{tagRewardFeeFactor})
	if len(bs) == 2 {
		return bin.Uint16(bs)
	}
	return 0
}

func (cont *PoolContract) wantLockedTotal(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagWantLockedTotal})
	return amount.NewAmountFromBytes(bs)
}
func (cont *PoolContract) sharesTotal(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagSharesTotal})
	return amount.NewAmountFromBytes(bs)
}
