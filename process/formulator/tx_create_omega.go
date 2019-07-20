package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// CreateOmega is used to make omega formulator account
type CreateOmega struct {
	Timestamp_       uint64
	Seq_             uint64
	From_            common.Address
	SigmaFormulators []common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateOmega) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateOmega) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateOmega) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CreateOmega) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.From() != tx.SigmaFormulators[0] {
		return ErrInvalidFormulatorAddress
	}
	if tx.Seq() <= loader.Seq(tx.SigmaFormulators[0]) {
		return types.ErrInvalidSequence
	}
	if len(tx.SigmaFormulators) != len(signers) {
		return types.ErrInvalidSignerCount
	}

	for i, From := range tx.SigmaFormulators {
		acc, err := loader.Account(From)
		if err != nil {
			return err
		}
		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return types.ErrInvalidAccountType
		}
		if frAcc.FormulatorType != SigmaFormulatorType {
			return types.ErrInvalidAccountType
		}
		if err := frAcc.Validate(loader, []common.PublicHash{signers[i]}); err != nil {
			return err
		}
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateOmega) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	policy := &OmegaPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagOmegaPolicy), &policy); err != nil {
		return err
	}
	if len(tx.SigmaFormulators) != int(policy.OmegaRequiredSigmaCount) {
		return ErrInvalidFormulatorCount
	}

	for _, addr := range tx.SigmaFormulators[1:] {
		if acc, err := ctw.Account(addr); err != nil {
			return err
		} else {
			subAcc := acc.(*FormulatorAccount)
			if ctw.TargetHeight() < subAcc.UpdatedHeight+policy.OmegaRequiredSigmaBlocks {
				return ErrInsufficientFormulatorBlocks
			}
			if err := sp.vault.AddBalance(ctw, tx.SigmaFormulators[0], subAcc.Amount); err != nil {
				return err
			}
			if err := ctw.DeleteAccount(subAcc); err != nil {
				return err
			}
		}
	}

	acc, err := ctw.Account(tx.SigmaFormulators[0])
	if err != nil {
		return err
	}
	frAcc := acc.(*FormulatorAccount)
	frAcc.FormulatorType = OmegaFormulatorType
	frAcc.UpdatedHeight = ctw.TargetHeight()
	return nil
}

// MarshalJSON is a marshaler function
func (tx *CreateOmega) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"sigma_formulators":[`)
	for i, addr := range tx.SigmaFormulators {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := addr.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
