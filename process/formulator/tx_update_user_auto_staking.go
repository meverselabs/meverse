package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// UpdateUserAutoStaking is used to update autostaking setup of the account at the hyper formulator
type UpdateUserAutoStaking struct {
	Timestamp_      uint64
	Seq_            uint64
	From_           common.Address
	HyperFormulator common.Address
	AutoStaking     bool
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateUserAutoStaking) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *UpdateUserAutoStaking) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *UpdateUserAutoStaking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *UpdateUserAutoStaking) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *UpdateUserAutoStaking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
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
	return nil
}

// Execute updates the context by the transaction
func (tx *UpdateUserAutoStaking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	if tx.Seq() != ctw.Seq(tx.From())+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From())

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
	frAcc := acc.(*FormulatorAccount)
	sp.setUserAutoStaking(ctw, frAcc.Address(), tx.From(), tx.AutoStaking)
	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateUserAutoStaking) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"auto_staking":`)
	if bs, err := json.Marshal(tx.AutoStaking); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
