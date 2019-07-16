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
	Seq_            uint64
	From_           common.Address
	HyperFormulator common.Address
	Amount          *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Unstaking) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Unstaking) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Unstaking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *Unstaking) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *Unstaking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

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
	return nil
}

// Execute updates the context by the transaction
func (tx *Unstaking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	if tx.Seq() != ctw.Seq(tx.From())+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From())

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return ErrInvalidStakingAmount
	}

	fromAcc, err := ctw.Account(tx.From())
	if err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, fromAcc.Address(), tx.Fee(ctw)); err != nil {
		return err
	}

	acc, err := ctw.Account(tx.HyperFormulator)
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

	fromStakingAmount := sp.GetStakingAmount(ctw, tx.HyperFormulator, tx.From())
	if fromStakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}
	fromStakingAmount = fromStakingAmount.Sub(tx.Amount)
	if fromStakingAmount.IsZero() {
		sp.removeStakingAmount(ctw, tx.HyperFormulator, tx.From())
		sp.setUserAutoStaking(ctw, tx.HyperFormulator, tx.From(), false)
	} else {
		sp.setStakingAmount(ctw, tx.HyperFormulator, tx.From(), fromStakingAmount)
	}
	if frAcc.StakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}
	frAcc.StakingAmount = frAcc.StakingAmount.Sub(tx.Amount)

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return err
	}
	if err := sp.vault.AddLockedBalance(ctw, fromAcc.Address(), ctw.TargetHeight()+policy.StakingUnlockRequiredBlocks, tx.Amount); err != nil {
		return err
	}
	return nil
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
	buffer.WriteString(`"Hyper_formulator":`)
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
