package pof

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// FormulationType is type of formulation account
type FormulationType uint8

// formulator types
const (
	AlphaFormulatorType = FormulationType(1)
	SigmaFormulatorType = FormulationType(2)
	OmegaFormulatorType = FormulationType(3)
	HyperFormulatorType = FormulationType(4)
)

// FormulationAccount is a consensus.FormulationAccount
// It is used to indentify Hyper formulator that supports the staking system
type FormulationAccount struct {
	Address_        common.Address
	Name_           string
	FormulationType FormulationType
	KeyHash         common.PublicHash
	Amount          *amount.Amount
	StakingAmount   *amount.Amount
	Policy          *HyperPolicy
}

// Address returns the address of the account
func (acc *FormulationAccount) Address() common.Address {
	return acc.Address_
}

// Name returns the name of the account
func (acc *FormulationAccount) Name() string {
	return acc.Name_
}

// Clone returns the clonend value of it
func (acc *FormulationAccount) Clone() types.Account {
	c := &FormulationAccount{
		Address_:        acc.Address_,
		Name_:           acc.Name_,
		FormulationType: acc.FormulationType,
		KeyHash:         acc.KeyHash.Clone(),
		Amount:          acc.Amount.Clone(),
	}
	if acc.FormulationType == HyperFormulatorType {
		c.StakingAmount = acc.StakingAmount.Clone()
		c.Policy = acc.Policy.Clone()
	}
	return c
}

// Validate validates account signers
func (acc *FormulationAccount) Validate(loader types.LoaderProcess, signers []common.PublicHash) error {
	if len(signers) != 1 {
		return ErrInvalidSignerCount
	}
	if acc.KeyHash != signers[0] {
		return ErrInvalidAccountSigner
	}
	return nil
}

// MarshalJSON is a marshaler function
func (acc *FormulationAccount) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"address":`)
	if bs, err := acc.Address_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(acc.Name_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"formulation_type":`)
	if bs, err := json.Marshal(acc.FormulationType); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := acc.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	if acc.FormulationType == HyperFormulatorType {
		buffer.WriteString(`,`)
		buffer.WriteString(`"policy":`)
		if bs, err := acc.Policy.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`,`)
		buffer.WriteString(`"amount":`)
		if bs, err := acc.Amount.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`,`)
		buffer.WriteString(`"staking_amount":`)
		if bs, err := acc.StakingAmount.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
