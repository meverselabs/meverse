package formulator

import "errors"

// consensus errors
var (
	ErrInvalidSignerCount    = errors.New("invalid signer count")
	ErrInvalidAccountSigner  = errors.New("invalid account signer")
	ErrInvalidAccountType    = errors.New("invalid account type")
	ErrInvalidRewardData     = errors.New("invalid reward data")
	ErrInvalidStakingAddress = errors.New("invalid staking address")
	ErrInvalidStakingAmount  = errors.New("invalid staking amount")
	ErrNotExistVault         = errors.New("not exist vault")
	ErrMinusInput            = errors.New("minus input")
	ErrMinusBalance          = errors.New("minus balance")

	/*
		ErrInvalidSequence                = errors.New("invalid sequence")
		ErrExistAccountName               = errors.New("exist account name")
		ErrInvalidAccountName             = errors.New("invaild account name")
	*/
	/*
		ErrInsufficientStakingAmount      = errors.New("insufficient staking amount")
		ErrExceedStakingAmount            = errors.New("exceed staking amount")
		ErrCriticalStakingAmount          = errors.New("critical staking amount")
		ErrInvalidFormulatorCount         = errors.New("invalid formulator count")
		ErrInsufficientFormulatorBlocks   = errors.New("insufficient formulator blocks")
		ErrNotExistFormulationPolicy      = errors.New("not exist formulator policy")
		ErrFormulatorCreationLimited      = errors.New("formulator creation limited")
		ErrUnauthorizedTransaction        = errors.New("unauthorized transaction")
	*/
)
