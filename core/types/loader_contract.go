package types

import "github.com/fletaio/fleta_v2/common"

// ContractLoader defines functions that loads state data from the target chain
type ContractLoader interface {
	Loader
	ContractData(name []byte) []byte
	AccountData(addr common.Address, name []byte) []byte
}
