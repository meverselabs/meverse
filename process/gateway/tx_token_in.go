package gateway

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// TokenIn is a TokenIn
type TokenIn struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	ERC20TXID  hash.Hash256
	ERC20From  ERC20Address
	To         common.Address
	Amount     *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *TokenIn) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *TokenIn) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *TokenIn) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *TokenIn) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Gateway)

	if tx.From() != sp.admin.AdminAddress() {
		return admin.ErrUnauthorizedTransaction
	}
	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if sp.HasERC20TXID(loader, tx.ERC20TXID) {
		return ErrProcessedERC20TXID
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
	return nil
}

// Execute updates the context by the transaction
func (tx *TokenIn) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Gateway)

	if err := sp.vault.SubBalance(ctw, tx.From(), tx.Amount); err != nil {
		return err
	}
	if err := sp.vault.AddBalance(ctw, tx.To, tx.Amount); err != nil {
		return err
	}

	sp.setERC20TXID(ctw, tx.ERC20TXID)

	return nil
}

// MarshalJSON is a marshaler function
func (tx *TokenIn) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"erc20_txid":`)
	if bs, err := tx.ERC20TXID.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"erc20_from":`)
	if bs, err := tx.ERC20From.MarshalJSON(); err != nil {
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
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
