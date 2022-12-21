package test

import "errors"

var (
	AmountNotNegative = "Amount can't be negative"

	ErrInvalidType   = errors.New("invalid Type")
	ErrFileRead      = errors.New("File Read Error")
	ErrNotFoundEvent = errors.New("Event Not Found")
	ErrArgument      = errors.New("Argument Error")
)
