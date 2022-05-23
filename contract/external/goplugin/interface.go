package goplugin

import "math/big"

type IPluginContract interface {
	Invoke(cc interface{}, plugin interface{}, MethodName string, param ...interface{}) ([]interface{}, error)
	ResigeredMethods() ([]string, error)
}

type ContractContext interface {
	// ChainID returns the id of the chain
	ContractAddress() string
	// ChainID returns the id of the chain
	Master() string
	// ChainID returns the id of the chain
	ChainID() *big.Int
	// Version returns the version of the chain
	Version() uint16
	// Hash returns the hash value of it
	Hash() string
	// TargetHeight returns the recorded target height when ContractContext generation
	TargetHeight() uint32
	// PrevHash returns the recorded prev hash when ContractContext generation
	PrevHash() string
	// LastTimestamp returns the recorded prev timestamp when ContractContext generation
	LastTimestamp() uint64
	// From returns current signer address
	From() string
	// IsGenerator returns the account is generator or not
	IsGenerator(addr string) bool
	// MainToken returns the MainToken
	MainToken() string
	// ContractData returns the contract data from the top snapshot
	ContractData(name []byte) []byte
	// SetContractData inserts the contract data to the top snapshot
	SetContractData(name []byte, value []byte)
	// AccountData returns the account data from the top snapshot
	AccountData(addr string, name []byte) []byte
	// SetAccountData inserts the account data to the top snapshot
	SetAccountData(addr string, name []byte, value []byte)
	// Seq returns the sequence of the target account
	AddrSeq(addr string) uint64
	// NextSeq returns the next squence number
	NextSeq() uint32
	// IsContract returns is the contract
	IsContract(addr string) bool

	Exec(Addr string, MethodName string, Args []interface{}) ([]interface{}, error)
}
