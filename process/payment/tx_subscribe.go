package payment

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// Subscribe is a Subscribe
type Subscribe struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Topic      uint64
	Amount     *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Subscribe) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Subscribe) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Subscribe) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Subscribe) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Payment)

	if !tx.Amount.IsZero() && tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if _, err := sp.GetTopicName(loader, tx.Topic); err != nil {
		return err
	}

	if len(signers) <= 1 {
		return admin.ErrUnauthorizedTransaction
	}

	adminAddr := sp.admin.AdminAddress(loader, p.Name())
	adminAcc, err := loader.Account(adminAddr)
	if err != nil {
		return err
	}
	if err := adminAcc.Validate(loader, signers[:1]); err != nil {
		return err
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers[1:]); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Subscribe) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Payment)

	if err := sp.addSubscribe(ctw, tx.Topic, tx.From(), tx.Amount); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Subscribe) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
