package types

import (
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
