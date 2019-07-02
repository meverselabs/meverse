package pof

import (
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

func init() {
	encoding.Register(rewardData{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(rewardData)
		if err := enc.EncodeUint32(item.lastPaidHeight); err != nil {
			return err
		}
		if err := enc.Encode(item.powerMap); err != nil {
			return err
		}
		if err := enc.Encode(item.stakingPowerMap); err != nil {
			return err
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := newRewardData()
		if v, err := dec.DecodeUint32(); err != nil {
			return err
		} else {
			item.lastPaidHeight = v
		}
		if err := dec.Decode(&item.powerMap); err != nil {
			return err
		}
		if err := dec.Decode(&item.stakingPowerMap); err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type rewardData struct {
	lastPaidHeight  uint32
	powerMap        *types.AddressAmountMap
	stakingPowerMap *types.AddressAddressAmountMap
}

func newRewardData() *rewardData {
	return &rewardData{
		powerMap:        types.NewAddressAmountMap(),
		stakingPowerMap: types.NewAddressAddressAmountMap(),
	}
}

func (rd *rewardData) getRewardPower(addr common.Address) *amount.Amount {
	if PowerSum, has := rd.powerMap.Get(addr); has {
		return PowerSum
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

func (rd *rewardData) addRewardPower(addr common.Address, Power *amount.Amount) {
	rd.powerMap.Put(addr, rd.getRewardPower(addr).Add(Power))
}

func (rd *rewardData) removeRewardPower(addr common.Address) {
	rd.powerMap.Delete(addr)
}

func (rd *rewardData) getStakingPower(addr common.Address, StakingAddress common.Address) *amount.Amount {
	if PowerMap, has := rd.stakingPowerMap.Get(addr); has {
		if PowerSum, has := PowerMap.Get(StakingAddress); has {
			return PowerSum
		} else {
			return amount.NewCoinAmount(0, 0)
		}
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

func (rd *rewardData) addStakingPower(addr common.Address, StakingAddress common.Address, Power *amount.Amount) {
	PowerMap, has := rd.stakingPowerMap.Get(addr)
	if !has {
		PowerMap = types.NewAddressAmountMap()
		rd.stakingPowerMap.Put(addr, PowerMap)
	}
	PowerMap.Put(StakingAddress, rd.getStakingPower(addr, StakingAddress).Add(Power))
}
