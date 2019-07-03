package formulator

import "errors"

// consensus errors
var (
	ErrInvalidSignerCount                      = errors.New("invalid signer count")
	ErrInvalidAccountSigner                    = errors.New("invalid account signer")
	ErrInvalidAccountType                      = errors.New("invalid account type")
	ErrInvalidAccountName                      = errors.New("invaild account name")
	ErrInvalidSequence                         = errors.New("invalid sequence")
	ErrInvalidStakingAddress                   = errors.New("invalid staking address")
	ErrInvalidStakingAmount                    = errors.New("invalid staking amount")
	ErrExistAddress                            = errors.New("exist address")
	ErrExistAccountName                        = errors.New("exist account name")
	ErrNotExistAccount                         = errors.New("not exist account")
	ErrNotExistVault                           = errors.New("not exist vault")
	ErrNotExistPolicyData                      = errors.New("not exist reward data")
	ErrNotExistRewardData                      = errors.New("not exist reward data")
	ErrFormulatorCreationLimited               = errors.New("formulator creation limited")
	ErrMinusInput                              = errors.New("minus input")
	ErrMinusBalance                            = errors.New("minus balance")
	ErrRewardPolicyShouldBeSetupInApplication  = errors.New("RewardPolicy should be setup in application")
	ErrAlphaPolicyShouldBeSetupInApplication   = errors.New("AlphaPolicy should be setup in application")
	ErrSigmaPolicyShouldBeSetupInApplication   = errors.New("SigmaPolicy should be setup in application")
	ErrOmegaPolicyShouldBeSetupInApplication   = errors.New("OmegaPolicy should be setup in application")
	ErrHyperPolicyShouldBeSetupInApplication   = errors.New("HyperPolicy should be setup in application")
	ErrStakingPolicyShouldBeSetupInApplication = errors.New("StakingPolicy should be setup in application")

	/*
		ErrNotExistRewardPolicy  = errors.New("not exist formulator policy")
	*/
	/*
		ErrInsufficientStakingAmount      = errors.New("insufficient staking amount")
		ErrExceedStakingAmount            = errors.New("exceed staking amount")
		ErrCriticalStakingAmount          = errors.New("critical staking amount")
		ErrInvalidFormulatorCount         = errors.New("invalid formulator count")
		ErrInsufficientFormulatorBlocks   = errors.New("insufficient formulator blocks")
		ErrUnauthorizedTransaction        = errors.New("unauthorized transaction")
	*/
)
