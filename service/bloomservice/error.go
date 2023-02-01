package bloomservice

import "errors"

var (
	ErrInvalidType   = errors.New("invalid Type")
	ErrFileRead      = errors.New("File Read Error")
	ErrNotFoundEvent = errors.New("Event Not Found")
	ErrArgument      = errors.New("Argument Error")
)
