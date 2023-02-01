package metamaskrelay

import "errors"

var (
	ErrInvalidContract = errors.New("invalid Contract")
	ErrInvalidData     = errors.New("invalid Data")
	ErrInvalidMethod   = errors.New("invalid Method")
	ErrInvalidType     = errors.New("invalid Type")
	ErrFileRead        = errors.New("File Read Error")
	ErrNotFoundEvent   = errors.New("Event Not Found")
	ErrArgument        = errors.New("Argument Error")
)
