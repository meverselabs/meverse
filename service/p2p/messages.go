package p2p

import (
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// message types
var (
	StatusMessageType          = types.DefineHashedType("p2p.StatusMessage")
	RequestMessageType         = types.DefineHashedType("p2p.RequestMessage")
	BlockMessageType           = types.DefineHashedType("p2p.BlockMessage")
	TransactionMessageType     = types.DefineHashedType("p2p.TransactionMessage")
	PeerListMessageType        = types.DefineHashedType("p2p.PeerListMessage")
	RequestPeerListMessageType = types.DefineHashedType("p2p.RequestPeerListMessage")
)

func init() {
	fc := encoding.Factory("transaction")
	encoding.Register(TransactionMessage{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(TransactionMessage)
		Len := len(item.Txs)
		if err := enc.EncodeArrayLen(Len); err != nil {
			return err
		}
		for i := 0; i < Len; i++ {
			if err := enc.EncodeUint16(item.Types[i]); err != nil {
				return err
			}
			if err := enc.Encode(item.Txs[i]); err != nil {
				return err
			}
			sigs := item.Signatures[i]
			if err := enc.EncodeArrayLen(len(sigs)); err != nil {
				return err
			}
			for _, sig := range sigs {
				if err := enc.Encode(sig); err != nil {
					return err
				}
			}
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := &TransactionMessage{}

		TxLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		if TxLen >= 65535 {
			return types.ErrInvalidTransactionCount
		}
		item.Types = make([]uint16, 0, TxLen)
		item.Txs = make([]types.Transaction, 0, TxLen)
		item.Signatures = make([][]common.Signature, 0, TxLen)
		for i := 0; i < TxLen; i++ {
			t, err := dec.DecodeUint16()
			if err != nil {
				return err
			}
			item.Types = append(item.Types, t)

			tx, err := fc.Create(t)
			if err != nil {
				return err
			}
			if err := dec.Decode(&tx); err != nil {
				return err
			}
			item.Txs = append(item.Txs, tx.(types.Transaction))

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
			item.Signatures = append(item.Signatures, sigs)
		}

		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

// RequestMessage used to request a chain data to a peer
type RequestMessage struct {
	Height uint32
	Count  uint8
}

// StatusMessage used to provide the chain information to a peer
type StatusMessage struct {
	Version  uint16
	Height   uint32
	LastHash hash.Hash256
}

// BlockMessage used to send a chain block to a peer
type BlockMessage struct {
	Blocks []*types.Block
}

// TransactionMessage is a message for a transaction
type TransactionMessage struct {
	Types      []uint16             //MAXLEN : 65535
	Txs        []types.Transaction  //MAXLEN : 65535
	Signatures [][]common.Signature //MAXLEN : 65535
}

// PeerListMessage is a message for a peer list
type PeerListMessage struct {
	Ips   []string
	Hashs []string
}

// RequestPeerListMessage is a request message for a peer list
type RequestPeerListMessage struct {
}
