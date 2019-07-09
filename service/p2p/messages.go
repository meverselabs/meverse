package p2p

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// PingMessage is a message for a block generation
type PingMessage struct {
}

// RequestMessage used to request a chain data to a peer
type RequestMessage struct {
	Height uint32
}

// StatusMessage used to provide the chain information to a peer
type StatusMessage struct {
	Version  uint16
	Height   uint32
	LastHash hash.Hash256
}

// BlockMessage used to send a chain block to a peer
type BlockMessage struct {
	Block *types.Block
}

// TransactionMessage is a message for a transaction
type TransactionMessage struct {
	TxType uint16
	Tx     types.Transaction
	Sigs   []common.Signature
}
