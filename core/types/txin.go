package types

// TxIn represents the position of the UTXO
type TxIn struct {
	Height uint32
	Index  uint16
	N      uint16
}

// Clone returns the clonend value of it
func (in *TxIn) Clone() *TxIn {
	return &TxIn{
		Height: in.Height,
		Index:  in.Index,
		N:      in.N,
	}
}

// NewTxIn returns a TxIn
func NewTxIn(id uint64) *TxIn {
	if id == 0 {
		return &TxIn{}
	}
	height, index, n := UnmarshalID(id)
	return &TxIn{
		Height: height,
		Index:  index,
		N:      n,
	}
}

// ID returns the packed id of the txin
func (in *TxIn) ID() uint64 {
	return MarshalID(in.Height, in.Index, in.N)
}
