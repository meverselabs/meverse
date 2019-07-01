package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// LoaderProcess is an interface to load state data from the target chain
type LoaderProcess interface {
	Name() string
	Version() uint16
	TargetHeight() uint32
	LastHash() hash.Hash256
	LastTimestamp() uint64
	Seq(addr common.Address) uint64
	Account(addr common.Address) (types.Account, error)
	AddressByName(Name string) (common.Address, error)
	IsExistAccount(addr common.Address) (bool, error)
	IsExistAccountName(Name string) (bool, error)
	AccountData(addr common.Address, name []byte) []byte
	AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error)
	IsExistUTXO(id uint64) (bool, error)
	UTXO(id uint64) (*types.UTXO, error)
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
