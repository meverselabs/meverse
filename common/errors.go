package common

import (
	"errors"
)

// common errors
var (
	ErrInvalidAddressFormat   = errors.New("invalid address format")
	ErrInvalidAddressCheckSum = errors.New("invalid address checksum")
	ErrInvalidSignatureFormat = errors.New("invalid signature format")
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrInvalidPublicKey       = errors.New("invalid public key")
	ErrInvalidPublicKeyFormat = errors.New("invalid public key format")
	ErrInsufficientSignature  = errors.New("insufficient signature")
	ErrDuplicatedSignature    = errors.New("duplicated signature")
)

type Causer interface {
	Cause() error
}
