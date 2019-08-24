package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
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

// Fee returns the fee of the transaction
func (tx *CreateOmega) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateOmega) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.From() != tx.SigmaFormulators[0] {
		return ErrInvalidFormulatorAddress
	}
	if tx.Seq() <= loader.Seq(tx.SigmaFormulators[0]) {
		return types.ErrInvalidSequence
	}
	if len(tx.SigmaFormulators) != len(signers) {
		return types.ErrInvalidSignerCount
	}

	policy := &OmegaPolicy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagOmegaPolicy), &policy); err != nil {
		return err
	}
	if len(tx.SigmaFormulators) != int(policy.OmegaRequiredSigmaCount) {
		return ErrInvalidFormulatorCount
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
		if loader.TargetHeight()+frAcc.PreHeight < frAcc.UpdatedHeight+policy.OmegaRequiredSigmaBlocks {
			return ErrInsufficientFormulatorBlocks
		}
		if err := frAcc.Validate(loader, []common.PublicHash{signers[i]}); err != nil {
			return err
		}
	}

	if err := sp.vault.CheckFeePayable(loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateOmega) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(ctw, tx, func() error {
		policy := &OmegaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagOmegaPolicy), &policy); err != nil {
			return err
		}
		acc, err := ctw.Account(tx.SigmaFormulators[0])
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		for _, addr := range tx.SigmaFormulators[1:] {
			if acc, err := ctw.Account(addr); err != nil {
				return err
			} else {
				subAcc := acc.(*FormulatorAccount)
				if err := sp.vault.AddBalance(ctw, tx.SigmaFormulators[0], sp.vault.Balance(ctw, addr)); err != nil {
					return err
				}
				if err := sp.vault.RemoveBalance(ctw, addr); err != nil {
					return err
				}
				frAcc.Amount = frAcc.Amount.Add(subAcc.Amount)
				if err := ctw.DeleteAccount(subAcc); err != nil {
					return err
				}
			}
		}

		frAcc.FormulatorType = OmegaFormulatorType
		frAcc.PreHeight = 0
		frAcc.UpdatedHeight = ctw.TargetHeight()
		return nil
	})
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
