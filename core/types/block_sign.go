package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

// BlockSign is the generator signature of the block
type BlockSign struct {
	HeaderHash         hash.Hash256
	GeneratorSignature common.Signature
}
