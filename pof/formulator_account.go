package pof

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

type FormulatorAccount interface {
	types.Account
	IsFormulator() bool
	GeneratorHash() common.PublicHash
	IsActivated() bool
}
