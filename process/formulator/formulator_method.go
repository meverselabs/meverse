package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/types"
)

func (p *Formulator) addGenCount(ctw *types.ContextWrapper, addr common.Address) {
	ctw.SetProcessData(toGenCountKey(addr), util.Uint64ToBytes(p.getGenCount(ctw, addr)+1))
}

func (p *Formulator) getGenCount(ctw *types.ContextWrapper, addr common.Address) uint64 {
	if bs := ctw.ProcessData(toGenCountKey(addr)); len(bs) > 0 {
		return util.BytesToUint64(bs)
	} else {
		return 0
	}
}

func (p *Formulator) removeGenCount(ctw *types.ContextWrapper, addr common.Address) {
	ctw.SetProcessData(toGenCountKey(addr), nil)
}

func (p *Formulator) getGenCountMap(ctw *types.ContextWrapper) (map[common.Address]uint64, error) {
	keys, err := ctw.ProcessDataKeys(tagGenCount)
	if err != nil {
		return nil, err
	}
	CountMap := map[common.Address]uint64{}
	for _, k := range keys {
		if addr, is := fromGenCountKey(k); is {
			CountMap[addr] = p.getGenCount(ctw, addr)
		}
	}
	return CountMap, nil
}

func (p *Formulator) getStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	if bs := ctw.AccountData(HyperAddress, toStakingAmountKey(StakingAddress)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

func (p *Formulator) addStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) {
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), p.getStakingAmount(ctw, HyperAddress, StakingAddress).Add(StakingAmount).Bytes())
}

func (p *Formulator) setStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) {
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), StakingAmount.Bytes())
}

func (p *Formulator) removeStakingAmount(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address) {
	ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), nil)
}

func (p *Formulator) getStakingAmountMap(ctw *types.ContextWrapper, HyperAddress common.Address) (map[common.Address]*amount.Amount, error) {
	keys, err := ctw.AccountDataKeys(HyperAddress, tagStakingAmount)
	if err != nil {
		return nil, err
	}
	PowerMap := map[common.Address]*amount.Amount{}
	for _, k := range keys {
		if addr, is := fromStakingAmountKey(k); is {
			PowerMap[addr] = p.getStakingAmount(ctw, HyperAddress, addr)
		}
	}
	return PowerMap, nil
}

func (p *Formulator) getRewardPower(ctw *types.ContextWrapper, addr common.Address) *amount.Amount {
	if bs := ctw.ProcessData(toRewardPowerKey(addr)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewCoinAmount(0, 0)
	}
}

func (p *Formulator) addRewardPower(ctw *types.ContextWrapper, addr common.Address, Power *amount.Amount) {
	ctw.SetProcessData(toRewardPowerKey(addr), p.getRewardPower(ctw, addr).Add(Power).Bytes())
}

func (p *Formulator) removeRewardPower(ctw *types.ContextWrapper, addr common.Address) {
	ctw.SetProcessData(toRewardPowerKey(addr), nil)
}

func (p *Formulator) getRewardPowerMap(ctw *types.ContextWrapper) (map[common.Address]*amount.Amount, error) {
	keys, err := ctw.ProcessDataKeys(tagRewardPower)
	if err != nil {
		return nil, err
	}
	PowerMap := map[common.Address]*amount.Amount{}
	for _, k := range keys {
		if addr, is := fromRewardPowerKey(k); is {
			PowerMap[addr] = p.getRewardPower(ctw, addr)
		}
	}
	return PowerMap, nil
}

func (p *Formulator) getLastPaidHeight(ctw *types.ContextWrapper) uint32 {
	if bs := ctw.ProcessData(tagLastPaidHeight); len(bs) > 0 {
		return util.BytesToUint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastPaidHeight(ctw *types.ContextWrapper, lastPaidHeight uint32) {
	ctw.SetProcessData(tagLastPaidHeight, util.Uint32ToBytes(lastPaidHeight))
}

func (p *Formulator) getLastStakingPaidHeight(ctw *types.ContextWrapper, Address common.Address) uint32 {
	if bs := ctw.AccountData(Address, tagLastStakingPaidHeight); len(bs) > 0 {
		return util.BytesToUint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastStakingPaidHeight(ctw *types.ContextWrapper, Address common.Address, lastPaidHeight uint32) {
	ctw.SetAccountData(Address, tagLastStakingPaidHeight, util.Uint32ToBytes(lastPaidHeight))
}

func (p *Formulator) getUserAutoStaking(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address) bool {
	if bs := ctw.AccountData(HyperAddress, toAutoStakingKey(StakingAddress)); len(bs) > 0 {
		return bs[0] == 1
	} else {
		return true
	}
}

func (p *Formulator) setUserAutoStaking(ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, IsAutoStaking bool) {
	if IsAutoStaking {
		ctw.SetAccountData(HyperAddress, toAutoStakingKey(StakingAddress), []byte{1})
	} else {
		ctw.SetAccountData(HyperAddress, toAutoStakingKey(StakingAddress), nil)
	}
}
