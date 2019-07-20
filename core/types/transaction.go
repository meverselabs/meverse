package types

import (
	"encoding/json"

	"github.com/fletaio/fleta/common"
)

// Transaction defines common transaction functions
type Transaction interface {
	json.Marshaler
	Timestamp() uint64
	Validate(p Process, loader LoaderWrapper, signers []common.PublicHash) error
	Execute(p Process, ctx *ContextWrapper, index uint16) error
}
