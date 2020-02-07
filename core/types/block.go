package types

import (
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
)

func init() {
	fc := encoding.Factory("transaction")
	encoding.Register(Block{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Block)

		if len(item.TransactionTypes) >= 65535 {
			return ErrInvalidTransactionCount
		}
		if len(item.TransactionTypes) != len(item.Transactions) {
			return ErrInvalidTransactionCount
		}
		if len(item.TransactionTypes) != len(item.TransactionSignatures) {
			return ErrInvalidTransactionCount
		}

		if err := enc.Encode(item.Header); err != nil {
			return err
		}
		Len := len(item.Transactions)
		if err := enc.EncodeArrayLen(Len); err != nil {
			return err
		}
		for i := 0; i < Len; i++ {
			if err := enc.EncodeUint16(item.TransactionTypes[i]); err != nil {
				return err
			}
			if err := enc.Encode(item.Transactions[i]); err != nil {
				return err
			}
			sigs := item.TransactionSignatures[i]
			if err := enc.EncodeArrayLen(len(sigs)); err != nil {
				return err
			}
			for _, sig := range sigs {
				if err := enc.Encode(sig); err != nil {
					return err
				}
			}
		}
		if err := enc.EncodeArrayLen(len(item.Signatures)); err != nil {
			return err
		}
		for _, sig := range item.Signatures {
			if err := enc.Encode(sig); err != nil {
				return err
			}
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := &Block{}
		if err := dec.Decode(&item.Header); err != nil {
			return err
		}
		TxLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		if TxLen >= 65535 {
			return ErrInvalidTransactionCount
		}
		item.TransactionTypes = make([]uint16, 0, TxLen)
		item.Transactions = make([]Transaction, 0, TxLen)
		item.TransactionSignatures = make([][]common.Signature, 0, TxLen)
		for i := 0; i < TxLen; i++ {
			t, err := dec.DecodeUint16()
			if err != nil {
				return err
			}
			item.TransactionTypes = append(item.TransactionTypes, t)

			tx, err := fc.Create(t)
			if err != nil {
				return err
			}
			if err := dec.Decode(&tx); err != nil {
				return err
			}
			item.Transactions = append(item.Transactions, tx.(Transaction))

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
			item.TransactionSignatures = append(item.TransactionSignatures, sigs)
		}
		SigLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		item.Signatures = make([]common.Signature, 0, SigLen)
		for j := 0; j < SigLen; j++ {
			var sig common.Signature
			if err := dec.Decode(&sig); err != nil {
				return err
			}
			item.Signatures = append(item.Signatures, sig)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

// Block includes a block header and a block body
type Block struct {
	Header                Header
	TransactionTypes      []uint16             //MAXLEN : 65535
	Transactions          []Transaction        //MAXLEN : 65535
	TransactionSignatures [][]common.Signature //MAXLEN : 65535
	Signatures            []common.Signature   //MAXLEN : 255
}
