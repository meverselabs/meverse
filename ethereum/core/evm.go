package core

import (
	"math/big"

	"github.com/meverselabs/meverse/ethereum/core/vm"
)

type ChainContext struct{}

// NewEVMTxContext creates a new transaction context for a single transaction.
func NewEVMTxContext(msg Message) vm.TxContext {
	return vm.TxContext{
		Origin:   msg.From(),
		GasPrice: new(big.Int).Set(msg.GasPrice()),
	}
}
