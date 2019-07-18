package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/vault"
)

// Formulator manages balance of accounts of the chain
type Formulator struct {
	*types.ProcessBase
	pid          uint8
	pm           types.ProcessManager
	cn           types.Provider
	vault        *vault.Vault
	adminAddress common.Address
}

// NewFormulator returns a Formulator
func NewFormulator(pid uint8, AdminAddress common.Address) *Formulator {
	p := &Formulator{
		pid:          pid,
		adminAddress: AdminAddress,
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
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Formulator) InitPolicy(ctw *types.ContextWrapper, adminAddress common.Address, rp *RewardPolicy, ap *AlphaPolicy, sp *SigmaPolicy, op *OmegaPolicy, hp *HyperPolicy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if p.adminAddress != adminAddress {
		return ErrInvalidAdminAddress
	}
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

// OnLoadChain called when the chain loaded
func (p *Formulator) OnLoadChain(loader types.LoaderWrapper) error {
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
		CountMap, err := p.getGenCountMap(ctw)
		if err != nil {
			return err
		}

		RewardPowerSum := amount.NewCoinAmount(0, 0)
		RewardPowerMap := map[common.Address]*amount.Amount{}
		StakingRewardPowerMap := map[common.Address]*amount.Amount{}
		Hypers := []*FormulatorAccount{}
		for GenAddress, GenCount := range CountMap {
			p.removeGenCount(ctw, GenAddress)

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
				if bs := ctw.AccountData(b.Header.Generator, tagStakingAmountMap); len(bs) > 0 {
					if err := encoding.Unmarshal(bs, &PrevAmountMap); err != nil {
						return err
					}
				}
				AmountMap, err := p.GetStakingAmountMap(ctw, b.Header.Generator)
				if err != nil {
					return err
				}
				CurrentAmountMap := types.NewAddressAmountMap()
				CrossAmountMap := map[common.Address]*amount.Amount{}
				for StakingAddress, StakingAmount := range AmountMap {
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
				if bs, err := encoding.Marshal(CurrentAmountMap); err != nil {
					return err
				} else {
					ctw.SetAccountData(b.Header.Generator, tagStakingAmountMap, bs)
				}

				StakingRewardPower := amount.NewCoinAmount(0, 0)
				StakingPowerMap := types.NewAddressAmountMap()
				if bs := ctw.AccountData(b.Header.Generator, tagStakingPowerMap); len(bs) > 0 {
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
				if bs, err := encoding.Marshal(StakingPowerMap); err != nil {
					return err
				} else {
					ctw.SetAccountData(b.Header.Generator, tagStakingPowerMap, bs)
				}
				StakingRewardPowerMap[GenAddress] = StakingRewardPower
				RewardPowerSum = RewardPowerSum.Add(StakingRewardPower)
			default:
				return types.ErrInvalidAccountType
			}
		}

		StackRewardMap := types.NewAddressAmountMap()
		if bs := ctw.AccountData(b.Header.Generator, tagStackRewardMap); len(bs) > 0 {
			if err := encoding.Unmarshal(bs, &StackRewardMap); err != nil {
				return err
			}
		}
		if !RewardPowerSum.IsZero() {
			TotalReward := policy.RewardPerBlock.MulC(int64(ctw.TargetHeight() - lastPaidHeight))
			Ratio := TotalReward.Mul(amount.COIN).Div(RewardPowerSum)
			for RewardAddress, RewardPower := range RewardPowerMap {
				RewardAmount := RewardPower.Mul(Ratio).Div(amount.COIN)
				if !RewardAmount.IsZero() {
					if err := p.vault.AddBalance(ctw, RewardAddress, RewardAmount); err != nil {
						return err
					}
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
				}
			}
		}
		for _, frAcc := range Hypers {
			if StackReward, has := StackRewardMap.Get(frAcc.Address()); has {
				lastStakingPaidHeight := p.getLastStakingPaidHeight(ctw, frAcc.Address())
				if ctw.TargetHeight() >= lastStakingPaidHeight+frAcc.Policy.PayOutInterval {
					StakingPowerMap := types.NewAddressAmountMap()
					if bs := ctw.AccountData(b.Header.Generator, tagStakingPowerMap); len(bs) > 0 {
						if err := encoding.Unmarshal(bs, &StakingPowerMap); err != nil {
							return err
						}
					}

					StakingPowerSum := amount.NewCoinAmount(0, 0)
					StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
						StakingPowerSum = StakingPowerSum.Add(StakingPower)
						return true
					})
					if !StakingPowerSum.IsZero() {
						CommissionSum := amount.NewCoinAmount(0, 0)
						var inErr error
						Ratio := StackReward.Mul(amount.COIN).Div(StakingPowerSum)
						StakingPowerMap.EachAll(func(StakingAddress common.Address, StakingPower *amount.Amount) bool {
							RewardAmount := StakingPower.Mul(Ratio).Div(amount.COIN)
							if !RewardAmount.IsZero() {
								Commission := RewardAmount.MulC(int64(frAcc.Policy.CommissionRatio1000)).DivC(1000)
								CommissionSum = CommissionSum.Add(Commission)
								RewardAmount = RewardAmount.Sub(Commission)
								if p.getUserAutoStaking(ctw, frAcc.Address(), StakingAddress) {
									p.addStakingAmount(ctw, frAcc.Address(), StakingAddress, RewardAmount)
								} else {
									if err := p.vault.AddBalance(ctw, StakingAddress, RewardAmount); err != nil {
										inErr = err
										return false
									}
								}
							}
							return true
						})
						if inErr != nil {
							return inErr
						}

						if err := p.vault.AddBalance(ctw, frAcc.Address(), CommissionSum); err != nil {
							return err
						}
					}
					ctw.SetAccountData(b.Header.Generator, tagStakingPowerMap, nil)

					StackRewardMap.Delete(frAcc.Address())
					p.setLastStakingPaidHeight(ctw, frAcc.Address(), ctw.TargetHeight())
				}
			}
		}
		if bs, err := encoding.Marshal(StackRewardMap); err != nil {
			return err
		} else {
			ctw.SetAccountData(b.Header.Generator, tagStackRewardMap, bs)
		}

		//log.Println("Paid at", ctw.TargetHeight())
		p.setLastPaidHeight(ctw, ctw.TargetHeight())
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Formulator) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
