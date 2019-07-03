package types

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// TxOut represents recipient of the UTXO
type TxOut struct {
	Amount     *amount.Amount
	PublicHash common.PublicHash
}

// NewTxOut returns a TxOut
func NewTxOut() *TxOut {
	out := &TxOut{
		Amount: amount.NewCoinAmount(0, 0),
	}
	return out
}

// Clone returns the clonend value of it
func (out *TxOut) Clone() *TxOut {
	return &TxOut{
		Amount:     out.Amount.Clone(),
		PublicHash: out.PublicHash.Clone(),
	}
}

// MarshalJSON is a marshaler function
func (out *TxOut) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"amount":`)
	if bs, err := out.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"public_hash":`)
	if bs, err := json.Marshal(out.PublicHash); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
