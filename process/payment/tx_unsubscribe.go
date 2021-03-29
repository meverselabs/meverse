package payment

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// Unsubscribe is a Unsubscribe
type Unsubscribe struct {
	Timestamp_ uint64
	From_      common.Address
	Topic      uint64
}

// Timestamp returns the timestamp of the transaction
func (tx *Unsubscribe) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *Unsubscribe) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Unsubscribe) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Payment)

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
func (tx *Unsubscribe) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Payment)

	sp.removeSubscribe(ctw, tx.Topic, tx.From())
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Unsubscribe) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
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
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
