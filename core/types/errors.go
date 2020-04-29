package types

import "errors"

// transaction errors
var (
	ErrInvalidAccountName            = errors.New("invalid account name")
	ErrInvalidAccountType            = errors.New("invalid account type")
	ErrExistAddress                  = errors.New("exist address")
	ErrExistAccount                  = errors.New("exist account")
	ErrExistAccountName              = errors.New("exist account name")
	ErrNotExistAccount               = errors.New("not exist account")
	ErrDeletedAccount                = errors.New("deleted account")
	ErrInvalidProcess                = errors.New("invalid process")
	ErrExistProcessName              = errors.New("exist process name")
	ErrExistProcessID                = errors.New("exist process id")
	ErrNotExistProcess               = errors.New("not exist process")
	ErrExistUTXO                     = errors.New("exist utxo")
	ErrNotExistUTXO                  = errors.New("not exist utxo")
	ErrUsedUTXO                      = errors.New("used utxo")
	ErrInvalidSequence               = errors.New("invalid sequence")
	ErrInvalidTransactionCount       = errors.New("invalid transaction count")
	ErrNotAllowedEmptyAddressAccount = errors.New("not allowed empty address account")
	ErrNotAllowedAddressAccountName  = errors.New("not allowed address account name")
	ErrInvalidAddressHeight          = errors.New("invalid address height")
	ErrInvalidSignerCount            = errors.New("invalid signer count")
	ErrInvalidAccountSigner          = errors.New("invalid account signer")
	ErrInvalidUTXOSigner             = errors.New("invalid utxo signer")
	ErrInvalidTxInCount              = errors.New("invalid tx in count")
	ErrInvalidOutputAmount           = errors.New("invalid output amount")
	ErrDustAmount                    = errors.New("dust amount")
	ErrInvalidTransactionIDFormat    = errors.New("invalid transaction id format")
)
