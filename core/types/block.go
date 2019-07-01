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
		if err := enc.Encode(item.Header); err != nil {
			return err
		}
		if err := enc.EncodeArrayLen(len(item.Transactions)); err != nil {
			return err
		}
		for _, tx := range item.Transactions {
			if t, err := fc.TypeOf(tx); err != nil {
				return err
			} else if err := enc.EncodeUint16(t); err != nil {
				return err
			}
			if err := enc.Encode(tx); err != nil {
				return err
			}
		}
		if err := enc.EncodeArrayLen(len(item.Signatures)); err != nil {
			return err
		}
		for _, sigs := range item.Signatures {
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
		item := Block{}
		if err := dec.Decode(&item.Header); err != nil {
			return err
		}
		TxLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		for i := 0; i < TxLen; i++ {
			t, err := dec.DecodeUint16()
			if err != nil {
				return err
			}
			tx, err := fc.Create(t)
			if err != nil {
				return err
			}
			if err := dec.Decode(&tx); err != nil {
				return err
			}
			item.Transactions = append(item.Transactions, tx.(Transaction))
		}
		SigsLen, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		for i := 0; i < SigsLen; i++ {
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
		return nil
	})
}

// Block includes a block header and a block body
type Block struct {
	Header       Header
	Transactions []Transaction        //MAXLEN : 65535
	Signatures   [][]common.Signature //MAXLEN : 65536
}
