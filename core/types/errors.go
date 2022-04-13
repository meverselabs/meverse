package types

import "errors"

// transaction errors
var (
	ErrExistContractType            = errors.New("exist contract type")
	ErrNotExistContract             = errors.New("not exist contract")
	ErrInvalidClassID               = errors.New("invalid class id")
	ErrInvalidContractTransactionID = errors.New("invalid contract transaction id")
	ErrInvalidTransactionCount      = errors.New("invalid transaction count")
	ErrInvalidSignerCount           = errors.New("invalid signer count")
	ErrInvalidSigner                = errors.New("invalid signer")
	ErrDustAmount                   = errors.New("dust amount")
	ErrInvalidTransactionIDFormat   = errors.New("invalid transaction id format")
	ErrUsedTimeSlot                 = errors.New("used timeslot")
	ErrInvalidTransactionTimeSlot   = errors.New("invalid transaction timeslot")
	ErrAlreadyAdmin                 = errors.New("already admin")
	ErrInvalidAdmin                 = errors.New("invalid admin")
	ErrAlreadyGenerator             = errors.New("already generator")
	ErrInvalidGenerator             = errors.New("invalid generator")
	ErrDirtyContext                 = errors.New("dirty context")
)
