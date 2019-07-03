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
	return nil
}

// InitPolicy called at OnInitGenesis of an application
func (p *Formulator) InitPolicy(ctw *types.ContextWrapper, rp *RewardPolicy, ap *AlphaPolicy, sp *SigmaPolicy, op *OmegaPolicy, hp *HyperPolicy) error {
	ctw = ctw.Switch(p.pid)

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
	policy := &RewardPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagRewardPolicy), &policy); err != nil {
		return err
	}

	if true {
		acc, err := ctw.Account(b.Header.Generator)
		if err != nil {
			return err
		}

		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return types.ErrInvalidAccountType
		}
		switch frAcc.FormulatorType {
		case AlphaFormulatorType:
			p.addRewardPower(ctw, b.Header.Generator, frAcc.Amount.MulC(int64(policy.AlphaEfficiency1000)).DivC(1000))
		case SigmaFormulatorType:
			p.addRewardPower(ctw, b.Header.Generator, frAcc.Amount.MulC(int64(policy.SigmaEfficiency1000)).DivC(1000))
		case OmegaFormulatorType:
			p.addRewardPower(ctw, b.Header.Generator, frAcc.Amount.MulC(int64(policy.OmegaEfficiency1000)).DivC(1000))
		case HyperFormulatorType:
			PowerSum := frAcc.Amount.MulC(int64(policy.HyperEfficiency1000)).DivC(1000)

			AmountMap, err := p.getStakingAmountMap(ctw, b.Header.Generator)
			if err != nil {
				return err
			}
			for StakingAddress, StakingAmount := range AmountMap {
				if StakingAmount.IsZero() {
					return ErrInvalidStakingAddress
				}

				if _, err := ctw.Account(StakingAddress); err != nil {
					if err != types.ErrNotExistAccount {
						return err
					}
					p.removeRewardPower(ctw, StakingAddress)
				} else {
					StakingPower := StakingAmount.MulC(int64(policy.StakingEfficiency1000)).DivC(1000)
					ComissionPower := StakingPower.MulC(int64(frAcc.Policy.CommissionRatio1000)).DivC(1000)

					if bs := ctw.AccountData(b.Header.Generator, toAutoStakingKey(StakingAddress)); len(bs) > 0 && bs[0] == 1 {
						p.addStakingPower(ctw, b.Header.Generator, StakingAddress, StakingPower.Sub(ComissionPower))
						PowerSum = PowerSum.Add(StakingPower)
					} else {
						p.addRewardPower(ctw, StakingAddress, StakingPower.Sub(ComissionPower))
						PowerSum = PowerSum.Add(ComissionPower)
					}
				}
			}
			p.addRewardPower(ctw, b.Header.Generator, PowerSum)
		default:
			return types.ErrInvalidAccountType
		}
	}

	lastPaidHeight := p.getLastPaidHeight(ctw)
	if ctw.TargetHeight() >= lastPaidHeight+policy.PayRewardEveryBlocks {
		TotalPower := amount.NewCoinAmount(0, 0)
		powerMap, err := p.getRewardPowerMap(ctw)
		if err != nil {
			return err
		}
		for _, PowerSum := range powerMap {
			TotalPower = TotalPower.Add(PowerSum)
		}
		TotalReward := policy.RewardPerBlock.MulC(int64(ctw.TargetHeight() - lastPaidHeight))
		Ratio := TotalReward.Mul(amount.COIN).Div(TotalPower)
		for addr, PowerSum := range powerMap {
			acc, err := ctw.Account(addr)
			if err != nil {
				if err != types.ErrNotExistAccount {
					return err
				}
			} else {
				frAcc := acc.(*FormulatorAccount)
				if err := p.vault.AddBalance(ctw, frAcc.Address(), PowerSum.Mul(Ratio).Div(amount.COIN)); err != nil {
					return err
				}
			}
			p.removeRewardPower(ctw, addr)
		}

		StakingPowerMap, err := p.getStakingPowerMap(ctw)
		if err != nil {
			return err
		}
		for HyperAddress, PowerMap := range StakingPowerMap {
			for StakingAddress, PowerSum := range PowerMap {
				p.addStakingAmount(ctw, HyperAddress, StakingAddress, PowerSum.Mul(Ratio).Div(amount.COIN))
				p.removeStakingPower(ctw, HyperAddress, StakingAddress)
			}
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
