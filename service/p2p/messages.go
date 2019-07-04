package p2p

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// TransactionMessage is a message for a transaction
type TransactionMessage struct {
	TxType uint8
	Tx     types.Transaction
	Sigs   []common.Signature
}

// HeaderMessage used to send a chain header to a peer
type HeaderMessage struct {
	Header     *types.Header
	Signatures []common.Signature
}

// BlockMessage used to send a chain block to a peer
type BlockMessage struct {
	Block *types.Block
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

// PingMessage is a message for a block generation
type PingMessage struct {
}
