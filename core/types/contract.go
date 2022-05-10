package types

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
)

// Contract defines chain Contract functions
type Contract interface {
	Address() common.Address
	Master() common.Address
	Init(addr common.Address, master common.Address)
	OnCreate(cc *ContractContext, Args []byte) error
	OnReward(cc *ContractContext, b *Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error)
	Front() interface{}
	// ExecuteTransaction(cc *ContractContext, tx *Transaction, TXID string, Exec ExecFunc) error
}

// ChargeFee defines Chargeable Contract functions
type ChargeFee interface {
	ChargeFee(cc *ContractContext, fee *amount.Amount) error
}
