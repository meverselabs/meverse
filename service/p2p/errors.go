package p2p

import "errors"

// errors
var (
	ErrInvalidHandshake = errors.New("invalid handshake")
	ErrInvalidLength    = errors.New("invalid length")
	ErrUnknownMessage   = errors.New("unknown message")
)
