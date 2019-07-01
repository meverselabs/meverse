package hash

import (
	"errors"
)

// hash256 errors
var (
	ErrInvalidHashSize   = errors.New("invalid hash size")
	ErrInvalidHashFormat = errors.New("invalid hash format")
)
