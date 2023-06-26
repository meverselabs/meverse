package types

import (
	"encoding/hex"
	"errors"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/extern/txparser"
)

var (
	tagCodeHash = []byte{0x01, 0x01}
	tagCodeSize = []byte{0x01, 0x02} // Dos 공격 방어
	tagState    = []byte{0x02}
)

type StateDB struct {
	ctx   *Context
	dbErr error

	thash   common.Hash
	txIndex int
	logs    map[common.Hash][]*etypes.Log
	logSize uint
}

func NewStateDB(ctx *Context) *StateDB {
	sdb := &StateDB{
		ctx:  ctx,
		logs: make(map[common.Hash][]*etypes.Log),
	}
	return sdb
}

// ChainID returns the id of the chain
func (s *StateDB) ChainID() *big.Int {
	return s.ctx.ChainID()
}

func (s *StateDB) TargetHeight() uint32 {
	return s.ctx.TargetHeight()
}

func (s *StateDB) Height() uint32 {
	return s.ctx.TargetHeight() - 1
}

func (s *StateDB) Hash() common.Hash {
	return s.ctx.Hash()
}

// LastTimestamp returns the prev timestamp of the chain
func (s *StateDB) LastTimestamp() uint64 {
	return s.ctx.LastTimestamp()
}

// StartPrefetcher initializes a new trie prefetcher to pull in nodes from the
// state trie concurrently while the state is mutated so that when we reach the
// commit phase, most of the needed data is already hot.
func (s *StateDB) StartPrefetcher(namespace string) {
}

// StopPrefetcher terminates a running prefetcher and reports any leftover stats
// from the gathered metrics.
func (s *StateDB) StopPrefetcher() {
}

// setError remembers the first non-nil error it is called with.
func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *StateDB) Error() error {
	return s.dbErr
}

func (s *StateDB) AddLog(log *etypes.Log) {
	log.TxHash = s.thash
	log.TxIndex = uint(s.txIndex)
	log.Index = s.logSize
	s.logs[s.thash] = append(s.logs[s.thash], log)
	s.logSize++
}

func (s *StateDB) GetLogs(hash common.Hash, blockHash common.Hash) []*etypes.Log {
	logs := s.logs[hash]
	for _, l := range logs {
		l.BlockHash = blockHash
	}
	return logs
}

