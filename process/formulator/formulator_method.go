package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

func (p *Formulator) getGenCount(lw types.LoaderWrapper, addr common.Address) uint32 {
	if bs := lw.ProcessData(toGenCountKey(addr)); len(bs) > 0 {
		return binutil.LittleEndian.Uint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) addGenCount(ctw *types.ContextWrapper, addr common.Address) {
	if ns := ctw.ProcessData(toGenCountNumberKey(addr)); len(ns) == 0 {
		var Count uint32
		if bs := ctw.ProcessData(tagGenCountCount); len(bs) > 0 {
			Count = binutil.LittleEndian.Uint32(bs)
		}
		ctw.SetProcessData(toGenCountNumberKey(addr), binutil.LittleEndian.Uint32ToBytes(Count))
		ctw.SetProcessData(toGenCountReverseKey(Count), addr[:])
		Count++
		ctw.SetProcessData(tagGenCountCount, binutil.LittleEndian.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toGenCountKey(addr), binutil.LittleEndian.Uint32ToBytes(p.getGenCount(ctw, addr)+1))
}

func (p *Formulator) flushGenCountMap(ctw *types.ContextWrapper) (map[common.Address]uint32, error) {
	CountMap := map[common.Address]uint32{}
	if bs := ctw.ProcessData(tagGenCountCount); len(bs) > 0 {
		Count := binutil.LittleEndian.Uint32(bs)
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
func (p *Formulator) GetStakingAmount(loader types.Loader, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	lw := types.NewLoaderWrapper(p.pid, loader)

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
			Count = binutil.LittleEndian.Uint32(bs)
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), binutil.LittleEndian.Uint32ToBytes(Count))
		ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count), StakingAddress[:])
		Count++
		ctw.SetAccountData(HyperAddress, tagStakingAmountCount, binutil.LittleEndian.Uint32ToBytes(Count))
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
				Count = binutil.LittleEndian.Uint32(bs)
			}
			Number := binutil.LittleEndian.Uint32(ns)
			if Number != Count-1 {
				var swapAddr common.Address
				copy(swapAddr[:], ctw.AccountData(HyperAddress, toStakingAmountReverseKey(Count-1)))
				ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Number), swapAddr[:])
				ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(swapAddr), binutil.LittleEndian.Uint32ToBytes(Number))
			}
			ctw.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), nil)
			ctw.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count-1), nil)
			Count--
			if Count == 0 {
				ctw.SetAccountData(HyperAddress, tagStakingAmountCount, nil)
			} else {
				ctw.SetAccountData(HyperAddress, tagStakingAmountCount, binutil.LittleEndian.Uint32ToBytes(Count))
			}
		}
		ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), nil)
	} else {
		ctw.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), total.Bytes())
	}
	return nil
}

