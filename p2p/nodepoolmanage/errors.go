package nodepoolmanage

import "errors"

//router error list
var (
	ErrIsBanAddress       = errors.New("is ban address")
	ErrIsAlreadyConnected = errors.New("is already connected")
)
