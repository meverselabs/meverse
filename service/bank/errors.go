package bank

import "errors"

// errors
var (
	ErrExistKeyName           = errors.New("exist key name")
	ErrInvalidTag             = errors.New("invalid key tag")
	ErrInvalidNameAddress     = errors.New("invalid name address")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrInvalidTXID            = errors.New("invalid txid")
	ErrTransactionTimeout     = errors.New("transaction timeout")
	ErrTransactionFailed      = errors.New("transaction failed")
)
