package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// FormulatorType is type of formulator account
type FormulatorType uint8

// formulator types
const (
	AlphaFormulatorType = FormulatorType(1)
	SigmaFormulatorType = FormulatorType(2)
	OmegaFormulatorType = FormulatorType(3)
	HyperFormulatorType = FormulatorType(4)
)

// FormulatorAccount is a FormulatorAccount
// It is used to indentify Hyper formulator that supports the staking system
type FormulatorAccount struct {
	Address_       common.Address
	Name_          string
	FormulatorType FormulatorType
	KeyHash        common.PublicHash
	GenHash        common.PublicHash
	Amount         *amount.Amount
	IsRevoked      bool
	PreHeight      uint32
	UpdatedHeight  uint32
	StakingAmount  *amount.Amount
	Policy         *ValidatorPolicy
}

// Address returns the address of the account
func (acc *FormulatorAccount) Address() common.Address {
	return acc.Address_
}

// Name returns the name of the account
func (acc *FormulatorAccount) Name() string {
	return acc.Name_
}

// IsFormulator returns it is formulator or not
func (acc *FormulatorAccount) IsFormulator() bool {
	return true
}

// GeneratorHash returns a generator public hash
func (acc *FormulatorAccount) GeneratorHash() common.PublicHash {
	return acc.GenHash
}

// IsActivated returns it is activated or not
func (acc *FormulatorAccount) IsActivated() bool {
	return !acc.IsRevoked
}

// Clone returns the clonend value of it
func (acc *FormulatorAccount) Clone() types.Account {
	c := &FormulatorAccount{
		Address_:       acc.Address_,
		Name_:          acc.Name_,
		FormulatorType: acc.FormulatorType,
		KeyHash:        acc.KeyHash.Clone(),
		GenHash:        acc.GenHash.Clone(),
		Amount:         acc.Amount.Clone(),
		IsRevoked:      acc.IsRevoked,
		PreHeight:      acc.PreHeight,
		UpdatedHeight:  acc.UpdatedHeight,
	}
	if acc.FormulatorType == HyperFormulatorType {
		c.StakingAmount = acc.StakingAmount.Clone()
		c.Policy = acc.Policy.Clone()
	}
	return c
}

// Validate validates account signers
func (acc *FormulatorAccount) Validate(loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(signers) != 1 {
		return types.ErrInvalidSignerCount
	}
	if acc.KeyHash != signers[0] {
		return types.ErrInvalidAccountSigner
	}
	return nil
}

// MarshalJSON is a marshaler function
func (acc *FormulatorAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"formulator_type":`)
	if bs, err := json.Marshal(acc.FormulatorType); err != nil {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := acc.GenHash.MarshalJSON(); err != nil {
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
	buffer.WriteString(`"is_revoked":`)
	if bs, err := json.Marshal(acc.IsRevoked); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"pre_height":`)
	if bs, err := json.Marshal(acc.PreHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"updated_height":`)
	if bs, err := json.Marshal(acc.UpdatedHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	if acc.FormulatorType == HyperFormulatorType {
		buffer.WriteString(`,`)
		buffer.WriteString(`"policy":`)
		if bs, err := acc.Policy.MarshalJSON(); err != nil {
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
