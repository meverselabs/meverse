package poa

import "errors"

// consensus errors
var (
	ErrInvalidAuthorityKey   = errors.New("invalid authority key")
	ErrInvalidSignatureCount = errors.New("invalid signature count")
	ErrExistAddress          = errors.New("exist address")
	ErrFoundForkedBlockGen   = errors.New("found forked block gen")
	ErrNotExistClientPeer    = errors.New("not exist client peer")
	ErrNotExistAuthorityPeer = errors.New("not exist authority peer")
)
