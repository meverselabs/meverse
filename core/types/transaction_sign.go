package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

// TransactionSign is the signature of the transaction creator
type TransactionSign struct {
	TransactionHash hash.Hash256
	Signatures      []common.Signature
}
