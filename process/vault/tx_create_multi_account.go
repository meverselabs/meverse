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
	return amount.COIN.MulC(10)
}

// Validate validates signatures of the transaction
func (tx *CreateMultiAccount) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}
	if len(tx.KeyHashes) <= 1 {
		return ErrInvalidMultiKeyHashCount
	}
	keyHashMap := map[common.PublicHash]bool{}
	for _, v := range tx.KeyHashes {
		keyHashMap[v] = true
	}
	if len(keyHashMap) != len(tx.KeyHashes) {
		return ErrInvalidMultiKeyHashCount
	}
	if len(tx.KeyHashes) > 10 {
		return ErrInvalidMultiKeyHashCount
	}
	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}
	if tx.Requried < 1 {
		return ErrInvalidRequiredKeyHashCount
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
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateMultiAccount) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Vault)

	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}
	if len(tx.KeyHashes) <= 1 {
		return ErrInvalidMultiKeyHashCount
	}
	keyHashMap := map[common.PublicHash]bool{}
	for _, v := range tx.KeyHashes {
		keyHashMap[v] = true
	}
	if len(keyHashMap) != len(tx.KeyHashes) {
		return ErrInvalidMultiKeyHashCount
	}
	if len(tx.KeyHashes) > 10 {
		return ErrInvalidMultiKeyHashCount
	}
	if len(tx.Name) < 8 || len(tx.Name) > 16 {
		return types.ErrInvalidAccountName
	}
	if tx.Requried < 1 {
		return ErrInvalidRequiredKeyHashCount
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
	if err := sp.SubBalance(ctw, tx.From(), tx.Fee(ctw)); err != nil {
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
		acc := &MultiAccount{
			Address_:  addr,
			Name_:     tx.Name,
			Required:  tx.Requried,
			KeyHashes: tx.KeyHashes,
		}
		ctw.CreateAccount(acc)
	}
	return nil
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
