package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// RevertRevoke is used to revert operation that to remove formulator account and get back staked coin
type RevertRevoke struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *RevertRevoke) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *RevertRevoke) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *RevertRevoke) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *RevertRevoke) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *RevertRevoke) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if !frAcc.IsRevoked {
		return ErrNotRevoked
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
func (tx *RevertRevoke) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.From())
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		frAcc.IsRevoked = false
		if err := sp.removeRevokedFormulator(ctw, acc.Address()); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *RevertRevoke) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
