package formulator

import "errors"

// consensus errors
var (
	ErrInvalidStakingAddress                   = errors.New("invalid staking address")
	ErrInvalidStakingAmount                    = errors.New("invalid staking amount")
	ErrAlphaCreationLimited                    = errors.New("alpha creation limited")
	ErrMinusInput                              = errors.New("minus input")
	ErrMinusBalance                            = errors.New("minus balance")
	ErrRewardPolicyShouldBeSetupInApplication  = errors.New("RewardPolicy should be setup in application")
	ErrAlphaPolicyShouldBeSetupInApplication   = errors.New("AlphaPolicy should be setup in application")
	ErrSigmaPolicyShouldBeSetupInApplication   = errors.New("SigmaPolicy should be setup in application")
	ErrOmegaPolicyShouldBeSetupInApplication   = errors.New("OmegaPolicy should be setup in application")
	ErrHyperPolicyShouldBeSetupInApplication   = errors.New("HyperPolicy should be setup in application")
	ErrStakingPolicyShouldBeSetupInApplication = errors.New("StakingPolicy should be setup in application")
	ErrInvalidFormulatorAddress                = errors.New("invalid formulator address")
	ErrInvalidFormulatorCount                  = errors.New("invalid formulator count")
	ErrInsufficientFormulatorBlocks            = errors.New("insufficient formulator blocks")
	ErrUnauthorizedTransaction                 = errors.New("unauthorized transaction")
	ErrCriticalStakingAmount                   = errors.New("critical staking amount")
	ErrInsufficientStakingAmount               = errors.New("insufficient staking amount")
	ErrInvalidPolicy                           = errors.New("invalid policy")
	ErrNotHyperFormulator                      = errors.New("not hyper formulator")
)
