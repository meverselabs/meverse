package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/admin"
)

// CreateHyper is used to make hyper formulator account
type CreateHyper struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Name       string
	KeyHash    common.PublicHash
	GenHash    common.PublicHash
	Policy     *ValidatorPolicy
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateHyper) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateHyper) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateHyper) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CreateHyper) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}

	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}

	if tx.Policy == nil {
		return ErrInvalidValidatorPolicy
	}
	if tx.Policy.CommissionRatio1000 > 1000 {
		return ErrInvalidValidatorPolicy
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
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateHyper) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, tx.From(), policy.HyperCreationAmount); err != nil {
		return err
	}

	acc := &FormulatorAccount{
		Address_:       common.NewAddress(ctw.TargetHeight(), index, 0),
		Name_:          tx.Name,
		FormulatorType: HyperFormulatorType,
		KeyHash:        tx.KeyHash,
		GenHash:        tx.GenHash,
		Amount:         policy.HyperCreationAmount,
		UpdatedHeight:  ctw.TargetHeight(),
		StakingAmount:  amount.NewCoinAmount(0, 0),
		Policy:         tx.Policy,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *CreateHyper) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"policy":`)
	if bs, err := tx.Policy.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
