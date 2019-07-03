package vault

import "errors"

// errors
var (
	ErrMinusInput                  = errors.New("minus input")
	ErrMinusBalance                = errors.New("minus balance")
	ErrInvalidMultiKeyHashCount    = errors.New("invalid multi key hash count")
	ErrInvalidRequiredKeyHashCount = errors.New("invalid required key hash count")
)
