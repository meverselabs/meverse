package bank

import "errors"

// errors
var (
	ErrExistKeyName       = errors.New("exist key name")
	ErrInvalidTag         = errors.New("invalid key tag")
	ErrInvalidNameAddress = errors.New("invalid name address")
)
