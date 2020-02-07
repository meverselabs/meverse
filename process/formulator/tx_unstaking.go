package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Unstaking is used to ustake coin from the hyper formulator
type Unstaking struct {
	Timestamp_      uint64
	From_           common.Address
	HyperFormulator common.Address
	Amount          *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Unstaking) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *Unstaking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *Unstaking) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *Unstaking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.Amount.Less(amount.COIN) {
		return ErrInvalidStakingAmount
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

	fromStakingAmount := sp.GetStakingAmount(loader, tx.HyperFormulator, tx.From())
	if fromStakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}
	if frAcc.StakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Unstaking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	acc, err := ctw.Account(tx.HyperFormulator)
	if err != nil {
		return err
	}
	frAcc := acc.(*FormulatorAccount)
	if err := sp.subStakingAmount(ctw, tx.HyperFormulator, tx.From(), tx.Amount); err != nil {
		return err
	}
	frAcc.StakingAmount = frAcc.StakingAmount.Sub(tx.Amount)

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayable(p, ctw, tx); err != nil {
		Fee := tx.Fee(p, ctw)
		if err := sp.vault.AddCollectedFee(ctw, Fee); err != nil {
			return err
		}
		if err := sp.addUnstakingAmount(ctw, tx.HyperFormulator, tx.From(), ctw.TargetHeight()+policy.StakingUnlockRequiredBlocks, tx.Amount.Sub(Fee)); err != nil {
			return err
		}
		return nil
	} else {
		return sp.vault.WithFee(p, ctw, tx, func() error {
			if err := sp.addUnstakingAmount(ctw, tx.HyperFormulator, tx.From(), ctw.TargetHeight()+policy.StakingUnlockRequiredBlocks, tx.Amount); err != nil {
				return err
			}
			return nil
		})
	}
}

// MarshalJSON is a marshaler function
func (tx *Unstaking) MarshalJSON() ([]byte, error) {
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
