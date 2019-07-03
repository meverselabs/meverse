package vault

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// MultiAccount is a basic account
type MultiAccount struct {
	Address_  common.Address
	Name_     string
	Required  uint8
	KeyHashes []common.PublicHash
}

// Address returns the address of the account
func (acc *MultiAccount) Address() common.Address {
	return acc.Address_
}

// Name returns the name of the account
func (acc *MultiAccount) Name() string {
	return acc.Name_
}

// Clone returns the clonend value of it
func (acc *MultiAccount) Clone() types.Account {
	KeyHashes := make([]common.PublicHash, 0, len(acc.KeyHashes))
	for _, pubhash := range acc.KeyHashes {
		KeyHashes = append(KeyHashes, pubhash.Clone())
	}
	c := &MultiAccount{
		Address_:  acc.Address_,
		Name_:     acc.Name_,
		Required:  acc.Required,
		KeyHashes: KeyHashes,
	}
	return c
}

// Validate validates account signers
func (acc *MultiAccount) Validate(loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(acc.KeyHashes) != len(signers) {
		return types.ErrInvalidSignerCount
	}
	signerMap := map[common.PublicHash]bool{}
	for _, signer := range signers {
		signerMap[signer] = true
	}
	matchCount := 0
	for _, pubhash := range acc.KeyHashes {
		if signerMap[pubhash] {
			matchCount++
		}
	}
	if matchCount != int(acc.Required) {
		return types.ErrInvalidAccountSigner
	}
	return nil
}

// MarshalJSON is a marshaler function
func (acc *MultiAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"required":`)
	if bs, err := json.Marshal(acc.Required); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hashes":`)
	buffer.WriteString(`[`)
	for i, pubhash := range acc.KeyHashes {
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
