package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// StateContext defines functions that Evm use as Database
type StateDBContext interface {
	ChainID() *big.Int
	TargetHeight() uint32
	Hash() common.Hash
	AddrSeq(addr common.Address) uint64
	SetNonce(addr common.Address, nonce uint64)
	Data(cont, addr common.Address, name []byte) []byte
	SetData(cont, addr common.Address, name, value []byte)
	IsContract(addr common.Address) bool
	Contract(addr common.Address) (Contract, error)
	ContractContext(cont Contract, from common.Address) *ContractContext
	Snapshot() int
	Revert(sn int)
	Commit(sn int)
}
