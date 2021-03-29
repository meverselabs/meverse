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
	From_           common.Address
	HyperFormulator common.Address
	Amount          *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Staking) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *Staking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *Staking) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *Staking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return ErrInvalidStakingAmount
	}

	if tx.From() == tx.HyperFormulator {
		return ErrInvalidStakingAddress
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
	if frAcc.IsRevoked {
		return ErrRevokedFormulator
	}
	if !frAcc.Policy.MinimumStaking.IsZero() && tx.Amount.Less(frAcc.Policy.MinimumStaking) {
		return ErrInvalidStakingAmount
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayableWith(p, loader, tx, tx.Amount); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Staking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.HyperFormulator)
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)

		fromAcc, err := ctw.Account(tx.From())
		if err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, fromAcc.Address(), tx.Amount); err != nil {
			return err
		}

		sp.AddStakingAmount(ctw, tx.HyperFormulator, tx.From(), tx.Amount)
		frAcc.StakingAmount = frAcc.StakingAmount.Add(tx.Amount)
		return nil
	})
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
