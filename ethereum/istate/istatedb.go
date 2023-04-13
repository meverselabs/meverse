package istate

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/core/ctypes"
)

type IStateDB interface {
	ChainID() *big.Int
	TargetHeight() uint32
	Height() uint32
	Hash() common.Hash
	LastTimestamp() uint64
	StartPrefetcher(namespace string)
	StopPrefetcher()
	Error() error
	AddLog(log *etypes.Log)
	GetLogs(hash common.Hash, blockHash common.Hash) []*etypes.Log
	Logs() []*etypes.Log
	AddPreimage(hash common.Hash, preimage []byte)
	Preimages() map[common.Hash][]byte
	AddRefund(gas uint64)
	SubRefund(gas uint64)
	Exist(addr common.Address) bool
	Empty(addr common.Address) bool
	GetBalance(addr common.Address) *big.Int
	GetNonce(addr common.Address) uint64
	TxIndex() int
	GetCode(addr common.Address) []byte
	GetCodeSize(addr common.Address) int
	GetCodeHash(addr common.Address) common.Hash
	GetState(addr common.Address, hash common.Hash) common.Hash
	GetProof(addr common.Address) ([][]byte, error)
	GetProofByHash(addrHash common.Hash) ([][]byte, error)
	GetStorageProof(a common.Address, key common.Hash) ([][]byte, error)
	GetCommittedState(addr common.Address, hash common.Hash) common.Hash
	HasSuicided(addr common.Address) bool
	AddBalance(addr common.Address, amount *big.Int)
	SubBalance(addr common.Address, amount *big.Int)
	SetBalance(addr common.Address, amount *big.Int)
	SetNonce(addr common.Address, nonce uint64)
	SetCode(addr common.Address, code []byte)
	SetState(addr common.Address, key, value common.Hash)
	SetStorage(addr common.Address, storage map[common.Hash]common.Hash)
	Suicide(addr common.Address) bool
	CreateAccount(addr common.Address)
	ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error
	// Copy() *StateDB
	Snapshot() int
	Revert(revid int)
	RevertToSnapshot(revid int)
	GetRefund() uint64
	Finalise(deleteEmptyObjects bool)
	IntermediateRoot(deleteEmptyObjects bool) common.Hash
	Prepare(thash common.Hash, ti int)
	PrepareAccessList(sender common.Address, dst *common.Address, precompiles []common.Address, list etypes.AccessList)
	AddAddressToAccessList(addr common.Address)
	AddSlotToAccessList(addr common.Address, slot common.Hash)
	AddressInAccessList(addr common.Address) bool
	SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool)
	IsExtContract(addr common.Address) bool
	Exec(user common.Address, contAddr common.Address, input []byte, gas uint64) ([]byte, uint64, []*ctypes.Event, error)
	// getCC(contAddr common.Address, user common.Address) (*types.ContractContext, error)
}
