package apiserver

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// Argument parses rpc arguments
type Argument struct {
	args []interface{}
}

// NewArgument returns a Argument
func NewArgument(args []interface{}) *Argument {
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
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}

	n, err := strconv.ParseInt(fmt.Sprintf("%v", a), 10, 32)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return int(n), nil
}

// Uint8 returns a uint8 value of the index
func (arg *Argument) Uint8(index int) (uint8, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseUint(fmt.Sprintf("%v", a), 10, 8)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return uint8(n), nil
}

// Uint16 returns a uint16 value of the index
func (arg *Argument) Uint16(index int) (uint16, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseUint(fmt.Sprintf("%v", a), 10, 16)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return uint16(n), nil
}

// Uint32 returns a uint32 value of the index
func (arg *Argument) Uint32(index int) (uint32, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseUint(fmt.Sprintf("%v", a), 10, 32)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return uint32(n), nil
}

// Uint64 returns a uint64 value of the index
func (arg *Argument) Uint64(index int) (uint64, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseUint(fmt.Sprintf("%v", a), 10, 64)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return uint64(n), nil
}

// Float32 returns a float32 value of the index
func (arg *Argument) Float32(index int) (float32, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 32)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return float32(n), nil
}

// Float64 returns a float64 value of the index
func (arg *Argument) Float64(index int) (float64, error) {
	if index < 0 || index >= len(arg.args) {
		return 0, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return 0, errors.WithStack(ErrInvalidArgumentType)
	}
	n, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return float64(n), nil
}

// Strings returns a string value of the index
func (arg *Argument) String(index int) (string, error) {
	if index < 0 || index >= len(arg.args) {
		return "", errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return "", errors.WithStack(ErrInvalidArgumentType)
	}
	return fmt.Sprintf("%v", a), nil
}

// Strings returns a string value of the index
func (arg *Argument) Array(index int) ([]interface{}, error) {
	if index < 0 || index >= len(arg.args) {
		return nil, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return nil, errors.WithStack(ErrInvalidArgumentType)
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(a)

		r := []interface{}{}
		for i := 0; i < s.Len(); i++ {
			r = append(r, s.Index(i).Interface())
		}
		return r, nil
	}
	return nil, errors.WithStack(ErrInvalidArgumentType)
}

// Strings returns a string value of the index
func (arg *Argument) Map(index int) (map[string]interface{}, error) {
	if index < 0 || index >= len(arg.args) {
		return nil, errors.WithStack(ErrInvalidArgumentIndex)
	}
	a := arg.args[index]
	if a == nil {
		return nil, errors.WithStack(ErrInvalidArgumentType)
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Map:
		s := reflect.ValueOf(a)

		r := map[string]interface{}{}
		s.MapKeys()
		mi := s.MapRange()
		for mi.Next() {
			k := mi.Key().Interface()
			v := mi.Value().Interface()
			r[fmt.Sprintf("%v", k)] = v
		}
		return r, nil
	}
	return nil, errors.WithStack(ErrInvalidArgumentType)
}
