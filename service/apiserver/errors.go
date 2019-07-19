package apiserver

import (
	"errors"
)

// errors
var (
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrInvalidArgumentIndex = errors.New("invalid argument index")
	ErrInvalidArgumentType  = errors.New("invalid argument type")
	ErrInvalidMethod        = errors.New("invalid method")
	ErrExistSubName         = errors.New("exist sub name")
)
