package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// ChangeSingleKey is used to remove formulator account and get back staked coin
type ChangeSingleKey struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	KeyHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *ChangeSingleKey) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *ChangeSingleKey) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *ChangeSingleKey) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *ChangeSingleKey) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Vault)
	return sp.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *ChangeSingleKey) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Vault)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	singleAcc, is := acc.(*SingleAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if err := singleAcc.Validate(loader, signers); err != nil {
		return err
	}

	if err := sp.CheckFeePayable(p, loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *ChangeSingleKey) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	return sp.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.From())
		if err != nil {
			return err
		}
		frAcc := acc.(*SingleAccount)
		frAcc.KeyHash = tx.KeyHash
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *ChangeSingleKey) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
