package types

import "errors"

// transaction errors
var (
	ErrInvalidAccountName      = errors.New("invalid account name")
	ErrInvalidAccountType      = errors.New("invalid account type")
	ErrExistAddress            = errors.New("exist address")
	ErrExistAccount            = errors.New("exist account")
	ErrExistAccountName        = errors.New("exist account name")
	ErrNotExistAccount         = errors.New("not exist account")
	ErrInvalidProcess          = errors.New("invalid process")
	ErrExistProcessName        = errors.New("exist process name")
	ErrExistProcessID          = errors.New("exist process id")
	ErrNotExistProcess         = errors.New("not exist process")
	ErrExistUTXO               = errors.New("exist utxo")
	ErrNotExistUTXO            = errors.New("not exist utxo")
	ErrUsedUTXO                = errors.New("used utxo")
	ErrInvalidSequence         = errors.New("invalid sequence")
	ErrInvalidTransactionCount = errors.New("invalid transaction count")
)
