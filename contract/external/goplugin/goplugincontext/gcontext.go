package goplugincontext

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
)

type PluginContextContract struct {
	cont types.Contract
	cc   *types.ContractContext
}

func NewPluginContextContract(cont types.Contract, cc *types.ContractContext) *PluginContextContract {
	return &PluginContextContract{cont, cc}
}

// ChainID returns the id of the chain
func (pc *PluginContextContract) ContractAddress() string {
	return pc.cont.Address().String()
}

// ChainID returns the id of the chain
func (pc *PluginContextContract) Master() string {
	return pc.cont.Master().String()
}

// ChainID returns the id of the chain
func (pc *PluginContextContract) ChainID() *big.Int {
	return pc.cc.ChainID()
}

// Version returns the version of the chain
func (pc *PluginContextContract) Version() uint16 {
	return pc.cc.Version()
}

// Hash returns the hash value of it
func (pc *PluginContextContract) Hash() string {
	return pc.cc.Hash().String()
}

// TargetHeight returns the recorded target height when ContractContext generation
func (pc *PluginContextContract) TargetHeight() uint32 {
	return pc.cc.TargetHeight()
}

// PrevHash returns the recorded prev hash when ContractContext generation
func (pc *PluginContextContract) PrevHash() string {
	return pc.cc.PrevHash().String()
}

// LastTimestamp returns the recorded prev timestamp when ContractContext generation
func (pc *PluginContextContract) LastTimestamp() uint64 {
	return pc.cc.LastTimestamp()
}

// From returns current signer address
func (pc *PluginContextContract) From() string {
	return pc.cc.From().String()
}

// IsGenerator returns the account is generator or not
func (pc *PluginContextContract) IsGenerator(addr string) bool {
	return pc.cc.IsGenerator(common.HexToAddress(addr))
}

// MainToken returns the MainToken
func (pc *PluginContextContract) MainToken() string {
	return pc.cc.MainToken().String()
}

// ContractData returns the contract data from the top snapshot
func (pc *PluginContextContract) ContractData(name []byte) []byte {
	return pc.cc.ContractData(name)
}

// SetContractData inserts the contract data to the top snapshot
func (pc *PluginContextContract) SetContractData(name []byte, value []byte) {
	pc.cc.SetContractData(name, value)
}

// AccountData returns the account data from the top snapshot
func (pc *PluginContextContract) AccountData(addr string, name []byte) []byte {
	return pc.cc.AccountData(common.HexToAddress(addr), name)
}

// SetAccountData inserts the account data to the top snapshot
func (pc *PluginContextContract) SetAccountData(addr string, name []byte, value []byte) {
	pc.cc.SetAccountData(common.HexToAddress(addr), name, value)
}

// Seq returns the sequence of the target account
func (pc *PluginContextContract) AddrSeq(addr string) uint64 {
	return pc.cc.AddrSeq(common.HexToAddress(addr))
}

// NextSeq returns the next squence number
func (pc *PluginContextContract) NextSeq() uint32 {
	return pc.cc.NextSeq()
}

// IsContract returns is the contract
func (pc *PluginContextContract) IsContract(addr string) bool {
	return pc.cc.IsContract(common.HexToAddress(addr))
}

func (pc *PluginContextContract) Exec(Addr string, MethodName string, Args []interface{}) ([]interface{}, error) {
	return pc.cc.Exec(pc.cc, common.HexToAddress(Addr), MethodName, Args)
}
