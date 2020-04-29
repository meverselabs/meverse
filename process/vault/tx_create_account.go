package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/encoding"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// CreateAccount is used to make a account
type CreateAccount struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Name       string
	KeyHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateAccount) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateAccount) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateAccount) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *CreateAccount) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Vault)
	return sp.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *CreateAccount) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Vault)

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

	policy := &Policy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagPolicy), &policy); err != nil {
		return err
	}

	if err := sp.CheckFeePayableWith(p, loader, tx, policy.AccountCreationAmount); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateAccount) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	return sp.WithFee(p, ctw, tx, func() error {
		policy := &Policy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagPolicy), &policy); err != nil {
			return err
		}

		if err := sp.SubBalance(ctw, tx.From(), policy.AccountCreationAmount); err != nil {
			return err
		}
		if err := sp.AddCollectedFee(ctw, policy.AccountCreationAmount); err != nil {
			return err
		}

		acc := &SingleAccount{
			Address_: sp.cn.NewAddress(ctw.TargetHeight(), index),
			Name_:    tx.Name,
			KeyHash:  tx.KeyHash,
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *CreateAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
