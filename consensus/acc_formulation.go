package consensus

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/account"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/data"
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

func init() {
	data.RegisterAccount("consensus.FormulationAccount", func(t account.Type) account.Account {
		return &FormulationAccount{
			Base: account.Base{
				Type_:    t,
				Balance_: amount.NewCoinAmount(0, 0),
			},
			Amount: amount.NewCoinAmount(0, 0),
			Policy: &HyperPolicy{
				MinimumStaking: amount.NewCoinAmount(0, 0),
				MaximumStaking: amount.NewCoinAmount(0, 0),
			},
			StakingAmount: amount.NewCoinAmount(0, 0),
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*FormulationAccount)
		if len(signers) != 1 {
			return ErrInvalidSignerCount
		}
		signer := signers[0]
		if !acc.KeyHash.Equal(signer) {
			return ErrInvalidAccountSigner
		}
		return nil
	})
}

// FormulationAccount is a consensus.FormulationAccount
// It is used to indentify Hyper formulator that supports the staking system
type FormulationAccount struct {
	account.Base
	FormulationType FormulationType
	KeyHash         common.PublicHash
	Amount          *amount.Amount
	Policy          *HyperPolicy
	StakingAmount   *amount.Amount
}

// Clone returns the clonend value of it
func (acc *FormulationAccount) Clone() account.Account {
	return &FormulationAccount{
		Base: account.Base{
			Type_:    acc.Type_,
			Address_: acc.Address_,
			Balance_: acc.Balance(),
		},
		FormulationType: acc.FormulationType,
		KeyHash:         acc.KeyHash.Clone(),
		Policy:          acc.Policy.Clone(),
		Amount:          acc.Amount.Clone(),
		StakingAmount:   acc.StakingAmount.Clone(),
	}
}

// WriteTo is a serialization function
func (acc *FormulationAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(acc.FormulationType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.Policy.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.StakingAmount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (acc *FormulationAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		acc.FormulationType = FormulationType(v)
	}
	if n, err := acc.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.Policy.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.StakingAmount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
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
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(acc.Type_); err != nil {
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
		if bs, err := acc.StakingAmount.MarshalJSON(); err != nil {
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
