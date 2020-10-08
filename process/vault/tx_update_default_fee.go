package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// UpdateDefaultFee is used to update vault policy
type UpdateDefaultFee struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	DefaultFee *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateDefaultFee) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *UpdateDefaultFee) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *UpdateDefaultFee) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *UpdateDefaultFee) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Vault)

	if tx.DefaultFee == nil {
		return ErrInvalidDefaultFee
	}
	if tx.DefaultFee.Less(amount.COIN.DivC(100000000)) {
		if !tx.DefaultFee.IsZero() {
			return ErrInvalidDefaultFee
		}
	}
	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *UpdateDefaultFee) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	ctw.SetProcessData(tagDefaultFee, tx.DefaultFee.Bytes())
	if tx.DefaultFee.IsZero() {
		ctw.SetProcessData(tagDefaultFeeIsZero, []byte{1})
	} else {
		ctw.SetProcessData(tagDefaultFeeIsZero, []byte{})
	}

	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateDefaultFee) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"default_fee":`)
	if bs, err := tx.DefaultFee.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
