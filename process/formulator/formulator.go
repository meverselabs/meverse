package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/vault"
)

// Formulator manages balance of accounts of the chain
type Formulator struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	vault *vault.Vault
	admin *admin.Admin
}

// NewFormulator returns a Formulator
func NewFormulator(pid uint8) *Formulator {
	p := &Formulator{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Formulator) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Formulator) Name() string {
	return "fleta.formulator"
}

// Version returns the version of the process
func (p *Formulator) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Formulator) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("fleta.vault"); err != nil {
		return err
	} else if v, is := vp.(*vault.Vault); !is {
		return types.ErrInvalidProcess
	} else {
		p.vault = v
	}
	if vp, err := pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if v, is := vp.(*admin.Admin); !is {
		return types.ErrInvalidProcess
	} else {
		p.admin = v
	}

	reg.RegisterAccount(1, &FormulatorAccount{})
	reg.RegisterTransaction(1, &CreateAlpha{})
	reg.RegisterTransaction(2, &CreateSigma{})
	reg.RegisterTransaction(3, &CreateOmega{})
	reg.RegisterTransaction(4, &CreateHyper{})
	reg.RegisterTransaction(5, &Revoke{})
	reg.RegisterTransaction(6, &Staking{})
	reg.RegisterTransaction(7, &Unstaking{})
	reg.RegisterTransaction(8, &UpdateValidatorPolicy{})
	reg.RegisterTransaction(9, &UpdateUserAutoStaking{})
	reg.RegisterTransaction(10, &ChangeOwner{})
	reg.RegisterEvent(1, &RewardEvent{})
	reg.RegisterEvent(2, &RevokedEvent{})
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Formulator) InitPolicy(ctw *types.ContextWrapper, rp *RewardPolicy, ap *AlphaPolicy, sp *SigmaPolicy, op *OmegaPolicy, hp *HyperPolicy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs, err := encoding.Marshal(rp); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagRewardPolicy, bs)
	}
	if bs, err := encoding.Marshal(ap); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagAlphaPolicy, bs)
	}
	if bs, err := encoding.Marshal(sp); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagSigmaPolicy, bs)
	}
	if bs, err := encoding.Marshal(op); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagOmegaPolicy, bs)
	}
	if bs, err := encoding.Marshal(hp); err != nil {
		return err
	} else {
		ctw.SetProcessData(tagHyperPolicy, bs)
	}
	return nil
}

