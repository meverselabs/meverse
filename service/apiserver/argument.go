package apiserver

import (
	"strconv"
)

// Argument parses rpc arguments
type Argument struct {
	args []*string
}

// NewArgument returns a Argument
func NewArgument(args []*string) *Argument {
	arg := &Argument{
		args: args,
	}
	return arg
}

// Len returns length of arguments
func (arg *Argument) Len() int {
	return len(arg.args)
}

// Int returns a int value of the index
func (arg *Argument) Int(index int) (int, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseInt((*a), 10, 32)
	if err != nil {
		return 0, err
	}
	return int(n), err
}

// Uint8 returns a uint8 value of the index
func (arg *Argument) Uint8(index int) (uint8, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseUint((*a), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), err
}

// Uint16 returns a uint16 value of the index
func (arg *Argument) Uint16(index int) (uint16, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseUint((*a), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), err
}

// Uint32 returns a uint32 value of the index
func (arg *Argument) Uint32(index int) (uint32, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseUint((*a), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), err
}

// Uint64 returns a uint64 value of the index
func (arg *Argument) Uint64(index int) (uint64, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseUint((*a), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(n), err
}

// Float32 returns a float32 value of the index
func (arg *Argument) Float32(index int) (float32, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseFloat((*a), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), err
}

// Float64 returns a float64 value of the index
func (arg *Argument) Float64(index int) (float64, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return 0, ErrInvalidArgumentType
	}
	n, err := strconv.ParseFloat((*a), 64)
	if err != nil {
		return 0, err
	}
	return float64(n), err
}

// Strings returns a string value of the index
func (arg *Argument) String(index int) (string, error) {
	if index < 0 || index >= len(arg.args) {
		return "", ErrInvalidArgumentIndex
	}
	a := arg.args[index]
	if a == nil {
		return "", ErrInvalidArgumentType
	}
	return (*a), nil
}
