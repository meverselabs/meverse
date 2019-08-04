package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// RewardEvent moves a ownership of utxos
type RewardEvent struct {
	Height_        uint32
	Index_         uint16
	N_             uint16
	GenBlockMap    *types.AddressUint32Map
	RewardMap      *types.AddressAmountMap
	StackedMap     *types.AddressAmountMap
	CommissionMap  *types.AddressAmountMap
	StakedMap      *types.AddressAddressAmountMap
	StakeRewardMap *types.AddressAddressAmountMap
}

// Height returns the height of the event
func (ev *RewardEvent) Height() uint32 {
	return ev.Height_
}

// Index returns the index of the event
func (ev *RewardEvent) Index() uint16 {
	return ev.Index_
}

// N returns the n of the event
func (ev *RewardEvent) N() uint16 {
	return ev.N_
}

// SetN updates the n of the event
func (ev *RewardEvent) SetN(n uint16) {
	ev.N_ = n
}

// AddReward adds the reward information to the event
func (ev *RewardEvent) AddReward(addr common.Address, am *amount.Amount) {
	if old, has := ev.RewardMap.Get(addr); has {
		ev.RewardMap.Put(addr, old.Add(am))
	} else {
		ev.RewardMap.Put(addr, am)
	}
}

// AddStacked adds the stacked information to the event
func (ev *RewardEvent) AddStacked(addr common.Address, am *amount.Amount) {
	if old, has := ev.StackedMap.Get(addr); has {
		ev.StackedMap.Put(addr, old.Add(am))
	} else {
		ev.StackedMap.Put(addr, am)
	}
}

// RemoveStacked removes the stacked information to the event
func (ev *RewardEvent) RemoveStacked(addr common.Address) {
	ev.StackedMap.Delete(addr)
}

// AddCommission adds the commission information to the event
func (ev *RewardEvent) AddCommission(addr common.Address, am *amount.Amount) {
	if old, has := ev.CommissionMap.Get(addr); has {
		ev.CommissionMap.Put(addr, old.Add(am))
	} else {
		ev.CommissionMap.Put(addr, am)
	}
}

// AddStaked adds the staked information to the event
func (ev *RewardEvent) AddStaked(HyperAddr common.Address, addr common.Address, am *amount.Amount) {
	mp, has := ev.StakedMap.Get(HyperAddr)
	if !has {
		mp = types.NewAddressAmountMap()
		ev.StakedMap.Put(HyperAddr, mp)
	}

	if old, has := mp.Get(addr); has {
		mp.Put(addr, old.Add(am))
	} else {
		mp.Put(addr, am)
	}
}

// AddStakeReward adds the stake reward information to the event
func (ev *RewardEvent) AddStakeReward(HyperAddr common.Address, addr common.Address, am *amount.Amount) {
	mp, has := ev.StakeRewardMap.Get(HyperAddr)
	if !has {
		mp = types.NewAddressAmountMap()
		ev.StakeRewardMap.Put(HyperAddr, mp)
	}

	if old, has := mp.Get(addr); has {
		mp.Put(addr, old.Add(am))
	} else {
		mp.Put(addr, am)
	}
}

// MarshalJSON is a marshaler function
func (ev *RewardEvent) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"height":`)
	if bs, err := json.Marshal(ev.Height_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"index":`)
	if bs, err := json.Marshal(ev.Index_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"n":`)
	if bs, err := json.Marshal(ev.N_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"reward_map":`)
	if bs, err := ev.RewardMap.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"stacked_map":`)
	if bs, err := ev.StackedMap.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"commission_map":`)
	if bs, err := ev.CommissionMap.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"staked_map":`)
	if bs, err := ev.StakedMap.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"stake_reward_map":`)
	if bs, err := ev.StakeRewardMap.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