// InitStakingMap called at OnInitGenesis of an application
func (p *Formulator) InitStakingMap(ctw *types.ContextWrapper, HyperAddresses []common.Address) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	for _, addr := range HyperAddresses {
		AmountMap, err := p.GetStakingAmountMap(ctw, addr)
		if err != nil {
			return err
		}
		CurrentAmountMap := types.NewAddressAmountMap()
		for StakingAddress, StakingAmount := range AmountMap {
			CurrentAmountMap.Put(StakingAddress, StakingAmount)
		}
		if bs, err := encoding.Marshal(CurrentAmountMap); err != nil {
			return err
		} else {
			ctw.SetAccountData(addr, tagStakingAmountMap, bs)
		}
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Formulator) OnLoadChain(loader types.LoaderWrapper) error {
	p.admin.AdminAddress(loader, p.Name())
	if bs := loader.ProcessData(tagRewardPolicy); len(bs) == 0 {
		return ErrRewardPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData(tagAlphaPolicy); len(bs) == 0 {
		return ErrAlphaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData(tagSigmaPolicy); len(bs) == 0 {
		return ErrSigmaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData(tagOmegaPolicy); len(bs) == 0 {
		return ErrOmegaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData(tagHyperPolicy); len(bs) == 0 {
		return ErrHyperPolicyShouldBeSetupInApplication
	}
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Formulator) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Formulator) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	p.addGenCount(ctw, b.Header.Generator)

	policy := &RewardPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagRewardPolicy), &policy); err != nil {
		return err
	}

	lastPaidHeight := p.getLastPaidHeight(ctw)
	if ctw.TargetHeight() >= lastPaidHeight+policy.PayRewardEveryBlocks {
		CountMap, err := p.flushGenCountMap(ctw)
		if err != nil {
			return err
		}

		ev := &RewardEvent{
			Height_:        ctw.TargetHeight(),
			Index_:         65535,
			GenBlockMap:    types.NewAddressUint32Map(),
			RewardMap:      types.NewAddressAmountMap(),
			CommissionMap:  types.NewAddressAmountMap(),
			StackedMap:     types.NewAddressAmountMap(),
			StakedMap:      types.NewAddressAddressAmountMap(),
			StakeRewardMap: types.NewAddressAddressAmountMap(),
		}

		StackRewardMap := types.NewAddressAmountMap()
		if bs := ctw.ProcessData(tagStackRewardMap); len(bs) > 0 {
			if err := encoding.Unmarshal(bs, &StackRewardMap); err != nil {
				return err
			}
		}

		RewardPowerSum := amount.NewCoinAmount(0, 0)
		RewardPowerMap := map[common.Address]*amount.Amount{}
		StakingRewardPowerMap := map[common.Address]*amount.Amount{}
		Hypers := []*FormulatorAccount{}
		for GenAddress, GenCount := range CountMap {
			ev.GenBlockMap.Put(GenAddress, GenCount)

			if has, err := ctw.HasAccount(GenAddress); err != nil {
				return err
			} else if !has {
				StackRewardMap.Delete(GenAddress)
			} else {
				acc, err := ctw.Account(GenAddress)
				if err != nil {
					return err
				}
				frAcc, is := acc.(*FormulatorAccount)
				if !is {
					return types.ErrInvalidAccountType
				}
				switch frAcc.FormulatorType {
				case AlphaFormulatorType:
					am := frAcc.Amount.MulC(int64(GenCount)).MulC(int64(policy.AlphaEfficiency1000)).DivC(1000)
					RewardPowerSum = RewardPowerSum.Add(am)
					RewardPowerMap[GenAddress] = am
				case SigmaFormulatorType:
					am := frAcc.Amount.MulC(int64(GenCount)).MulC(int64(policy.SigmaEfficiency1000)).DivC(1000)
					RewardPowerSum = RewardPowerSum.Add(am)
					RewardPowerMap[GenAddress] = am
				case OmegaFormulatorType:
					am := frAcc.Amount.MulC(int64(GenCount)).MulC(int64(policy.OmegaEfficiency1000)).DivC(1000)
					RewardPowerSum = RewardPowerSum.Add(am)
					RewardPowerMap[GenAddress] = am
				case HyperFormulatorType:
					Hypers = append(Hypers, frAcc)

					am := frAcc.Amount.MulC(int64(GenCount)).MulC(int64(policy.HyperEfficiency1000)).DivC(1000)
					RewardPowerSum = RewardPowerSum.Add(am)
					RewardPowerMap[GenAddress] = am

					PrevAmountMap := types.NewAddressAmountMap()
					if bs := ctw.AccountData(frAcc.Address(), tagStakingAmountMap); len(bs) > 0 {
						if err := encoding.Unmarshal(bs, &PrevAmountMap); err != nil {
							return err
						}
					}
					AmountMap, err := p.GetStakingAmountMap(ctw, frAcc.Address())
					if err != nil {
						return err
					}
					CurrentAmountMap := types.NewAddressAmountMap()
					CrossAmountMap := map[common.Address]*amount.Amount{}
					for StakingAddress, StakingAmount := range AmountMap {
						if has, err := ctw.HasAccount(StakingAddress); err != nil {
							return err
						} else if !has {
							p.subStakingAmount(ctw, frAcc.Address(), StakingAddress, StakingAmount)
						} else {
							CurrentAmountMap.Put(StakingAddress, StakingAmount)
							if PrevStakingAmount, has := PrevAmountMap.Get(StakingAddress); has {
								if !PrevStakingAmount.IsZero() && !StakingAmount.IsZero() {
									if StakingAmount.Less(PrevStakingAmount) {
										CrossAmountMap[StakingAddress] = StakingAmount
									} else {
										CrossAmountMap[StakingAddress] = PrevStakingAmount
									}
								}
							}
						}
					}
					if bs, err := encoding.Marshal(CurrentAmountMap); err != nil {
						return err
					} else {
						ctw.SetAccountData(frAcc.Address(), tagStakingAmountMap, bs)
					}

					StakingRewardPower := amount.NewCoinAmount(0, 0)
					StakingPowerMap := types.NewAddressAmountMap()
					if bs := ctw.AccountData(frAcc.Address(), tagStakingPowerMap); len(bs) > 0 {
						if err := encoding.Unmarshal(bs, &StakingPowerMap); err != nil {
							return err
						}
					}
					for StakingAddress, StakingAmount := range CrossAmountMap {
						if sm, has := StakingPowerMap.Get(StakingAddress); has {
							StakingPowerMap.Put(StakingAddress, sm.Add(StakingAmount))
						} else {
							StakingPowerMap.Put(StakingAddress, StakingAmount)
						}
						StakingRewardPower = StakingRewardPower.Add(StakingAmount.MulC(int64(GenCount)).MulC(int64(policy.StakingEfficiency1000)).DivC(1000))
					}

					StackReward, has := StackRewardMap.Get(frAcc.Address())
					if has {
						StakingPowerSum := amount.NewCoinAmount(0, 0)
						Deleteds := []common.Address{}
						var inErr error
						StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
							if has, err := ctw.HasAccount(StakingAddress); err != nil {
								inErr = err
								return false
							} else if !has {
								Deleteds = append(Deleteds, StakingAddress)
							} else {
								StakingPowerSum = StakingPowerSum.Add(StakingPower)
							}
							return true
						})
						if inErr != nil {
							return inErr
						}
						for _, StakingAddress := range Deleteds {
							StakingPowerMap.Delete(StakingAddress)
						}
						if !StakingPowerSum.IsZero() {
							var inErr error
							Ratio := StackReward.Mul(amount.COIN).Div(StakingPowerSum)
							StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
								StackStakingAmount := StakingPower.Mul(Ratio).Div(amount.COIN)
								StakingPowerMap.Put(StakingAddress, StakingPower.Add(StackStakingAmount))
								StakingRewardPower = StakingRewardPower.Add(StackStakingAmount.MulC(int64(GenCount)).MulC(int64(policy.StakingEfficiency1000)).DivC(1000))
								return true
							})
							if inErr != nil {
								return inErr
							}
						}
					}

					if bs, err := encoding.Marshal(StakingPowerMap); err != nil {
						return err
					} else {
						ctw.SetAccountData(frAcc.Address(), tagStakingPowerMap, bs)
					}
					StakingRewardPowerMap[GenAddress] = StakingRewardPower
					RewardPowerSum = RewardPowerSum.Add(StakingRewardPower)
				default:
					return types.ErrInvalidAccountType
				}
			}
		}

		if !RewardPowerSum.IsZero() {
			TotalReward := policy.RewardPerBlock.MulC(int64(ctw.TargetHeight() - lastPaidHeight))
			TotalFee := p.vault.CollectedFee(ctw)
			if err := p.vault.SubCollectedFee(ctw, TotalFee); err != nil {
				return err
			}
			TotalReward = TotalReward.Add(TotalFee)

			Ratio := TotalReward.Mul(amount.COIN).Div(RewardPowerSum)
			for RewardAddress, RewardPower := range RewardPowerMap {
				RewardAmount := RewardPower.Mul(Ratio).Div(amount.COIN)
				if !RewardAmount.IsZero() {
					if err := p.vault.AddBalance(ctw, RewardAddress, RewardAmount); err != nil {
						return err
					}
					ev.AddReward(RewardAddress, RewardAmount)
				}
			}
			for GenAddress, StakingRewardPower := range StakingRewardPowerMap {
				if has, err := ctw.HasAccount(GenAddress); err != nil {
					return err
				} else if has {
					RewardAmount := StakingRewardPower.Mul(Ratio).Div(amount.COIN)
					if sm, has := StackRewardMap.Get(GenAddress); has {
						StackRewardMap.Put(GenAddress, sm.Add(RewardAmount))
					} else {
						StackRewardMap.Put(GenAddress, RewardAmount)
					}
					ev.AddStacked(GenAddress, RewardAmount)
				}
			}
		}
		for _, frAcc := range Hypers {
			if StackReward, has := StackRewardMap.Get(frAcc.Address()); has {
				ev.RemoveStacked(frAcc.Address())
				lastStakingPaidHeight := p.getLastStakingPaidHeight(ctw, frAcc.Address())
				if ctw.TargetHeight() >= lastStakingPaidHeight+policy.PayRewardEveryBlocks*frAcc.Policy.PayOutInterval {
					StakingPowerMap := types.NewAddressAmountMap()
					if bs := ctw.AccountData(frAcc.Address(), tagStakingPowerMap); len(bs) > 0 {
						if err := encoding.Unmarshal(bs, &StakingPowerMap); err != nil {
							return err
						}
					}

					StakingPowerSum := amount.NewCoinAmount(0, 0)
					var inErr error
					Deleteds := []common.Address{}
					StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
						if has, err := ctw.HasAccount(StakingAddress); err != nil {
							inErr = err
							return false
						} else if !has {
							Deleteds = append(Deleteds, StakingAddress)
						} else {
							StakingPowerSum = StakingPowerSum.Add(StakingPower)
						}
						return true
					})
					for _, StakingAddress := range Deleteds {
						StakingPowerMap.Delete(StakingAddress)
					}
					if inErr != nil {
						return inErr
					}
					if !StakingPowerSum.IsZero() {
						CommissionSum := amount.NewCoinAmount(0, 0)
						var inErr error
						Ratio := StackReward.Mul(amount.COIN).Div(StakingPowerSum)
						StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
							RewardAmount := StakingPower.Mul(Ratio).Div(amount.COIN)
							if frAcc.Policy.CommissionRatio1000 > 0 {
								Commission := RewardAmount.MulC(int64(frAcc.Policy.CommissionRatio1000)).DivC(1000)
								CommissionSum = CommissionSum.Add(Commission)
								RewardAmount = RewardAmount.Sub(Commission)
							}
							if !RewardAmount.IsZero() {
								if p.getUserAutoStaking(ctw, frAcc.Address(), StakingAddress) {
									p.AddStakingAmount(ctw, frAcc.Address(), StakingAddress, RewardAmount)
									ev.AddStaked(frAcc.Address(), StakingAddress, RewardAmount)
								} else {
									if err := p.vault.AddBalance(ctw, StakingAddress, RewardAmount); err != nil {
										inErr = err
										return false
									}
									ev.AddStakeReward(frAcc.Address(), StakingAddress, RewardAmount)
								}
							}
							return true
						})
						if inErr != nil {
							return inErr
						}

						if !CommissionSum.IsZero() {
							if err := p.vault.AddBalance(ctw, frAcc.Address(), CommissionSum); err != nil {
								return err
							}
							ev.AddCommission(frAcc.Address(), CommissionSum)
						}
					}
					ctw.SetAccountData(frAcc.Address(), tagStakingPowerMap, nil)

					StackRewardMap.Delete(frAcc.Address())
					p.setLastStakingPaidHeight(ctw, frAcc.Address(), ctw.TargetHeight())
				}
			}
		}
		if bs, err := encoding.Marshal(StackRewardMap); err != nil {
			return err
		} else {
			ctw.SetProcessData(tagStackRewardMap, bs)
		}

		if err := ctw.EmitEvent(ev); err != nil {
			return err
		}

		//log.Println("Paid at", ctw.TargetHeight())
		p.setLastPaidHeight(ctw, ctw.TargetHeight())
	}

	RevokedMap, err := p.flushRevokedFormulatorMap(ctw, ctw.TargetHeight())
	if err != nil {
		return err
	}
	for FormulatorAddr, Heritor := range RevokedMap {
		acc, err := ctw.Account(FormulatorAddr)
		if err != nil {
			return err
		}
		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return types.ErrInvalidAccountType
		}
		if has, err := ctw.HasAccount(Heritor); err != nil {
			return err
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
	}

	UnstakedMap, err := p.flushUnstakingAmountMap(ctw, ctw.TargetHeight())
	if err != nil {
		return err
	}
	for Addr, AmountMap := range UnstakedMap {
		var inErr error
		AmountMap.EachAll(func(HyperAddr common.Address, am *amount.Amount) bool {
			if err := p.vault.AddBalance(ctw, Addr, am); err != nil {
				inErr = err
				return false
			}
			ev := &UnstakedEvent{
				Height_:         ctw.TargetHeight(),
				Index_:          65535,
				HyperFormulator: HyperAddr,
				Address:         Addr,
				Amount:          am,
			}
			if err := ctw.EmitEvent(ev); err != nil {
				inErr = err
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
		if err := p.removeUnstakingAmount(ctw, Addr, ctw.TargetHeight()); err != nil {
			return err
		}
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Formulator) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
