package types

import (
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// Transaction is an interface that defines common transaction functions
type Transaction interface {
	json.Marshaler
	Timestamp() uint64
	Fee(loader LoaderWrapper) *amount.Amount
	Validate(loader LoaderWrapper, signers []common.PublicHash) error
	Execute(p Process, ctx *ContextWrapper, index uint16) error
}
