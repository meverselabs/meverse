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

// Fee returns the fee of the transaction
func (tx *CreateOmega) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateOmega) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.SigmaFormulators[0]) {
		return types.ErrInvalidSequence
	}

	for _, From := range tx.SigmaFormulators {
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
		if err := frAcc.Validate(loader, signers); err != nil {
			return err
		}
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateOmega) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	policy := &OmegaPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData([]byte("OmegaPolicy")), &policy); err != nil {
		return err
	}
	if len(tx.SigmaFormulators) != int(policy.OmegaRequiredSigmaCount) {
		return ErrInvalidFormulatorCount
	}

	sn := ctw.Snapshot()
	defer ctw.Revert(sn)

	if tx.Seq() != ctw.Seq(tx.SigmaFormulators[0])+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.SigmaFormulators[0])

	acc, err := ctw.Account(tx.SigmaFormulators[0])
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	} else if frAcc.FormulatorType != SigmaFormulatorType {
		return types.ErrInvalidAccountType
	} else if ctw.TargetHeight() < frAcc.UpdatedHeight+policy.OmegaRequiredSigmaBlocks {
		return ErrInsufficientFormulatorBlocks
	}

	for _, addr := range tx.SigmaFormulators[1:] {
		if acc, err := ctw.Account(addr); err != nil {
			return err
		} else if subAcc, is := acc.(*FormulatorAccount); !is {
			return types.ErrInvalidAccountType
		} else if subAcc.FormulatorType != SigmaFormulatorType {
			return types.ErrInvalidAccountType
		} else if ctw.TargetHeight() < subAcc.UpdatedHeight+policy.OmegaRequiredSigmaBlocks {
			return ErrInsufficientFormulatorBlocks
		} else {
			sp.vault.AddBalance(ctw, tx.SigmaFormulators[0], subAcc.Amount)
			ctw.DeleteAccount(subAcc)
		}
	}
	frAcc.FormulatorType = OmegaFormulatorType
	frAcc.UpdatedHeight = ctw.TargetHeight()

	ctw.Commit(sn)
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
