package formulator

import "errors"

// consensus errors
var (
	ErrInvalidStakingAddress                   = errors.New("invalid staking address")
	ErrInvalidStakingAmount                    = errors.New("invalid staking amount")
	ErrAlphaCreationLimited                    = errors.New("alpha creation limited")
	ErrMinusInput                              = errors.New("minus input")
	ErrMinusBalance                            = errors.New("minus balance")
	ErrMinustUnstakingAmount                   = errors.New("minus unstaking amount")
	ErrRewardPolicyShouldBeSetupInApplication  = errors.New("RewardPolicy should be setup in application")
	ErrAlphaPolicyShouldBeSetupInApplication   = errors.New("AlphaPolicy should be setup in application")
	ErrSigmaPolicyShouldBeSetupInApplication   = errors.New("SigmaPolicy should be setup in application")
	ErrOmegaPolicyShouldBeSetupInApplication   = errors.New("OmegaPolicy should be setup in application")
	ErrHyperPolicyShouldBeSetupInApplication   = errors.New("HyperPolicy should be setup in application")
	ErrStakingPolicyShouldBeSetupInApplication = errors.New("StakingPolicy should be setup in application")
	ErrInvalidFormulatorAddress                = errors.New("invalid formulator address")
	ErrInvalidFormulatorCount                  = errors.New("invalid formulator count")
	ErrInsufficientFormulatorBlocks            = errors.New("insufficient formulator blocks")
	ErrCriticalStakingAmount                   = errors.New("critical staking amount")
	ErrInsufficientStakingAmount               = errors.New("insufficient staking amount")
	ErrInvalidRewardPolicy                     = errors.New("invalid reward policy")
	ErrInvalidValidatorPolicy                  = errors.New("invalid validator policy")
	ErrNotHyperFormulator                      = errors.New("not hyper formulator")
	ErrNotRevoked                              = errors.New("not revoked")
	ErrNotExistRevokedFormulator               = errors.New("not exist revoked formulator")
	ErrInvalidHeritor                          = errors.New("invalid heritor")
	ErrNotExistUnstakingAmount                 = errors.New("not exist unstaking amount")
	ErrRevokedFormulator                       = errors.New("revoked formulator")
	ErrInvalidTransmuteCount                   = errors.New("invalid transmute count")
	ErrInvalidTransmuteAmount                  = errors.New("invalid transmute amount")
	ErrInvalidTransmuteHeight                  = errors.New("invalid transmute height")
	ErrNotExistTransmutePolicy                 = errors.New("not exist transmute policy")
)
