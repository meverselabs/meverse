package gateway

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// TokenOut is a TokenOut
type TokenOut struct {
	Timestamp_ uint64
	From_      common.Address
	ERC20To    ERC20Address
	Amount     *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *TokenOut) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *TokenOut) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *TokenOut) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Gateway)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *TokenOut) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Gateway)

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}

	AdminAddress := sp.admin.AdminAddress(loader, p.Name())
	if has, err := loader.HasAccount(AdminAddress); err != nil {
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

	policy := &Policy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagPolicy), &policy); err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayableWith(p, loader, tx, tx.Amount.Add(policy.WithdrawFee)); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *TokenOut) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Gateway)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		policy := &Policy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagPolicy), &policy); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, tx.From(), policy.WithdrawFee); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, tx.From(), tx.Amount); err != nil {
			return err
		}
		AdminAddress := sp.admin.AdminAddress(ctw, p.Name())
		if err := sp.vault.AddBalance(ctw, AdminAddress, policy.WithdrawFee); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, AdminAddress, tx.Amount); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *TokenOut) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"erc20_to":`)
	if bs, err := tx.ERC20To.MarshalJSON(); err != nil {
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
