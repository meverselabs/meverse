package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Revoke is used to remove formulator account and get back staked coin
type Revoke struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Heritor    common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *Revoke) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Revoke) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Revoke) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *Revoke) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *Revoke) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.From() == tx.Heritor {
		return ErrInvalidHeritor
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if has, err := loader.HasAccount(tx.Heritor); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if frAcc.IsRevoked {
		return ErrRevokedFormulator
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
func (tx *Revoke) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.From())
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		frAcc.IsRevoked = true

		switch frAcc.FormulatorType {
		case AlphaFormulatorType:
			policy := &AlphaPolicy{}
			if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &policy); err != nil {
				return err
			}
			if err := sp.addRevokedFormulator(ctw, acc.Address(), ctw.TargetHeight()+policy.AlphaUnlockRequiredBlocks, tx.Heritor); err != nil {
				return err
			}
		case SigmaFormulatorType:
			policy := &SigmaPolicy{}
			if err := encoding.Unmarshal(ctw.ProcessData(tagSigmaPolicy), &policy); err != nil {
				return err
			}
			if err := sp.addRevokedFormulator(ctw, acc.Address(), ctw.TargetHeight()+policy.SigmaUnlockRequiredBlocks, tx.Heritor); err != nil {
				return err
			}
		case OmegaFormulatorType:
			policy := &OmegaPolicy{}
			if err := encoding.Unmarshal(ctw.ProcessData(tagOmegaPolicy), &policy); err != nil {
				return err
			}
			if err := sp.addRevokedFormulator(ctw, acc.Address(), ctw.TargetHeight()+policy.OmegaUnlockRequiredBlocks, tx.Heritor); err != nil {
				return err
			}
		case HyperFormulatorType:
			policy := &HyperPolicy{}
			if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
				return err
			}
			if err := sp.addRevokedFormulator(ctw, acc.Address(), ctw.TargetHeight()+policy.HyperUnlockRequiredBlocks, tx.Heritor); err != nil {
				return err
			}
		default:
			return types.ErrInvalidAccountType
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *Revoke) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"heritor":`)
	if bs, err := tx.Heritor.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
