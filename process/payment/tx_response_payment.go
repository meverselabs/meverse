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

// ResponsePayment is a ResponsePayment
type ResponsePayment struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	TXID       string
	Amount     *amount.Amount
	IsAccept   bool
}

// Timestamp returns the timestamp of the transaction
func (tx *ResponsePayment) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *ResponsePayment) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *ResponsePayment) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *ResponsePayment) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Payment)

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	req, err := sp.getRequestPayment(loader, tx.TXID)
	if err != nil {
		return err
	}
	if req.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if req.To != tx.From() {
		return ErrInvalidRequestPayment
	}
	if !req.Amount.Equal(tx.Amount) {
		return ErrInvalidRequestPayment
	}
	if _, err := sp.GetTopicName(loader, req.Topic); err != nil {
		return err
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
func (tx *ResponsePayment) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Payment)

	if tx.IsAccept {
		if err := sp.vault.SubBalance(ctw, tx.From(), tx.Amount); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, sp.admin.AdminAddress(ctw, p.Name()), tx.Amount); err != nil {
			return err
		}
	}
	sp.removeRequestPayment(ctw, tx.TXID)
	return nil
}

// MarshalJSON is a marshaler function
func (tx *ResponsePayment) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"txid":`)
	if bs, err := json.Marshal(tx.TXID); err != nil {
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
	buffer.WriteString(`"is_accept":`)
	if bs, err := json.Marshal(tx.IsAccept); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
