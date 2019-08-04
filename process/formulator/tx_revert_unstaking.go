package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// RevertUnstaking is used to ustake coin from the hyper formulator
type RevertUnstaking struct {
	Timestamp_      uint64
	Seq_            uint64
	From_           common.Address
	HyperFormulator common.Address
	UnstakedHeight  uint32
	Amount          *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *RevertUnstaking) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *RevertUnstaking) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *RevertUnstaking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *RevertUnstaking) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *RevertUnstaking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.Amount.Less(amount.COIN) {
		return ErrInvalidStakingAmount
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.HyperFormulator)
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if frAcc.FormulatorType != HyperFormulatorType {
		return types.ErrInvalidAccountType
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	am, err := sp.getUnstakingAmount(loader, tx.HyperFormulator, tx.From(), tx.UnstakedHeight)
	if err != nil {
		return err
	}
	if am.Less(tx.Amount) {
		return ErrMinustUnstakingAmount
	}

	if err := sp.vault.CheckFeePayable(loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *RevertUnstaking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(ctw, tx, func() error {
		acc, err := ctw.Account(tx.HyperFormulator)
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)

		if err := sp.subUnstakingAmount(ctw, tx.HyperFormulator, tx.From(), tx.UnstakedHeight, tx.Amount); err != nil {
			return err
		}

		sp.AddStakingAmount(ctw, tx.HyperFormulator, tx.From(), tx.Amount)
		frAcc.StakingAmount = frAcc.StakingAmount.Add(tx.Amount)
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *RevertUnstaking) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"hyper_formulator":`)
	if bs, err := tx.HyperFormulator.MarshalJSON(); err != nil {
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
