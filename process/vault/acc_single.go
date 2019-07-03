package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// SingleAccount is a basic account
type SingleAccount struct {
	Address_ common.Address
	Name_    string
	KeyHash  common.PublicHash
}

// Address returns the address of the account
func (acc *SingleAccount) Address() common.Address {
	return acc.Address_
}

// Name returns the name of the account
func (acc *SingleAccount) Name() string {
	return acc.Name_
}

// Clone returns the clonend value of it
func (acc *SingleAccount) Clone() types.Account {
	c := &SingleAccount{
		Address_: acc.Address_,
		Name_:    acc.Name_,
		KeyHash:  acc.KeyHash.Clone(),
	}
	return c
}

// Validate validates account signers
func (acc *SingleAccount) Validate(loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(signers) != 1 {
		return types.ErrInvalidSignerCount
	}
	if acc.KeyHash != signers[0] {
		return types.ErrInvalidAccountSigner
	}
	return nil
}

// MarshalJSON is a marshaler function
func (acc *SingleAccount) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"address":`)
	if bs, err := acc.Address_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(acc.Name_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := acc.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
