package testlib

import "errors"

var (
	ErrFileRead = errors.New("File Read Error")
	ErrArgument = errors.New("Argument Error")
)
