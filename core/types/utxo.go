package types

// UTXO represents usable coins in the UTXO model
type UTXO struct {
	*TxIn
	*TxOut
}

// NewUTXO returns a UTXO
func NewUTXO() *UTXO {
	return &UTXO{
		TxIn:  NewTxIn(0),
		TxOut: NewTxOut(),
	}
}

// Clone returns the clonend value of it
func (utxo *UTXO) Clone() *UTXO {
	return &UTXO{
		TxIn:  utxo.TxIn.Clone(),
		TxOut: utxo.TxOut.Clone(),
	}
}
