package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// Staking is used to stake coin to the hyper formulator
type Staking struct {
	Timestamp_      uint64
	Seq_            uint64
	From            common.Address
	HyperFormulator common.Address
	Amount          *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Staking) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Staking) Seq() uint64 {
	return tx.Seq_
}

// Fee returns the fee of the transaction
func (tx *Staking) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *Staking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.From) {
		return types.ErrInvalidSequence
	}

	if tx.Amount.Less(amount.COIN.DivC(10)) {
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
	if !frAcc.Policy.MinimumStaking.IsZero() && tx.Amount.Less(frAcc.Policy.MinimumStaking) {
		return ErrInvalidStakingAmount
	}

	fromAcc, err := loader.Account(tx.From)
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Staking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	sn := ctw.Snapshot()
	defer ctw.Revert(sn)

	if tx.Seq() != ctw.Seq(tx.From)+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From)

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return ErrInvalidStakingAmount
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
	if !frAcc.Policy.MinimumStaking.IsZero() && tx.Amount.Less(frAcc.Policy.MinimumStaking) {
		return ErrInsufficientStakingAmount
	}

	fromAcc, err := ctw.Account(tx.From)
	if err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, fromAcc.Address(), tx.Fee(ctw)); err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, fromAcc.Address(), tx.Amount); err != nil {
		return err
	}

	sp.addStakingAmount(ctw, tx.HyperFormulator, tx.From, tx.Amount)
	frAcc.StakingAmount = frAcc.StakingAmount.Add(tx.Amount)

	ctw.Commit(sn)
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Staking) MarshalJSON() ([]byte, error) {
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
	if bs, err := tx.From.MarshalJSON(); err != nil {
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