// GetStakingAmountMap returns all staking amount of the hyper formulator
func (p *Formulator) GetStakingAmountMap(loader types.Loader, HyperAddress common.Address) (map[common.Address]*amount.Amount, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	PowerMap := map[common.Address]*amount.Amount{}
	if bs := lw.AccountData(HyperAddress, tagStakingAmountCount); len(bs) > 0 {
		Count := binutil.LittleEndian.Uint32(bs)
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
		return binutil.LittleEndian.Uint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastPaidHeight(ctw *types.ContextWrapper, lastPaidHeight uint32) {
	ctw.SetProcessData(tagLastPaidHeight, binutil.LittleEndian.Uint32ToBytes(lastPaidHeight))
}

func (p *Formulator) getLastStakingPaidHeight(lw types.LoaderWrapper, Address common.Address) uint32 {
	if bs := lw.AccountData(Address, tagLastStakingPaidHeight); len(bs) > 0 {
		return binutil.LittleEndian.Uint32(bs)
	} else {
		return 0
	}
}

func (p *Formulator) setLastStakingPaidHeight(ctw *types.ContextWrapper, Address common.Address, lastPaidHeight uint32) {
	ctw.SetAccountData(Address, tagLastStakingPaidHeight, binutil.LittleEndian.Uint32ToBytes(lastPaidHeight))
}

// GetUserAutoStaking returns the user auto staking status of the address
func (p *Formulator) GetUserAutoStaking(loader types.Loader, HyperAddress common.Address, StakingAddress common.Address) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

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

// GetRevokedFormulatorHeight returns the revoke height of the formulator
func (p *Formulator) GetRevokedFormulatorHeight(loader types.Loader, addr common.Address) (uint32, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.AccountData(addr, tagRevokedHeight); len(bs) > 0 {
		return binutil.LittleEndian.Uint32(bs), nil
	} else {
		return 0, ErrNotRevoked
	}
}

func (p *Formulator) getRevokedFormulatorHeritor(lw types.LoaderWrapper, addr common.Address, RevokeHeight uint32) (common.Address, error) {
	if bs := lw.ProcessData(toRevokedFormulatorKey(RevokeHeight, addr)); len(bs) > 0 {
		var Heritor common.Address
		copy(Heritor[:], bs)
		return Heritor, nil
	} else {
		return common.Address{}, ErrNotExistRevokedFormulator
	}
}

func (p *Formulator) addRevokedFormulator(ctw *types.ContextWrapper, addr common.Address, RevokeHeight uint32, Heritor common.Address) error {
	if bs := ctw.AccountData(addr, tagRevokedHeight); len(bs) > 0 {
		return ErrRevokedFormulator
	}
	ctw.SetAccountData(addr, tagRevokedHeight, binutil.LittleEndian.Uint32ToBytes(RevokeHeight))
	if ns := ctw.ProcessData(toRevokedFormulatorNumberKey(RevokeHeight, addr)); len(ns) == 0 {
		var Count uint32
		if bs := ctw.ProcessData(toRevokedFormulatorCountKey(RevokeHeight)); len(bs) > 0 {
			Count = binutil.LittleEndian.Uint32(bs)
		}
		ctw.SetProcessData(toRevokedFormulatorNumberKey(RevokeHeight, addr), binutil.LittleEndian.Uint32ToBytes(Count))
		ctw.SetProcessData(toRevokedFormulatorReverseKey(RevokeHeight, Count), addr[:])
		Count++
		ctw.SetProcessData(toRevokedFormulatorCountKey(RevokeHeight), binutil.LittleEndian.Uint32ToBytes(Count))
	}
	ctw.SetProcessData(toRevokedFormulatorKey(RevokeHeight, addr), Heritor[:])
	return nil
}

func (p *Formulator) removeRevokedFormulator(ctw *types.ContextWrapper, addr common.Address) error {
	RevokeHeight, err := p.GetRevokedFormulatorHeight(ctw, addr)
	if err != nil {
		return err
	}
	ctw.SetAccountData(addr, tagRevokedHeight, nil)

	ns := ctw.ProcessData(toRevokedFormulatorNumberKey(RevokeHeight, addr))
	if len(ns) == 0 {
		return ErrNotRevoked
	}
	var Count uint32
	if bs := ctw.ProcessData(toRevokedFormulatorCountKey(RevokeHeight)); len(bs) > 0 {
		Count = binutil.LittleEndian.Uint32(bs)
	}
	Number := binutil.LittleEndian.Uint32(ns)
	if Number != Count-1 {
		var swapAddr common.Address
		copy(swapAddr[:], ctw.ProcessData(toRevokedFormulatorReverseKey(RevokeHeight, Count-1)))
		ctw.SetProcessData(toRevokedFormulatorReverseKey(RevokeHeight, Number), swapAddr[:])
		ctw.SetProcessData(toRevokedFormulatorNumberKey(RevokeHeight, swapAddr), binutil.LittleEndian.Uint32ToBytes(Number))
	}
	ctw.SetProcessData(toRevokedFormulatorNumberKey(RevokeHeight, addr), nil)
	ctw.SetProcessData(toRevokedFormulatorReverseKey(RevokeHeight, Count-1), nil)
	Count--
	if Count == 0 {
		ctw.SetProcessData(toRevokedFormulatorCountKey(RevokeHeight), nil)
	} else {
		ctw.SetProcessData(toRevokedFormulatorCountKey(RevokeHeight), binutil.LittleEndian.Uint32ToBytes(Count))
	}
	return nil
}

func (p *Formulator) flushRevokedFormulatorMap(ctw *types.ContextWrapper, RevokeHeight uint32) (*types.AddressAddressMap, error) {
	RevokedFormulatorMap := types.NewAddressAddressMap()
	if bs := ctw.ProcessData(toRevokedFormulatorCountKey(RevokeHeight)); len(bs) > 0 {
		Count := binutil.LittleEndian.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], ctw.ProcessData(toRevokedFormulatorReverseKey(RevokeHeight, i)))
			Heritor, err := p.getRevokedFormulatorHeritor(ctw, addr, RevokeHeight)
			if err != nil {
				return nil, err
			}
			RevokedFormulatorMap.Put(addr, Heritor)

			ctw.SetAccountData(addr, tagRevokedHeight, nil)

			ctw.SetProcessData(toRevokedFormulatorKey(RevokeHeight, addr), nil)
			ctw.SetProcessData(toRevokedFormulatorNumberKey(RevokeHeight, addr), nil)
			ctw.SetProcessData(toRevokedFormulatorReverseKey(RevokeHeight, i), nil)
		}
		ctw.SetProcessData(toRevokedFormulatorCountKey(RevokeHeight), nil)
	}
	return RevokedFormulatorMap, nil
}

