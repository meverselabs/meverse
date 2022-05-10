package p2p

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

// Status represents the status of the peer
type Status struct {
	Height uint32
}

// TxMsgItem used to store transaction message
type TxMsgItem struct {
	TxHash hash.Hash256
	Tx     *types.Transaction
	Sig    common.Signature
	PeerID string
	ErrCh  *chan error
}

// RecvMessageItem used to store recv message
type RecvMessageItem struct {
	PeerID string
	Packet []byte
}

// SendMessageItem used to store send message
type SendMessageItem struct {
	Target  common.PublicKey
	Address common.Address
	Except  bool
	Packet  []byte
}
