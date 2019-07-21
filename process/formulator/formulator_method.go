package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/types"
)

func (p *Formulator) getGenCount(lw types.LoaderWrapper, addr common.Address) uint64 {
	if bs := lw.ProcessData(toGenCountKey(addr)); len(bs) > 0 {
		return util.BytesToUint64(bs)
	} else {
		return 0
	}
}

func (p *Formulator) addGenCount(ctw *types.ContextWrapper, addr common.Address) {
	genCount := p.getGenCount(ctw, addr)
	if genCount == 0 {
		var Count uint32
		if bs := ctw.ProcessData(tagGenCountCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		ctw.SetProcessData(toGenCountNumberKey(addr), util.Uint32ToBytes(Count))
		ctw.SetProcessData(toGenCountReverseKey(Count), addr[:])
		Count++
		ctw.SetProcessData(tagGenCountCount, util.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toGenCountKey(addr), util.Uint64ToBytes(p.getGenCount(ctw, addr)+1))
}

func (p *Formulator) removeGenCount(ctw *types.ContextWrapper, addr common.Address) {
	genCount := p.getGenCount(ctw, addr)
	if genCount != 0 {
		var Count uint32
		if bs := ctw.ProcessData(tagGenCountCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		Number := util.BytesToUint32(ctw.ProcessData(toGenCountNumberKey(addr)))
		if Number != Count-1 {
			var swapAddr common.Address
			copy(swapAddr[:], ctw.ProcessData(toGenCountReverseKey(Count-1)))
			ctw.SetProcessData(toGenCountNumberKey(swapAddr), util.Uint32ToBytes(Number))
			ctw.SetProcessData(toGenCountReverseKey(Number), swapAddr[:])
		} else {
		}
		ctw.SetProcessData(toGenCountNumberKey(addr), nil)
		ctw.SetProcessData(toGenCountReverseKey(Count-1), nil)
		Count--
		ctw.SetProcessData(tagGenCountCount, util.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toGenCountKey(addr), nil)
}

func (p *Formulator) getGenCountMap(lw types.LoaderWrapper) (map[common.Address]uint64, error) {
	CountMap := map[common.Address]uint64{}
	if bs := lw.ProcessData(tagGenCountCount); len(bs) > 0 {
		Count := util.BytesToUint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], lw.ProcessData(toGenCountReverseKey(i)))
			CountMap[addr] = p.getGenCount(lw, addr)
		}
	}
	return CountMap, nil
}

// GetStakingAmount returns the staking amount of the address at the hyper formulator
func (p *Formulator) GetStakingAmount(lw types.LoaderWrapper, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.AccountData(HyperAddress, toStakingAmountKey(StakingAddress)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

// AddStakingAmount adds staking amount of the address at the hyper formulator
func (p *Formulator) AddStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	am := p.GetStakingAmount(ctw, HyperAddress, StakingAddress)
	if am.IsZero() {
		var Count uint32
		if bs := ctw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), util.Uint32ToBytes(Count))
		ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count), StakingAddress[:])
		Count++
		ctw.SetAccountData(HyperAddress, tagStakingAmountCount, util.Uint32ToBytes(Count))
	}
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), am.Add(StakingAmount).Bytes())
}

func (p *Formulator) subStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, am *amount.Amount) error {
	total := p.GetStakingAmount(ctw, HyperAddress, StakingAddress)
	if total.Less(am) {
		return ErrMinusBalance
	}
	//log.Println("SubBalance", ctw.TargetHeight(), addr.String(), am.String(), p.Balance(ctw, addr).Sub(am).String())

	total = total.Sub(am)
	if total.IsZero() {
		p.removeStakingAmount(ctw, HyperAddress, StakingAddress)
	} else {
		ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), total.Bytes())
	}
	return nil
}

func (p *Formulator) removeStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address) {
	am := p.GetStakingAmount(ctw, HyperAddress, StakingAddress)
	if !am.IsZero() {
		var Count uint32
		if bs := ctw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		Number := util.BytesToUint32(ctw.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)))
		if Number != Count-1 {
			var swapAddr common.Address
			copy(swapAddr[:], ctw.AccountData(HyperAddress, toStakingAmountReverseKey(Count-1)))
			ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(swapAddr), util.Uint32ToBytes(Number))
			ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Number), swapAddr[:])
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), nil)
		ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count-1), nil)
		Count--
		ctw.SetAccountData(HyperAddress, tagStakingAmountCount, util.Uint32ToBytes(Count))
	}
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), nil)
}