func (p *Formulator) getUnstakingAmount(lw types.LoaderWrapper, HyperAddr common.Address, addr common.Address, UnstakedHeight uint32) (*amount.Amount, error) {
	mp, err := p.GetUnstakingAmountMap(lw, addr, UnstakedHeight)
	if err != nil {
		return nil, err
	}
	am, has := mp.Get(HyperAddr)
	if !has {
		return nil, ErrNotExistUnstakingAmount
	}
	return am, nil
}

// GetUnstakingAmountMap returns the amount map of the unstaking
func (p *Formulator) GetUnstakingAmountMap(loader types.Loader, addr common.Address, UnstakedHeight uint32) (*types.AddressAmountMap, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toUnstakingAmountKey(UnstakedHeight, addr)); len(bs) > 0 {
		mp := types.NewAddressAmountMap()
		if err := encoding.Unmarshal(bs, &mp); err != nil {
			return nil, err
		}
		return mp, nil
	} else {
		return nil, ErrNotExistUnstakingAmount
	}
}

func (p *Formulator) addUnstakingAmount(ctw *types.ContextWrapper, HyperAddr common.Address, addr common.Address, UnstakedHeight uint32, am *amount.Amount) error {
	if ns := ctw.ProcessData(toUnstakingAmountNumberKey(UnstakedHeight, addr)); len(ns) == 0 {
		var Count uint32
		if bs := ctw.ProcessData(toUnstakingAmountCountKey(UnstakedHeight)); len(bs) > 0 {
			Count = binutil.LittleEndian.Uint32(bs)
		}
		ctw.SetProcessData(toUnstakingAmountNumberKey(UnstakedHeight, addr), binutil.LittleEndian.Uint32ToBytes(Count))
		ctw.SetProcessData(toUnstakingAmountReverseKey(UnstakedHeight, Count), addr[:])
		Count++
		ctw.SetProcessData(toUnstakingAmountCountKey(UnstakedHeight), binutil.LittleEndian.Uint32ToBytes(Count))
	}
	mp, err := p.GetUnstakingAmountMap(ctw, addr, UnstakedHeight)
	if err != nil {
		if err != ErrNotExistUnstakingAmount {
			return err
		}
		mp = types.NewAddressAmountMap()
	}
	sum, has := mp.Get(HyperAddr)
	if !has {
		sum = amount.NewCoinAmount(0, 0)
	}
	mp.Put(HyperAddr, sum.Add(am))
	data, err := encoding.Marshal(mp)
	if err != nil {
		return err
	}
	ctw.SetProcessData(toUnstakingAmountKey(UnstakedHeight, addr), data)
	return nil
}