func (s *StateDB) Logs() []*etypes.Log {
	var logs []*etypes.Log
	for _, lgs := range s.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (s *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (s *StateDB) Preimages() map[common.Hash][]byte {
	return nil
}

// AddRefund adds gas to the refund counter
func (s *StateDB) AddRefund(gas uint64) {
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (s *StateDB) SubRefund(gas uint64) {
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (s *StateDB) Exist(addr common.Address) bool {
	return s.GetNonce(addr) > 0 || s.GetCodeHash(addr) != common.Hash{}
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (s *StateDB) Empty(addr common.Address) bool {
	return !s.Exist(addr)
}

// GetBalance retrieves the balance from the given address or 0 if object not found
func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	bs := s.ctx.Data(*s.ctx.Top().MainToken(), addr, []byte{byte(0x10)}) // tagTokenAmount = byte(0x10)
	bi := big.NewInt(0).SetBytes(bs)

	return bi
	// return new(big.Int).Set(params.MaxUint256)
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	return s.ctx.AddrSeq(addr)
}

// TxIndex returns the current transaction index set by Prepare.
func (s *StateDB) TxIndex() int {
	return s.txIndex
}

func (s *StateDB) IsEvmContract(addr common.Address) bool {
	return len(s.GetCode(addr)) != 0
}

func (s *StateDB) GetCode(addr common.Address) []byte {
	hash := s.GetCodeHash(addr)
	return s.ctx.Data(addr, common.Address{}, hash[:])
}

func (s *StateDB) GetCodeSize(addr common.Address) int {
	// compiled contract의 경우 1을 true를 만들어 주기 위해
	if s.IsExtContract(addr) {
		return 1
	}
	bs := s.ctx.Data(addr, common.Address{}, tagCodeSize)
	if len(bs) > 0 {
		return int(bin.Uint32(bs))
	} else {
		return 0
	}
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	bs := s.ctx.Data(addr, common.Address{}, tagCodeHash)
	return common.BytesToHash(bs)
}

// GetState retrieves a value from the given account's storage trie.
func (s *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	bs := s.ctx.Data(addr, common.Address{}, append(tagState, hash[:]...))
	return common.BytesToHash(bs)
}

// GetProof returns the Merkle proof for a given account.
func (s *StateDB) GetProof(addr common.Address) ([][]byte, error) {
	return s.GetProofByHash(crypto.Keccak256Hash(addr.Bytes()))
}

// GetProofByHash returns the Merkle proof for a given account.
func (s *StateDB) GetProofByHash(addrHash common.Hash) ([][]byte, error) {
	return nil, nil
}

// GetStorageProof returns the Merkle proof for given storage slot.
func (s *StateDB) GetStorageProof(a common.Address, key common.Hash) ([][]byte, error) {
	return nil, nil
}

// GetCommittedState retrieves a value from the given account's committed storage trie.
func (s *StateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return common.Hash{}
}

func (s *StateDB) HasSuicided(addr common.Address) bool {
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (s *StateDB) AddBalance(addr common.Address, amt *big.Int) {
	if amt.Sign() < 0 {
		panic("balance can not be minus")
	}
	mainToken := *s.ctx.Top().MainToken()
	bs := s.ctx.Data(mainToken, addr, []byte{byte(0x10)}) // tagTokenAmount = byte(0x10)
	am := amount.NewAmountFromBytes(bs)
	added := am.Add(&amount.Amount{Int: amt})
	s.ctx.SetData(*s.ctx.cache.MainToken(), addr, []byte{byte(0x10)}, added.Bytes())
}

// SubBalance subtracts amount from the account associated with addr.
func (s *StateDB) SubBalance(addr common.Address, amt *big.Int) {
	if amt.Sign() < 0 {
		panic("balance can not be minus")
	}
	mainToken := *s.ctx.Top().MainToken()
	bs := s.ctx.Data(mainToken, addr, []byte{byte(0x10)}) // tagTokenAmount = byte(0x10)
	am := amount.NewAmountFromBytes(bs)
	subed := am.Sub(&amount.Amount{Int: amt})
	if subed.IsMinus() {
		panic("balance can not be minus")
	}

	s.ctx.SetData(*s.ctx.cache.MainToken(), addr, []byte{byte(0x10)}, subed.Bytes())
}

// SetBalance set amount of the account associated with addr.
func (s *StateDB) SetBalance(addr common.Address, amt *big.Int) {
	if amt.Sign() < 0 {
		panic("balance can not be minus")
	}

	rbs := amount.NewAmountFromBytes(amt.Bytes())
	s.ctx.SetData(*s.ctx.cache.MainToken(), addr, []byte{byte(0x10)}, rbs.Bytes())
}

// SetBalance set nonce of the account associated with addr.
func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	s.ctx.SetNonce(addr, nonce)
}

// SetCode save codeHash, size, code of the account(contract) associated with addr
func (s *StateDB) SetCode(addr common.Address, code []byte) {
	ln := uint32(len(code))
	if ln == 0 {
		return
	}
	hash := crypto.Keccak256Hash(code)
	s.ctx.SetData(addr, common.Address{}, tagCodeHash, hash[:])
	s.ctx.SetData(addr, common.Address{}, tagCodeSize, bin.Uint32Bytes(ln))
	s.ctx.SetData(addr, common.Address{}, hash[:], code)
	// log.Println("address:SetCode", addr)
	// log.Println("hash", hash, "size", ln)
	// log.Println("code", hex.EncodeToString(code))
}

// SetState save data of the account(contract) associated with addr
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
	s.ctx.SetData(addr, common.Address{}, append(tagState, key[:]...), value[:])
}

// SetStorage replaces the entire storage for the specified account with given
// storage. This function should only be used for debugging.
func (s *StateDB) SetStorage(addr common.Address, storage map[common.Hash]common.Hash) {
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (s *StateDB) Suicide(addr common.Address) bool {

	return true
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//  1. sends funds to sha(account ++ (nonce + 1))
//  2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (s *StateDB) CreateAccount(addr common.Address) {
	return
}

func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	return nil
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (s *StateDB) Copy() *StateDB {
	return nil
}

// Snapshot returns an identifier for the current revision of the state.
func (s *StateDB) Snapshot() int {
	return s.ctx.Snapshot()
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateDB) Revert(revid int) {
	s.ctx.Revert(revid)
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateDB) RevertToSnapshot(revid int) {
	s.Revert(revid)
}

// GetRefund returns the current value of the refund counter.
func (s *StateDB) GetRefund() uint64 {
	return 0
}

// Finalise finalises the state by removing the s destructed objects and clears
// the journal as well as the refunds. Finalise, however, will not push any updates
// into the tries just yet. Only IntermediateRoot or Commit will do that.
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return common.Hash{}
}

// Prepare sets the current transaction hash and index which are
// used when the EVM emits new state logs.
func (s *StateDB) Prepare(thash common.Hash, ti int) {
	s.thash = thash
	s.txIndex = ti
}

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// - Add sender to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
//
// This method should only be called if Berlin/2929+2930 is applicable at the current number.
func (s *StateDB) PrepareAccessList(sender common.Address, dst *common.Address, precompiles []common.Address, list etypes.AccessList) {
}

// AddAddressToAccessList adds the given address to the access list
func (s *StateDB) AddAddressToAccessList(addr common.Address) {
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (s *StateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
}

// AddressInAccessList returns true if the given address is in the access list.
func (s *StateDB) AddressInAccessList(addr common.Address) bool {
	return true
}

// SlotInAccessList returns true if the given (address, slot)-tuple is in the access list.
func (s *StateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
	return true, true
}

// IsContract checks if the given address is contract one.
func (s *StateDB) IsExtContract(addr common.Address) bool {
	return s.ctx.IsContract(addr)
}

// Exec executes the method of compiled contract in evm
func (s *StateDB) Exec(user common.Address, contAddr common.Address, input []byte, gas uint64) ([]byte, uint64, []*ctypes.Event, error) {
	intr, cc, err := s.getCC(contAddr, user)
	if err != nil {
		return nil, 0, nil, err
	}
	method := hex.EncodeToString(input[:4])
	m := txparser.Abi(method)
	if m.Name == "" {
		return nil, 0, nil, errors.New("not exist abi")
	}

	methodName := strings.ToUpper(string(m.Name[0])) + m.Name[1:]
	args, err := m.Inputs.Unpack(input[4:])
	if err != nil {
		return nil, 0, nil, err
	}

	is, err := cc.Exec(cc, contAddr, methodName, args)
	if err != nil {
		return nil, 0, nil, err
	}

	result := []interface{}{}
	for _, i := range is {
		if i != nil {
			if reflect.TypeOf(i).Kind() == reflect.Slice {
				s := reflect.ValueOf(i)
				if s.Len() > 0 {
					result = append(result, i)
				}
			} else {
				result = append(result, i)
			}
		}
	}

	bs, err := txparser.Outputs(m.Sig, result)
	if err != nil {
		return nil, 0, nil, err
	}

	gh := intr.GasHistory()
	usedGas := gh[0] - 21000

	if usedGas > gas {
		return nil, 0, nil, errors.New("out of gas")
	}

	return bs, gas - usedGas, intr.EventList(), nil
}

// getCC returns the ContractContext under contract and user context
func (s *StateDB) getCC(contAddr common.Address, user common.Address) (IInteractor, *ContractContext, error) {
	cont, err := s.ctx.Contract(contAddr)
	if err != nil {
		return nil, nil, err
	}
	cc := s.ctx.ContractContext(cont, user)
	intr := NewInteractor(s.ctx, cont, cc, "000000000000", true)
	cc.Exec = intr.Exec

	return intr, cc, nil
}

// BasicFee returns the basic Fee
func (s *StateDB) BasicFee() *big.Int {
	return s.ctx.BasicFee().Int
}
