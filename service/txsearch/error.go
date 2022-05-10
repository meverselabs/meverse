package txsearch

import (
	"errors"
	"fmt"
)

type ErrCannotReadBlock struct {
	err error
	tag string
}

func (e *ErrCannotReadBlock) Error() string {
	return fmt.Sprintln("CannotSetBlockHeight", e.tag, "err:", e.err)
}

type ErrCannotSetHeight struct {
	err error
	h   uint32
}

func (e *ErrCannotSetHeight) Error() string {
	return fmt.Sprintln("CannotSetHeight", e.h, "err:", e.err)
}

type ErrIsNotNextBlock struct {
}

func (e *ErrIsNotNextBlock) Error() string {
	return fmt.Sprintln("IsNotNextBlock")
}

var ErrFailTx = errors.New("failtx")
