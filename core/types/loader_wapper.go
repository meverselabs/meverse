package types

import "github.com/fletaio/fleta/common"

// LoaderWrapper is an interface to load state data from the target chain
type LoaderWrapper interface {
	Loader
	AccountData(addr common.Address, name []byte) []byte
	AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error)
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
