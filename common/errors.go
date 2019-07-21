package common

import (
	"errors"
)

// common errors
var (
	ErrInvalidAddressFormat    = errors.New("invalid address format")
	ErrInvalidAddressCheckSum  = errors.New("invalid address checksum")
	ErrInvalidSignatureFormat  = errors.New("invalid signature format")
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrInvalidPublicKeyFormat  = errors.New("invalid public key format")
	ErrInvalidPublicHash       = errors.New("invalid public hash")
	ErrInvalidPublicHashFormat = errors.New("invalid public hash format")
	ErrInsufficientSignature   = errors.New("insufficient signature")
	ErrDuplicatedSignature     = errors.New("duplicated signature")
)
