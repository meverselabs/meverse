package types

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
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
func (cc *ContractContext) Version() uint16 {
	return cc.ctx.Version()
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
