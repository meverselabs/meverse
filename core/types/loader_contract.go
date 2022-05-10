package types

import "github.com/meverselabs/meverse/common"

// ContractLoader defines functions that loads state data from the target chain
type ContractLoader interface {
	Loader
	ContractData(name []byte) []byte
	AccountData(addr common.Address, name []byte) []byte
}
