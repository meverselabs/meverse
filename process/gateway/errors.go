package gateway

import "errors"

// errors
var (
	ErrMinusInput                       = errors.New("minus input")
	ErrMinusBalance                     = errors.New("minus balance")
	ErrInvalidMultiKeyHashCount         = errors.New("invalid multi key hash count")
	ErrInvalidRequiredKeyHashCount      = errors.New("invalid required key hash count")
	ErrInvalidERC20AddressFormat        = errors.New("invalid erc20 address format")
	ErrInvalidPolicy                    = errors.New("invalid policy")
	ErrInvalidAddressCount              = errors.New("invalid address count")
	ErrProcessedERC20TXID               = errors.New("processed erc20 txid")
	ErrPolicyShouldBeSetupInApplication = errors.New("policy should be setup in application")
)
