package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// ChangeOwner is used to remove formulator account and get back staked coin
type ChangeOwner struct {
	Timestamp_ uint64
	From_      common.Address
	KeyHash    common.PublicHash
	GenHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *ChangeOwner) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *ChangeOwner) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *ChangeOwner) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *ChangeOwner) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if err := frAcc.Validate(loader, signers); err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayable(p, loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *ChangeOwner) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.From())
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		frAcc.KeyHash = tx.KeyHash
		frAcc.GenHash = tx.GenHash
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *ChangeOwner) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := tx.GenHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