func (p *Formulator) subUnstakingAmount(ctw *types.ContextWrapper, HyperAddr common.Address, addr common.Address, UnstakedHeight uint32, am *amount.Amount) error {
	mp, err := p.GetUnstakingAmountMap(ctw, addr, UnstakedHeight)
	if err != nil {
		return err
	}
	sum, has := mp.Get(HyperAddr)
	if !has {
		return ErrNotExistUnstakingAmount
	}
	if sum.Less(am) {
		return ErrMinustUnstakingAmount
	}
	if sum.IsZero() {
		mp.Delete(HyperAddr)
	} else {
		mp.Put(HyperAddr, sum.Sub(am))
	}
	if mp.Len() > 0 {
		data, err := encoding.Marshal(mp)
		if err != nil {
			return err
		}
		ctw.SetProcessData(toUnstakingAmountKey(UnstakedHeight, addr), data)
	} else {
		p.removeUnstakingAmount(ctw, addr, UnstakedHeight)
	}
	return nil
}

func (p *Formulator) removeUnstakingAmount(ctw *types.ContextWrapper, addr common.Address, UnstakedHeight uint32) error {
	ns := ctw.ProcessData(toUnstakingAmountNumberKey(UnstakedHeight, addr))
	if len(ns) == 0 {
		return ErrNotExistUnstakingAmount
	}
	var Count uint32
	if bs := ctw.ProcessData(toUnstakingAmountCountKey(UnstakedHeight)); len(bs) > 0 {
		Count = binutil.LittleEndian.Uint32(bs)
	}
	Number := binutil.LittleEndian.Uint32(ns)
	if Number != Count-1 {
		var swapAddr common.Address
		copy(swapAddr[:], ctw.ProcessData(toUnstakingAmountReverseKey(UnstakedHeight, Count-1)))
		ctw.SetProcessData(toUnstakingAmountReverseKey(UnstakedHeight, Number), swapAddr[:])
		ctw.SetProcessData(toUnstakingAmountNumberKey(UnstakedHeight, swapAddr), binutil.LittleEndian.Uint32ToBytes(Number))
	}
	ctw.SetProcessData(toUnstakingAmountNumberKey(UnstakedHeight, addr), nil)
	ctw.SetProcessData(toUnstakingAmountReverseKey(UnstakedHeight, Count-1), nil)
	Count--
	if Count == 0 {
		ctw.SetProcessData(toUnstakingAmountCountKey(UnstakedHeight), nil)
	} else {
		ctw.SetProcessData(toUnstakingAmountCountKey(UnstakedHeight), binutil.LittleEndian.Uint32ToBytes(Count))
	}
	return nil
}

func (p *Formulator) flushUnstakingAmountMap(ctw *types.ContextWrapper, RevokeHeight uint32) (*types.AddressAddressAmountMap, error) {
	UnstakingAmountMap := types.NewAddressAddressAmountMap()
	if bs := ctw.ProcessData(toUnstakingAmountCountKey(RevokeHeight)); len(bs) > 0 {
		Count := binutil.LittleEndian.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], ctw.ProcessData(toUnstakingAmountReverseKey(RevokeHeight, i)))
			mp, err := p.GetUnstakingAmountMap(ctw, addr, RevokeHeight)
			if err != nil {
				return nil, err
			}
			UnstakingAmountMap.Put(addr, mp)

			ctw.SetProcessData(toUnstakingAmountKey(RevokeHeight, addr), nil)
			ctw.SetProcessData(toUnstakingAmountNumberKey(RevokeHeight, addr), nil)
			ctw.SetProcessData(toUnstakingAmountReverseKey(RevokeHeight, i), nil)
		}
		ctw.SetProcessData(toUnstakingAmountCountKey(RevokeHeight), nil)
	}
	return UnstakingAmountMap, nil
}

