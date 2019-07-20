package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// Transfer is a Transfer
type Transfer struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	To         common.Address
	Amount     *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Transfer) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Transfer) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Transfer) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Transfer) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if is, err := loader.HasAccount(tx.To); err != nil {
		return err
	} else if !is {
		return types.ErrNotExistAccount
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	if has, err := loader.HasAccount(tx.To); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Transfer) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	if err := sp.SubBalance(ctw, tx.From(), amount.COIN.DivC(10)); err != nil {
		return err
	}
	if err := sp.SubBalance(ctw, tx.From(), tx.Amount); err != nil {
		return err
	}
	if err := sp.AddBalance(ctw, tx.To, tx.Amount); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Transfer) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"to":`)
	if bs, err := tx.To.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
