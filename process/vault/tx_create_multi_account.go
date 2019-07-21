package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// CreateMultiAccount is used to make multi account
type CreateMultiAccount struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Name       string
	Requried   uint8
	KeyHashes  []common.PublicHash
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateMultiAccount) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateMultiAccount) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateMultiAccount) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *CreateMultiAccount) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateMultiAccount) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Vault)

	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}
	if tx.Requried < 1 {
		return ErrInvalidRequiredKeyHashCount
	}
	if len(tx.KeyHashes) <= 1 || len(tx.KeyHashes) > 10 {
		return ErrInvalidMultiKeyHashCount
	}
	keyHashMap := map[common.PublicHash]bool{}
	for _, v := range tx.KeyHashes {
		keyHashMap[v] = true
	}
	if len(keyHashMap) != len(tx.KeyHashes) {
		return ErrInvalidMultiKeyHashCount
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	CreationCost := amount.COIN.MulC(10)
	if err := sp.CheckFeePayableWith(loader, tx, CreationCost); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateMultiAccount) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	return sp.WithFee(ctw, tx, func() error {
		CreationCost := amount.COIN.MulC(10)
		if err := sp.SubBalance(ctw, tx.From(), CreationCost); err != nil {
			return err
		}
		if err := sp.AddCollectedFee(ctw, CreationCost); err != nil {
			return err
		}

		acc := &MultiAccount{
			Address_:  common.NewAddress(ctw.TargetHeight(), index, 0),
			Name_:     tx.Name,
			Required:  tx.Requried,
			KeyHashes: tx.KeyHashes,
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *CreateMultiAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"key_hashes":`)
	buffer.WriteString(`[`)
	for i, pubhash := range tx.KeyHashes {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := pubhash.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