// GetRewardPolicy returns the reward policy
func (p *Formulator) GetRewardPolicy(loader types.Loader) (*RewardPolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	policy := &RewardPolicy{}
	if err := encoding.Unmarshal(lw.ProcessData(tagRewardPolicy), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetAlphaPolicy returns the alpha policy
func (p *Formulator) GetAlphaPolicy(loader types.Loader) (*AlphaPolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	policy := &AlphaPolicy{}
	if err := encoding.Unmarshal(lw.ProcessData(tagAlphaPolicy), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetSigmaPolicy returns the sigma policy
func (p *Formulator) GetSigmaPolicy(loader types.Loader) (*SigmaPolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	policy := &SigmaPolicy{}
	if err := encoding.Unmarshal(lw.ProcessData(tagSigmaPolicy), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetOmegaPolicy returns the omega policy
func (p *Formulator) GetOmegaPolicy(loader types.Loader) (*OmegaPolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	policy := &OmegaPolicy{}
	if err := encoding.Unmarshal(lw.ProcessData(tagOmegaPolicy), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetHyperPolicy returns the hyper policy
func (p *Formulator) GetHyperPolicy(loader types.Loader) (*HyperPolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(lw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetTransmutePolicy returns the transmute policy
func (p *Formulator) GetTransmutePolicy(loader types.Loader) (*TransmutePolicy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.ProcessData(tagTransmutePolicy)
	if len(bs) == 0 {
		return nil, ErrNotExistTransmutePolicy
	}

	policy := &TransmutePolicy{}
	if err := encoding.Unmarshal(bs, &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// GetRewardCount returns the reward count of the formulator
func (p *Formulator) GetRewardCount(loader types.Loader, addr common.Address) (uint32, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	a, err := lw.Account(addr)
	if err != nil {
		return 0, err
	}
	acc, is := a.(*FormulatorAccount)
	if !is {
		return 0, ErrInvalidFormulatorAddress
	}

	Begin := acc.UpdatedHeight / 172800
	if acc.UpdatedHeight%172800 != 0 {
		Begin++
	}
	End := lw.TargetHeight() / 172800

	var RewardCount uint32
	for h := Begin; h <= End; h++ {
		evs, err := p.cn.Events(h*172800, h*172800)
		if err != nil {
			return 0, err
		}
		for _, v := range evs {
			switch ev := v.(type) {
			case *RewardEvent:
				if cnt, has := ev.GenBlockMap.Get(addr); has {
					if cnt > 0 {
						RewardCount++
					}
				}
			}
		}
	}
	return RewardCount, nil
}

// IsRewardBaseUpgrade returns reward base upgrade on/off
func (p *Formulator) IsRewardBaseUpgrade(loader types.Loader) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.ProcessData(tagRewardBaseUpgrade)
	return len(bs) > 0 && bs[0] == 1
}

func (p *Formulator) setRewardBaseUpgrade(ctw *types.ContextWrapper, Enable bool) {
	if Enable {
		ctw.SetProcessData(tagRewardBaseUpgrade, []byte{1})
	} else {
		ctw.SetProcessData(tagRewardBaseUpgrade, nil)
	}
}

func (p *Formulator) revokeFormulator(ctw *types.ContextWrapper, FormulatorAddr common.Address, Heritor common.Address) error {
	acc, err := ctw.Account(FormulatorAddr)
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if has, err := ctw.HasAccount(Heritor); err != nil {
		if err == types.ErrDeletedAccount {
		} else {
			return err
		}
	} else if !has {
	} else {
		if err := p.vault.AddBalance(ctw, Heritor, frAcc.Amount); err != nil {
			return err
		}
		if err := p.vault.AddBalance(ctw, Heritor, p.vault.Balance(ctw, FormulatorAddr)); err != nil {
			return err
		}
		if err := p.vault.RemoveBalance(ctw, FormulatorAddr); err != nil {
			return err
		}
	}
	if frAcc.FormulatorType == HyperFormulatorType {
		StakingAmountMap, err := p.GetStakingAmountMap(ctw, FormulatorAddr)
		if err != nil {
			return err
		}
		for addr, StakingAmount := range StakingAmountMap {
			if StakingAmount.IsZero() {
				return ErrInvalidStakingAddress
			}
			if frAcc.StakingAmount.Less(StakingAmount) {
				return ErrCriticalStakingAmount
			}
			frAcc.StakingAmount = frAcc.StakingAmount.Sub(StakingAmount)

			if err := p.vault.AddBalance(ctw, addr, StakingAmount); err != nil {
				return err
			}
		}
		if !frAcc.StakingAmount.IsZero() {
			return ErrCriticalStakingAmount
		}
	}
	if err := ctw.DeleteAccount(frAcc); err != nil {
		return err
	}
	ev := &RevokedEvent{
		Height_:    ctw.TargetHeight(),
		Index_:     65535,
		Formulator: FormulatorAddr,
	}
	if err := ctw.EmitEvent(ev); err != nil {
		return err
	}
	return nil
}
