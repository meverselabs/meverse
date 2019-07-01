package factory

import "errors"

// transaction errors
var (
	ErrUnknownType   = errors.New("unknown type")
	ErrExistType     = errors.New("exist type")
	ErrExistTypeName = errors.New("exist type name")
)
