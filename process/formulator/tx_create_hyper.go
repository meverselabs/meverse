package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
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

// Fee returns the fee of the transaction
func (tx *CreateHyper) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateHyper) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}

	if tx.Policy == nil {
		return ErrInvalidPolicy
	}
	if tx.Policy.CommissionRatio1000 > 1000 {
		return ErrInvalidPolicy
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if tx.From() != sp.adminAddress {
		return ErrUnauthorizedTransaction
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

	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}

	if tx.Policy == nil {
		return ErrInvalidPolicy
	}
	if tx.Policy.CommissionRatio1000 > 1000 {
		return ErrInvalidPolicy
	}

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return err
	}

	if tx.Seq() != ctw.Seq(tx.From())+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From())

	if has, err := ctw.HasAccount(tx.From()); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}
	if err := sp.vault.SubBalance(ctw, tx.From(), tx.Fee(ctw)); err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, tx.From(), policy.HyperCreationAmount); err != nil {
		return err
	}

	addr := common.NewAddress(ctw.TargetHeight(), index, 0)
	if is, err := ctw.HasAccount(addr); err != nil {
		return err
	} else if is {
		return types.ErrExistAddress
	} else if isn, err := ctw.HasAccountName(tx.Name); err != nil {
		return err
	} else if isn {
		return types.ErrExistAccountName
	} else {
		acc := &FormulatorAccount{
			Address_:       addr,
			Name_:          tx.Name,
			FormulatorType: HyperFormulatorType,
			KeyHash:        tx.KeyHash,
			GenHash:        tx.GenHash,
			Amount:         policy.HyperCreationAmount,
			UpdatedHeight:  ctw.TargetHeight(),
			StakingAmount:  amount.NewCoinAmount(0, 0),
			Policy:         tx.Policy,
		}
		ctw.CreateAccount(acc)
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
