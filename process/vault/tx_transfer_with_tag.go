package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// TransferWithTag is a TransferWithTag
type TransferWithTag struct {
	Timestamp_ uint64
	From_      common.Address
	To         common.Address
	Amount     *amount.Amount
	Tag        string
}

// Timestamp returns the timestamp of the transaction
func (tx *TransferWithTag) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *TransferWithTag) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *TransferWithTag) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Vault)
	return sp.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *TransferWithTag) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Vault)

	if len(tx.Tag) > 32 {
		return ErrInvalidTagSize
	}
	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}

	if has, err := loader.HasAccount(tx.To); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	if err := sp.CheckFeePayableWith(p, loader, tx, tx.Amount); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *TransferWithTag) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	return sp.WithFee(p, ctw, tx, func() error {
		if err := sp.SubBalance(ctw, tx.From(), tx.Amount); err != nil {
			return err
		}
		if err := sp.AddBalance(ctw, tx.To, tx.Amount); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *TransferWithTag) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"to":`)
	if bs, err := tx.To.MarshalJSON(); err != nil {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"tag":`)
	if bs, err := json.Marshal(tx.Tag); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
