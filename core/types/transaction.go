package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// Transaction is an interface that defines common transaction functions
type Transaction interface {
	Timestamp() uint64
	Fee(loader Loader) *amount.Amount
	Validate(loader Loader, signers []common.PublicHash) error
	Execute(ctx *Context, index uint16) error
}
