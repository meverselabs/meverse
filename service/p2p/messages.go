package p2p

import (
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

func init() {
	fc := encoding.Factory("transaction")
	encoding.Register(TransactionMessage{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(TransactionMessage)
		if err := enc.EncodeUint16(item.TxType); err != nil {
			return err
		}
		if err := enc.Encode(item.Tx); err != nil {
			return err
		}
		if err := enc.EncodeArrayLen(len(item.Sigs)); err != nil {
			return err
		}
		for _, sig := range item.Sigs {
			if err := enc.Encode(sig); err != nil {
				return err
			}
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := &TransactionMessage{}

		t, err := dec.DecodeUint16()
		if err != nil {
			return err
		}
		item.TxType = t

		tx, err := fc.Create(t)
		if err != nil {
			return err
		}
		if err := dec.Decode(&tx); err != nil {
			return err
		}
		item.Tx = tx.(types.Transaction)

		SigLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		sigs := make([]common.Signature, 0, SigLen)
		for j := 0; j < SigLen; j++ {
			var sig common.Signature
			if err := dec.Decode(&sig); err != nil {
				return err
			}
			sigs = append(sigs, sig)
		}
		item.Sigs = sigs

		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

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

// PeerListMessage is a message for a peer list
type PeerListMessage struct {
	Ips   []string
	Hashs []string
}

// RequestPeerListMessage is a request message for a peer list
type RequestPeerListMessage struct {
}
