package chain

import (
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// Provider is a interface to give a chain data
type Provider interface {
	Version() uint16
	Height() uint32
	LastHash() hash.Hash256
	LastTimestamp() uint64
	Hash(height uint32) (hash.Hash256, error)
	Header(height uint32) (*types.Header, error)
	Block(height uint32) (*types.Block, error)
}
