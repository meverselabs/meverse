package types

// UTXO represents usable coins in the UTXO model
type UTXO struct {
	*TxIn
	*TxOut
}

// Clone returns the clonend value of it
func (utxo *UTXO) Clone() *UTXO {
	return &UTXO{
		TxIn:  utxo.TxIn.Clone(),
		TxOut: utxo.TxOut.Clone(),
	}
}
