package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meverselabs/meverse/ethereum/core/state"
	"github.com/meverselabs/meverse/ethereum/core/vm"
	"github.com/meverselabs/meverse/ethereum/params"
)

type ChainContext struct{}

// NewEVMTxContext creates a new transaction context for a single transaction.
func NewEVMTxContext(msg Message) vm.TxContext {
	return vm.TxContext{
		Origin:   msg.From(),
		GasPrice: new(big.Int).Set(msg.GasPrice()),
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	// return db.GetBalance(addr).Cmp(amount) >= 0
	return true
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	// db.SubBalance(sender, amount)
	// db.AddBalance(recipient, amount)
}

// Default EVM with statedb
func DefaultEVM(statedb *state.StateDB) *vm.EVM {
	config := &params.ChainConfig{
		ChainID: statedb.ChainID(),
	}
	return DefaultEVMWithConfig(statedb, config)
}

// Default EVM with statedb and config
func DefaultEVMWithConfig(statedb *state.StateDB, config *params.ChainConfig) *vm.EVM {

	blockContext := vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		BlockNumber: big.NewInt(int64(statedb.Height())),
		BaseFee:     new(big.Int),
		Time:        new(big.Int).SetUint64(statedb.LastTimestamp()),
	}
	cfg := vm.Config{
		EnablePreimageRecording: false,
		NoBaseFee:               true,
	}

	return vm.NewEVM(blockContext, vm.TxContext{}, statedb, config, cfg)
}
