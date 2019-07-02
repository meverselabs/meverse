package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// Transaction is an interface that defines common transaction functions
type Transaction interface {
	Timestamp() uint64
	Fee(loader LoaderProcess) *amount.Amount
	Validate(loader LoaderProcess, signers []common.PublicHash) error
	Execute(ctx *ContextProcess, index uint16) error
}
