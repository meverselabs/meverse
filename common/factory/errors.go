package factory

import "errors"

// transaction errors
var (
	ErrExistType     = errors.New("exist type")
	ErrExistTypeName = errors.New("exist type name")
	ErrUnknownType   = errors.New("unknown type")
)
