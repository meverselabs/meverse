package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// CreateAlpha is used to make alpha formulator account
type CreateAlpha struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Name       string
	KeyHash    common.PublicHash
	GenHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateAlpha) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateAlpha) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateAlpha) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *CreateAlpha) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateAlpha) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if has, err := loader.HasAccountName(tx.Name); err != nil {
		return err
	} else if has {
		return types.ErrExistAccountName
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	policy := &AlphaPolicy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagAlphaPolicy), &policy); err != nil {
		return err
	}
	if loader.TargetHeight() < policy.AlphaCreationLimitHeight {
		return ErrAlphaCreationLimited
	}
	if err := sp.vault.CheckFeePayableWith(loader, tx, policy.AlphaCreationAmount); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateAlpha) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(ctw, tx, func() error {
		policy := &AlphaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &policy); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, tx.From(), policy.AlphaCreationAmount); err != nil {
			return err
		}

		acc := &FormulatorAccount{
			Address_:       common.NewAddress(ctw.TargetHeight(), index, 0),
			Name_:          tx.Name,
			FormulatorType: AlphaFormulatorType,
			KeyHash:        tx.KeyHash,
			GenHash:        tx.GenHash,
			Amount:         policy.AlphaCreationAmount,
			UpdatedHeight:  ctw.TargetHeight(),
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *CreateAlpha) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(tx.Name); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := tx.GenHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
