package vault

import "errors"

// errors
var (
	ErrMinusInput                       = errors.New("minus input")
	ErrMinusBalance                     = errors.New("minus balance")
	ErrMinusCollectedFee                = errors.New("minus collected fee")
	ErrInvalidMultiKeyHashCount         = errors.New("invalid multi key hash count")
	ErrInvalidRequiredKeyHashCount      = errors.New("invalid required key hash count")
	ErrInvalidLockedBalanceKey          = errors.New("invalid locked balance key")
	ErrInvalidPolicy                    = errors.New("invalid policy")
	ErrInsufficientFee                  = errors.New("insufficient fee")
	ErrInsufficientBalance              = errors.New("insufficient balance")
	ErrNotExistFeeOfTransaction         = errors.New("not exist fee of transaction")
	ErrPolicyShouldBeSetupInApplication = errors.New("policy should be setup in application")
)
