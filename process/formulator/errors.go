package formulator

import "errors"

// consensus errors
var (
	ErrInvalidSignerCount                      = errors.New("invalid signer count")
	ErrInvalidAccountSigner                    = errors.New("invalid account signer")
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
	ErrInvalidFormulatorCount                  = errors.New("invalid formulator count")
	ErrInsufficientFormulatorBlocks            = errors.New("insufficient formulator blocks")

	/*
		ErrInsufficientStakingAmount      = errors.New("insufficient staking amount")
		ErrExceedStakingAmount            = errors.New("exceed staking amount")
		ErrCriticalStakingAmount          = errors.New("critical staking amount")
		ErrUnauthorizedTransaction        = errors.New("unauthorized transaction")
	*/
)