// GetStakingAmountMap returns all staking amount of the hyper formulator
func (p *Formulator) GetStakingAmountMap(lw types.LoaderWrapper, HyperAddress common.Address) (map[common.Address]*amount.Amount, error) {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	PowerMap := map[common.Address]*amount.Amount{}
	if bs := lw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
		Count := util.BytesToUint32(bs)
		for i := uint32(0); i < Count; i++ {
			var StakingAddress common.Address
			copy(StakingAddress[:], lw.AccountData(HyperAddress, toStakingAmountReverseKey(i)))
			PowerMap[StakingAddress] = p.GetStakingAmount(lw, HyperAddress, StakingAddress)
		}
	}
	return PowerMap, nil
}

func (p *Formulator) getRewardPower(lw types.LoaderWrapper, addr common.Address) *amount.Amount {
	if bs := lw.ProcessData(toRewardPowerKey(addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

func (p *Formulator) addRewardPower(ctw *types.ContextWrapper, addr common.Address, Power *amount.Amount) {
	am := p.getRewardPower(ctw, addr)
	if am.IsZero() {
		var Count uint32
		if bs := ctw.ProcessData(tagRewardPowerCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		ctw.SetProcessData(toRewardPowerNumberKey(addr), util.Uint32ToBytes(Count))
		ctw.SetProcessData(toRewardPowerReverseKey(Count), addr[:])
		Count++
		ctw.SetProcessData(tagRewardPowerCount, util.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toRewardPowerKey(addr), p.getRewardPower(ctw, addr).Add(Power).Bytes())
}

func (p *Formulator) removeRewardPower(ctw *types.ContextWrapper, addr common.Address) {
	am := p.getRewardPower(ctw, addr)
	if !am.IsZero() {
		var Count uint32
		if bs := ctw.ProcessData(tagRewardPowerCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		Number := util.BytesToUint32(ctw.ProcessData(toRewardPowerNumberKey(addr)))
		if Number != Count-1 {
			var addr common.Address
			copy(addr[:], ctw.ProcessData(toRewardPowerReverseKey(Count-1)))
			ctw.SetProcessData(toRewardPowerNumberKey(addr), util.Uint32ToBytes(Number))
			ctw.SetProcessData(toRewardPowerReverseKey(Number), addr[:])
		}
		ctw.SetProcessData(toRewardPowerNumberKey(addr), nil)
		ctw.SetProcessData(toRewardPowerReverseKey(Count-1), nil)
		Count--
		ctw.SetProcessData(tagRewardPowerCount, util.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toRewardPowerKey(addr), nil)
}

func (p *Formulator) getRewardPowerMap(lw types.LoaderWrapper) (map[common.Address]*amount.Amount, error) {
	PowerMap := map[common.Address]*amount.Amount{}
	if bs := lw.ProcessData(tagRewardPowerCount); len(bs) > 0 {
		Count := util.BytesToUint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], lw.ProcessData(toRewardPowerReverseKey(i)))
			PowerMap[addr] = p.getRewardPower(lw, addr)
		}
	}
	return PowerMap, nil
}

func (p *Formulator) getLastPaidHeight(lw types.LoaderWrapper) uint32 {
	if bs := lw.ProcessData(tagLastPaidHeight); len(bs) > 0 {
		return util.BytesToUint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastPaidHeight(ctw *types.ContextWrapper, lastPaidHeight uint32) {
	ctw.SetProcessData(tagLastPaidHeight, util.Uint32ToBytes(lastPaidHeight))
}

func (p *Formulator) getLastStakingPaidHeight(lw types.LoaderWrapper, Address common.Address) uint32 {
	if bs := lw.AccountData(Address, tagLastStakingPaidHeight); len(bs) > 0 {
		return util.BytesToUint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastStakingPaidHeight(ctw *types.ContextWrapper, Address common.Address, lastPaidHeight uint32) {
	ctw.SetAccountData(Address, tagLastStakingPaidHeight, util.Uint32ToBytes(lastPaidHeight))
}

func (p *Formulator) getUserAutoStaking(lw types.LoaderWrapper, HyperAddress common.Address, StakingAddress common.Address) bool {
	if bs := lw.AccountData(HyperAddress, toAutoStakingKey(StakingAddress)); len(bs) > 0 {
		return bs[0] == 1
	} else {
		return false
	}
}

func (p *Formulator) setUserAutoStaking(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, IsAutoStaking bool) {
	if IsAutoStaking {
		ctw.SetAccountData(HyperAddress, toAutoStakingKey(StakingAddress), []byte{1})
	} else {
		ctw.SetAccountData(HyperAddress, toAutoStakingKey(StakingAddress), nil)
	}
}
