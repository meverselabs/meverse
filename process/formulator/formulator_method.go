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
	if ns := ctw.ProcessData(toGenCountNumberKey(addr)); len(ns) == 0 {
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

func (p *Formulator) flushGenCountMap(ctw *types.ContextWrapper) (map[common.Address]uint64, error) {
	CountMap := map[common.Address]uint64{}
	if bs := ctw.ProcessData(tagGenCountCount); len(bs) > 0 {
		Count := util.BytesToUint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], ctw.ProcessData(toGenCountReverseKey(i)))
			CountMap[addr] = p.getGenCount(ctw, addr)

			ctw.SetProcessData(toGenCountKey(addr), nil)
			ctw.SetProcessData(toGenCountNumberKey(addr), nil)
			ctw.SetProcessData(toGenCountReverseKey(i), nil)
		}
		ctw.SetProcessData(tagGenCountCount, nil)
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

	if ns := ctw.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)); len(ns) == 0 {
		var Count uint32
		if bs := ctw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
			Count = util.BytesToUint32(bs)
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), util.Uint32ToBytes(Count))
		ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count), StakingAddress[:])
		Count++
		ctw.SetAccountData(HyperAddress, tagStakingAmountCount, util.Uint32ToBytes(Count))
	}
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), p.GetStakingAmount(ctw, HyperAddress, StakingAddress).Add(StakingAmount).Bytes())
}

func (p *Formulator) subStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, am *amount.Amount) error {
	total := p.GetStakingAmount(ctw, HyperAddress, StakingAddress)
	if total.Less(am) {
		return ErrMinusBalance
	}
	//log.Println("SubBalance", ctw.TargetHeight(), addr.String(), am.String(), p.Balance(ctw, addr).Sub(am).String())

	total = total.Sub(am)
	if total.IsZero() {
		if ns := ctw.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)); len(ns) > 0 {
			var Count uint32
			if bs := ctw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
				Count = util.BytesToUint32(bs)
			}
			Number := util.BytesToUint32(ns)
			if Number != Count-1 {
				var swapAddr common.Address
				copy(swapAddr[:], ctw.AccountData(HyperAddress, toStakingAmountReverseKey(Count-1)))
				ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Number), swapAddr[:])
				ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(swapAddr), util.Uint32ToBytes(Number))
			}
			ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), nil)
			ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count-1), nil)
			Count--
			ctw.SetAccountData(HyperAddress, tagStakingAmountCount, util.Uint32ToBytes(Count))
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), nil)
	} else {
		ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), total.Bytes())
	}
	return nil
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
