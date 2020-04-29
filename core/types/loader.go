package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

// Loader defines functions that loads state data from the target chain
type Loader interface {
	ChainID() uint8
	Name() string
	Version() uint16
	TargetHeight() uint32
	Seq(addr common.Address) uint64
	Account(addr common.Address) (Account, error)
	AddressByName(Name string) (common.Address, error)
	HasAccount(addr common.Address) (bool, error)
	HasAccountName(Name string) (bool, error)
	HasUTXO(id uint64) (bool, error)
	UTXO(id uint64) (*UTXO, error)
}

type internalLoader interface {
	Loader
	LastHash() hash.Hash256
	LastTimestamp() uint64
	AccountData(addr common.Address, pid uint8, name []byte) []byte
	ProcessData(pid uint8, name []byte) []byte
}

type emptyLoader struct {
}

// newEmptyLoader is used for generating genesis state
func newEmptyLoader() internalLoader {
	return &emptyLoader{}
}

// ChainID returns 0
func (st *emptyLoader) ChainID() uint8 {
	return 0
}

// Name returns ""
func (st *emptyLoader) Name() string {
	return ""
}

// Version returns 0
func (st *emptyLoader) Version() uint16 {
	return 0
}

// TargetHeight returns 0
func (st *emptyLoader) TargetHeight() uint32 {
	return 0
}

// LastStatus returns 0, hash.Hash256{}
func (st *emptyLoader) LastStatus() (uint32, hash.Hash256) {
	return 0, hash.Hash256{}
}

// LastHash returns hash.Hash256{}
func (st *emptyLoader) LastHash() hash.Hash256 {
	return hash.Hash256{}
}

// LastTimestamp returns 0
func (st *emptyLoader) LastTimestamp() uint64 {
	return 0
}

// Seq returns 0
func (st *emptyLoader) Seq(addr common.Address) uint64 {
	return 0
}

// Account returns ErrNotExistAccount
func (st *emptyLoader) Account(addr common.Address) (Account, error) {
	return nil, ErrNotExistAccount
}

// AddressByName returns ErrNotExistAccount
func (st *emptyLoader) AddressByName(Name string) (common.Address, error) {
	return common.Address{}, ErrNotExistAccount
}

// HasAccount returns false
func (st *emptyLoader) HasAccount(addr common.Address) (bool, error) {
	return false, nil
}

// HasAccountName returns false
func (st *emptyLoader) HasAccountName(Name string) (bool, error) {
	return false, nil
}

// AccountData returns nil
func (st *emptyLoader) AccountData(addr common.Address, pid uint8, name []byte) []byte {
	return nil
}

// HasUTXO returns false
func (st *emptyLoader) HasUTXO(id uint64) (bool, error) {
	return false, nil
}

// UTXO returns ErrNotExistUTXO
func (st *emptyLoader) UTXO(id uint64) (*UTXO, error) {
	return nil, ErrNotExistUTXO
}

// ProcessData returns nil
func (st *emptyLoader) ProcessData(pid uint8, name []byte) []byte {
	return nil
}
