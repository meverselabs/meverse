package payment

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/vault"
)

// Billing is a Billing
type Billing struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Topic      uint64
	To         common.Address
	Amount     *amount.Amount
	Content    string
}

// Timestamp returns the timestamp of the transaction
func (tx *Billing) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Billing) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Billing) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Billing) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Payment)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if len(tx.Content) > 255 {
		return ErrExceedContentSize
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if _, err := sp.GetTopicName(loader, tx.Topic); err != nil {
		return err
	}

	am, err := sp.getSubscribe(loader, tx.Topic, tx.To)
	if err != nil {
		return err
	}
	if !am.IsZero() && !am.Equal(tx.Amount) {
		return ErrInvalidRequestPayment
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	b := sp.vault.Balance(loader, tx.From())
	if b.Less(tx.Amount) {
		return vault.ErrInsufficientBalance
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Billing) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Payment)

	if err := sp.vault.SubBalance(ctw, tx.To, tx.Amount); err != nil {
		return err
	}
	if err := sp.vault.AddBalance(ctw, sp.admin.AdminAddress(ctw, p.Name()), tx.Amount); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Billing) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"topic":`)
	if bs, err := json.Marshal(tx.Topic); err != nil {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"content":`)
	if bs, err := json.Marshal(tx.Content); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
