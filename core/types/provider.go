package types

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

// Provider defines functions that loads chain data from the chain
type Provider interface {
	ChainID() *big.Int
	Version(uint32) uint16
	Height() uint32
	InitHeight() uint32
	LastHash() hash.Hash256
	LastTimestamp() uint64
	Hash(uint32) (hash.Hash256, error)
	Header(uint32) (*Header, error)
	Block(uint32) (*Block, error)
	Receipts(uint32) (Receipts, error)
	AddrSeq(common.Address) uint64
}
