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
	pm    types.ProcessManager
	cn    types.Provider
	vault *vault.Vault
}

// NewFormulator returns a Formulator
func NewFormulator() *Formulator {
	p := &Formulator{}
	return p
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

	reg.RegisterAccount(1, &FormulatorAccount{})
	if vp, err := pm.ProcessByName("fleta.vault"); err != nil {
		return ErrNotExistVault
	} else if v, is := vp.(*vault.Vault); !is {
		return ErrNotExistVault
	} else {
		p.vault = v
	}
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Formulator) InitPolicy(ctw *types.ContextWrapper, rp *RewardPolicy, ap *AlphaPolicy, sp *SigmaPolicy, op *OmegaPolicy, hp *HyperPolicy, kp *StakingPolicy) error {
	if bs, err := encoding.Marshal(rp); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("RewardPolicy"), bs)
	}
	if bs, err := encoding.Marshal(ap); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("AlphaPolicy"), bs)
	}
	if bs, err := encoding.Marshal(sp); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("SigmaPolicy"), bs)
	}
	if bs, err := encoding.Marshal(op); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("OmegaPolicy"), bs)
	}
	if bs, err := encoding.Marshal(hp); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("HyperPolicy"), bs)
	}
	if bs, err := encoding.Marshal(kp); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("StakingPolicy"), bs)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Formulator) OnLoadChain(loader types.LoaderWrapper) error {
	if bs := loader.ProcessData([]byte("RewardPolicy")); len(bs) == 0 {
		return ErrRewardPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData([]byte("AlphaPolicy")); len(bs) == 0 {
		return ErrAlphaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData([]byte("SigmaPolicy")); len(bs) == 0 {
		return ErrSigmaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData([]byte("OmegaPolicy")); len(bs) == 0 {
		return ErrOmegaPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData([]byte("HyperPolicy")); len(bs) == 0 {
		return ErrHyperPolicyShouldBeSetupInApplication
	}
	if bs := loader.ProcessData([]byte("StakingPolicy")); len(bs) == 0 {
		return ErrStakingPolicyShouldBeSetupInApplication
	}
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Formulator) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	if ctw.TargetHeight() == 1 {
		rd := newRewardData()
		if data, err := encoding.Marshal(&rd); err != nil {
			return err
		} else {
			ctw.SetProcessData([]byte("RewardData"), data)
		}
	}
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Formulator) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	policy := &RewardPolicy{}
	if bs := ctw.ProcessData([]byte("RewardPolicy")); len(bs) == 0 {
		return ErrNotExistPolicyData
	} else if err := encoding.Unmarshal(bs, &policy); err != nil {
		return err
	}
	rd := newRewardData()
	if bs := ctw.ProcessData([]byte("RewardData")); len(bs) == 0 {
		return ErrNotExistRewardData
	} else if err := encoding.Unmarshal(bs, &rd); err != nil {
		return err
	}

	if true {
		acc, err := ctw.Account(b.Header.Generator)
		if err != nil {
			return err
		}

		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return ErrInvalidAccountType
		}
		switch frAcc.FormulatorType {
		case AlphaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.AlphaEfficiency1000)).DivC(1000))
		case SigmaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.SigmaEfficiency1000)).DivC(1000))
		case OmegaFormulatorType:
			rd.addRewardPower(b.Header.Generator, frAcc.Amount.MulC(int64(policy.OmegaEfficiency1000)).DivC(1000))
		case HyperFormulatorType:
			PowerSum := frAcc.Amount.MulC(int64(policy.HyperEfficiency1000)).DivC(1000)

			keys, err := ctw.AccountDataKeys(b.Header.Generator, tagStaking)
			if err != nil {
				return err
			}
			for _, k := range keys {
				if StakingAddress, is := fromStakingKey(k); is {
					bs := ctw.AccountData(b.Header.Generator, k)
					if len(bs) == 0 {
						return ErrInvalidStakingAddress
					}
					StakingAmount := amount.NewAmountFromBytes(bs)

					if _, err := ctw.Account(StakingAddress); err != nil {
						if err != types.ErrNotExistAccount {
							return err
						}
						rd.removeRewardPower(StakingAddress)
					} else {
						StakingPower := StakingAmount.MulC(int64(policy.StakingEfficiency1000)).DivC(1000)
						ComissionPower := StakingPower.MulC(int64(frAcc.Policy.CommissionRatio1000)).DivC(1000)

						if bs := ctw.AccountData(b.Header.Generator, toAutoStakingKey(StakingAddress)); len(bs) > 0 && bs[0] == 1 {
							rd.addStakingPower(b.Header.Generator, StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(StakingPower)
						} else {
							rd.addRewardPower(StakingAddress, StakingPower.Sub(ComissionPower))
							PowerSum = PowerSum.Add(ComissionPower)
						}
					}
				}
			}
			rd.addRewardPower(b.Header.Generator, PowerSum)
		default:
			return ErrInvalidAccountType
		}
	}

	if ctw.TargetHeight() >= rd.lastPaidHeight+policy.PayRewardEveryBlocks {
		TotalPower := amount.NewCoinAmount(0, 0)
		rd.powerMap.EachAll(func(addr common.Address, PowerSum *amount.Amount) bool {
			TotalPower = TotalPower.Add(PowerSum)
			return true
		})
		TotalReward := policy.RewardPerBlock.MulC(int64(ctw.TargetHeight() - rd.lastPaidHeight))
		Ratio := TotalReward.Mul(amount.COIN).Div(TotalPower)
		var inErr error
		rd.powerMap.EachAll(func(RewardAddress common.Address, PowerSum *amount.Amount) bool {
			acc, err := ctw.Account(RewardAddress)
			if err != nil {
				if err != types.ErrNotExistAccount {
					inErr = err
					return false
				}
			} else {
				frAcc := acc.(*FormulatorAccount)
				if err := p.pm.SwitchProcess(ctw, p.vault, func(cts *types.ContextWrapper) error {
					if err := p.vault.AddBalance(cts, frAcc.Address(), PowerSum.Mul(Ratio).Div(amount.COIN)); err != nil {
						return err
					}
					return nil
				}); err != nil {
					inErr = err
					return false
				}
			}
			rd.removeRewardPower(RewardAddress)
			return true
		})
		if inErr != nil {
			return inErr
		}

		rd.stakingPowerMap.EachAll(func(HyperAddress common.Address, PowerMap *types.AddressAmountMap) bool {
			PowerMap.EachAll(func(StakingAddress common.Address, PowerSum *amount.Amount) bool {
				bs := ctw.AccountData(HyperAddress, toStakingKey(StakingAddress))
				if len(bs) == 0 {
					inErr = ErrInvalidStakingAddress
					return false
				}
				StakingAmount := amount.NewAmountFromBytes(bs)
				ctw.SetAccountData(HyperAddress, toStakingKey(StakingAddress), StakingAmount.Add(PowerSum.Mul(Ratio).Div(amount.COIN)).Bytes())
				return true
			})
			if inErr != nil {
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
		rd.stakingPowerMap = types.NewAddressAddressAmountMap()

		//log.Println("Paid at", ctw.TargetHeight())
		rd.lastPaidHeight = ctw.TargetHeight()
	}

	if data, err := encoding.Marshal(&rd); err != nil {
		return err
	} else {
		ctw.SetProcessData([]byte("RewardData"), data)
	}
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Formulator) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
