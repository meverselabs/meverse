package types

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

// Provider defines functions that loads chain data from the chain
type Provider interface {
	ChainID() *big.Int
	Version() uint16
	Height() uint32
	LastHash() hash.Hash256
	LastTimestamp() uint64
	Hash(height uint32) (hash.Hash256, error)
	Header(height uint32) (*Header, error)
	Block(height uint32) (*Block, error)
	AddrSeq(addr common.Address) uint64
}
