package p2p

import "errors"

// errors
var (
	ErrInvalidHandshake           = errors.New("invalid handshake")
	ErrInvalidLength              = errors.New("invalid length")
	ErrUnknownMessage             = errors.New("unknown message")
	ErrNotExistPeer               = errors.New("not exist peer")
	ErrSelfConnection             = errors.New("self connection")
	ErrInvalidUTXO                = errors.New("invalid UTXO")
	ErrTooManyTrasactionInMessage = errors.New("too many transaction in message")
	ErrExistSerializableType      = errors.New("exist serializable type")
	ErrInvalidSerializableTypeID  = errors.New("invalid serializable type id")
)
