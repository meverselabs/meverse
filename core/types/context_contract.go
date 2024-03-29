package types

import (
	"math"
	"math/big"
	"reflect"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/ethereum/core/defaultevm"
	"github.com/meverselabs/meverse/ethereum/core/vm"
)

// ContractContext is an context for the contract
type ContractContext struct {
	cont common.Address
	from common.Address
	ctx  *Context
	Exec ExecFunc
}

// ChainID returns the id of the chain
func (cc *ContractContext) ChainID() *big.Int {
	return cc.ctx.ChainID()
}

// Version returns the version of the chain
func (cc *ContractContext) Version(h uint32) uint16 {
	return cc.ctx.Version(h)
}

// Hash returns the hash value of it
func (cc *ContractContext) Hash() hash.Hash256 {
	return cc.ctx.Hash()
}

// TargetHeight returns the recorded target height when ContractContext generation
func (cc *ContractContext) TargetHeight() uint32 {
	return cc.ctx.TargetHeight()
}

// PrevHash returns the recorded prev hash when ContractContext generation
func (cc *ContractContext) PrevHash() hash.Hash256 {
	return cc.ctx.PrevHash()
}

// LastTimestamp returns the recorded prev timestamp when ContractContext generation
func (cc *ContractContext) LastTimestamp() uint64 {
	return cc.ctx.LastTimestamp()
}

// From returns current signer address
func (cc *ContractContext) From() common.Address {
	return cc.from
}

// IsGenerator returns the account is generator or not
func (cc *ContractContext) IsGenerator(addr common.Address) bool {
	return cc.ctx.Top().IsGenerator(addr)
}

// SetGenerator adds the account as a generator or not
// only formulator contract can call
func (cc *ContractContext) SetGenerator(addr common.Address, is bool) error {
	if cont, err := cc.ctx.Top().Contract(cc.cont); err != nil {
		return err
	} else {
		if t := reflect.TypeOf(cont); t.Elem().String() != "formulator.FormulatorContract" {
			return ErrOnlyFormulatorAllowed
		}
	}

	return cc.ctx.Top().SetGenerator(addr, is)
}

// MainToken returns the MainToken
func (cc *ContractContext) MainToken() *common.Address {
	return cc.ctx.Top().MainToken()
}

// ContractData returns the contract data from the top snapshot
func (cc *ContractContext) ContractData(name []byte) []byte {
	return cc.ctx.Top().Data(cc.cont, common.Address{}, name)
}

// DeployContract deploy contract to the chain
func (cc *ContractContext) DeployContractWithAddress(owner common.Address, ClassID uint64, addr common.Address, Args []byte) (Contract, error) {
	cc.ctx.isLatestHash = false
	return cc.ctx.Top().DeployContractWithAddress(owner, ClassID, addr, Args)
}

// SetContractData inserts the contract data to the top snapshot
func (cc *ContractContext) SetContractData(name []byte, value []byte) {
	cc.ctx.Top().SetData(cc.cont, common.Address{}, name, value)
}

// AccountData returns the account data from the top snapshot
func (cc *ContractContext) AccountData(addr common.Address, name []byte) []byte {
	return cc.ctx.Top().Data(cc.cont, addr, name)
}

// SetAccountData inserts the account data to the top snapshot
func (cc *ContractContext) SetAccountData(addr common.Address, name []byte, value []byte) {
	cc.ctx.Top().SetData(cc.cont, addr, name, value)
}

// IsUsedTimeSlot returns timeslot is used or not
func (cc *ContractContext) IsUsedTimeSlot(slot uint32, key string) bool {
	return cc.ctx.Top().IsUsedTimeSlot(slot, key)
}

// Seq returns the sequence of the target account
func (cc *ContractContext) AddrSeq(addr common.Address) uint64 {
	return cc.ctx.Top().AddrSeq(addr)
}

// AddSeq update the sequence of the target account
func (cc *ContractContext) AddAddrSeq(addr common.Address) {
	cc.ctx.Top().AddAddrSeq(addr)
}

// NextSeq returns the next squence number
func (cc *ContractContext) NextSeq() uint32 {
	return cc.ctx.Top().NextSeq()
}

// IsContract returns is the contract
func (cc *ContractContext) IsContract(addr common.Address) bool {
	return cc.ctx.Top().IsContract(addr)
}

// EvmCall execute evm.Call function and returns result, usedGas, error
func (cc *ContractContext) EvmCall(caller vm.ContractRef, to common.Address, input []byte) ([]byte, uint64, error) {
	statedb := NewStateDB(cc.ctx)
	evm := defaultevm.DefaultEVM(statedb, nil)

	inputGas := uint64(math.MaxUint64)
	ret, leftOverGas, err := evm.Call(caller, to, input, inputGas, big.NewInt(0))
	return ret, inputGas - leftOverGas, err
}
