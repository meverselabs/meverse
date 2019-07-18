package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// OpenAccount moves a ownership of utxos
type OpenAccount struct {
	Timestamp_ uint64
	Seq_       uint64
	Vin        []*types.TxIn
	Vout       []*types.TxOut
	Name       string
	KeyHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *OpenAccount) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *OpenAccount) Seq() uint64 {
	return tx.Seq_
}

// Fee returns the fee of the transaction
func (tx *OpenAccount) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.MulC(10)
}

// Validate validates signatures of the transaction
func (tx *OpenAccount) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(tx.Vin) == 0 {
		return types.ErrInvalidTxInCount
	}
	if len(signers) > 1 {
		return types.ErrInvalidSignerCount
	}
	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}

	for _, vin := range tx.Vin {
		if utxo, err := loader.UTXO(vin.ID()); err != nil {
			return err
		} else {
			if utxo.PublicHash != signers[0] {
				return types.ErrInvalidUTXOSigner
			}
		}
	}

	for _, vout := range tx.Vout {
		if vout.Amount.Less(amount.COIN.DivC(10)) {
			return types.ErrDustAmount
		}
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *OpenAccount) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}

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

	outsum := tx.Fee(ctw)
	for n, vout := range tx.Vout {
		if vout.Amount.Less(amount.COIN.DivC(10)) {
			return types.ErrDustAmount
		}
		outsum = outsum.Add(vout.Amount)
		if err := ctw.CreateUTXO(types.MarshalID(ctw.TargetHeight(), index, uint16(n)), vout); err != nil {
			return err
		}
	}

	if !insum.Equal(outsum) {
		return types.ErrInvalidOutputAmount
	}

	addr := common.NewAddress(ctw.TargetHeight(), index, 0)
	if is, err := ctw.HasAccount(addr); err != nil {
		return err
	} else if is {
		return types.ErrExistAddress
	} else if isn, err := ctw.HasAccountName(tx.Name); err != nil {
		return err
	} else if isn {
		return types.ErrExistAccountName
	} else {
		acc := &SingleAccount{
			Address_: addr,
			Name_:    tx.Name,
			KeyHash:  tx.KeyHash,
		}
		ctw.CreateAccount(acc)
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *OpenAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(tx.Name); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
