package engincontext

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

var (
	ENGIN_CONTRACT_PREFIX = byte(0x01)
)

type EnginContextContract struct {
	contAddr common.Address
	marster  common.Address
	cc       *types.ContractContext
}

func NewEnginContextContract(contAddr, marster common.Address, cc *types.ContractContext) *EnginContextContract {
	return &EnginContextContract{contAddr, marster, cc}
}

// ChainID returns the id of the chain
func (pc *EnginContextContract) ContractAddress() string {
	return pc.contAddr.String()
}

// ChainID returns the id of the chain
func (pc *EnginContextContract) Master() string {
	return pc.marster.String()
}

// ChainID returns the id of the chain
func (pc *EnginContextContract) ChainID() *big.Int {
	return pc.cc.ChainID()
}

// Version returns the version of the chain
func (pc *EnginContextContract) Version() uint16 {
	return pc.cc.Version()
}

// Hash returns the hash value of it
func (pc *EnginContextContract) Hash() string {
	return pc.cc.Hash().String()
}

// TargetHeight returns the recorded target height when ContractContext generation
func (pc *EnginContextContract) TargetHeight() uint32 {
	return pc.cc.TargetHeight()
}

// PrevHash returns the recorded prev hash when ContractContext generation
func (pc *EnginContextContract) PrevHash() string {
	return pc.cc.PrevHash().String()
}

// LastTimestamp returns the recorded prev timestamp when ContractContext generation
func (pc *EnginContextContract) LastTimestamp() uint64 {
	return pc.cc.LastTimestamp()
}

// From returns current signer address
func (pc *EnginContextContract) From() string {
	return pc.cc.From().String()
}

// IsGenerator returns the account is generator or not
func (pc *EnginContextContract) IsGenerator(addr string) bool {
	return pc.cc.IsGenerator(common.HexToAddress(addr))
}

// MainToken returns the MainToken
func (pc *EnginContextContract) MainToken() string {
	return pc.cc.MainToken().String()
}

func addPrefix(name []byte) []byte {
	return append([]byte{ENGIN_CONTRACT_PREFIX}, name...)
}

// ContractData returns the contract data from the top snapshot
func (pc *EnginContextContract) ContractData(name []byte) []byte {
	return pc.cc.ContractData(addPrefix(name))
}

// SetContractData inserts the contract data to the top snapshot
func (pc *EnginContextContract) SetContractData(name []byte, value []byte) {
	pc.cc.SetContractData(addPrefix(name), value)
}

// AccountData returns the account data from the top snapshot
func (pc *EnginContextContract) AccountData(addr string, name []byte) []byte {
	return pc.cc.AccountData(common.HexToAddress(addr), addPrefix(name))
}

// SetAccountData inserts the account data to the top snapshot
func (pc *EnginContextContract) SetAccountData(addr string, name []byte, value []byte) {
	pc.cc.SetAccountData(common.HexToAddress(addr), addPrefix(name), value)
}

// Seq returns the sequence of the target account
func (pc *EnginContextContract) AddrSeq(addr string) uint64 {
	return pc.cc.AddrSeq(common.HexToAddress(addr))
}

// NextSeq returns the next squence number
func (pc *EnginContextContract) NextSeq() uint32 {
	return pc.cc.NextSeq()
}

// IsContract returns is the contract
func (pc *EnginContextContract) IsContract(addr string) bool {
	return pc.cc.IsContract(common.HexToAddress(addr))
}

func (pc *EnginContextContract) Exec(Addr string, MethodName string, Args []interface{}) ([]interface{}, error) {
	is, err := pc.cc.Exec(pc.cc, common.HexToAddress(Addr), MethodName, Args)
	if err != nil {
		return nil, err
	}
	for i, v := range is {
		if am, ok := v.(*amount.Amount); ok {
			is[i] = am.Int
		}
	}
	return is, nil
}
