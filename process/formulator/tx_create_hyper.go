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
	From       common.Address
	Name       string
	KeyHash    common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateHyper) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateHyper) Seq() uint64 {
	return tx.Seq_
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

	if tx.Seq() <= loader.Seq(tx.From) {
		return types.ErrInvalidSequence
	}

	if tx.From != sp.adminAddress {
		return ErrUnauthorizedTransaction
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
func (tx *CreateHyper) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}

	policy := &HyperPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
		return err
	}

	sn := ctw.Snapshot()
	defer ctw.Revert(sn)

	if tx.Seq() != ctw.Seq(tx.From)+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From)

	if has, err := ctw.HasAccount(tx.From); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}
	if err := sp.vault.SubBalance(ctw, tx.From, tx.Fee(ctw)); err != nil {
		return err
	}
	if err := sp.vault.SubBalance(ctw, tx.From, policy.HyperCreationAmount); err != nil {
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
			Amount:         policy.HyperCreationAmount,
			UpdatedHeight:  ctw.TargetHeight(),
		}
		ctw.CreateAccount(acc)
	}
	ctw.Commit(sn)
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
	if bs, err := tx.From.MarshalJSON(); err != nil {
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
