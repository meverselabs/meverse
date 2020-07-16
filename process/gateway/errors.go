package gateway

import "errors"

// errors
var (
	ErrMinusInput                       = errors.New("minus input")
	ErrMinusBalance                     = errors.New("minus balance")
	ErrInvalidMultiKeyHashCount         = errors.New("invalid multi key hash count")
	ErrInvalidRequiredKeyHashCount      = errors.New("invalid required key hash count")
	ErrInvalidTokenAddressFormat        = errors.New("invalid token address format")
	ErrInvalidPolicy                    = errors.New("invalid policy")
	ErrInvalidAddressCount              = errors.New("invalid address count")
	ErrNotExistPolicy                   = errors.New("not exist policy")
	ErrProcessedTokenTXID               = errors.New("processed token txid")
	ErrProcessedOutTXID                 = errors.New("processed out txid")
	ErrPolicyShouldBeSetupInApplication = errors.New("policy should be setup in application")
)
