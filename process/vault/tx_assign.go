package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// Assign moves a ownership of utxos
type Assign struct {
	Timestamp_ uint64
	Seq_       uint64
	Vin        []*types.TxIn
	Vout       []*types.TxOut
}

// Timestamp returns the timestamp of the transaction
func (tx *Assign) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Assign) Seq() uint64 {
	return tx.Seq_
}

// Fee returns the fee of the transaction
func (tx *Assign) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *Assign) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(tx.Vin) == 0 {
		return types.ErrInvalidTxInCount
	}
	if len(signers) > 1 {
		return types.ErrInvalidSignerCount
	}

	insum := amount.NewCoinAmount(0, 0)
	for _, vin := range tx.Vin {
		if utxo, err := loader.UTXO(vin.ID()); err != nil {
			return err
		} else {
			if utxo.PublicHash != signers[0] {
				return types.ErrInvalidUTXOSigner
			}
			insum = insum.Add(utxo.Amount)
		}
	}

	outsum := tx.Fee(loader)
	for _, vout := range tx.Vout {
		if vout.Amount.Less(amount.COIN.DivC(10)) {
			return types.ErrDustAmount
		}
		outsum = outsum.Add(vout.Amount)
	}

	if !insum.Equal(outsum) {
		return types.ErrInvalidOutputAmount
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Assign) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	insum := amount.NewCoinAmount(0, 0)
	for _, vin := range tx.Vin {
		if utxo, err := ctw.UTXO(vin.ID()); err != nil {
			return err
		} else {
			insum = insum.Add(utxo.Amount)
			if err := ctw.DeleteUTXO(utxo); err != nil {
				return err
			}
		}
	}

	for n, vout := range tx.Vout {
		if err := ctw.CreateUTXO(types.MarshalID(ctw.TargetHeight(), index, uint16(n)), vout); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Assign) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"vin":`)
	buffer.WriteString(`[`)
	for i, vin := range tx.Vin {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(vin.ID()); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`,`)
	buffer.WriteString(`"vout":`)
	buffer.WriteString(`[`)
	for i, vout := range tx.Vout {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := vout.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
