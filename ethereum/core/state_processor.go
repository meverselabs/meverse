package core

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/meverselabs/meverse/ethereum/core/defaultevm"
	"github.com/meverselabs/meverse/ethereum/core/types"
	"github.com/meverselabs/meverse/ethereum/core/vm"
	"github.com/meverselabs/meverse/ethereum/istate"
	"github.com/meverselabs/meverse/ethereum/params"
)

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(statedb istate.IStateDB, tx *etypes.Transaction) (*etypes.Receipt, error) {

	bc := ChainContext{}
	author := common.Address{}
	//gp := new(core.GasPool).AddGas(8000000) //ethconfig.Defaults.Miner.GasCeil
	gp := new(core.GasPool).AddGas(math.MaxUint64) //eth_call 설정
	bh := statedb.Height()
	blockNumber := big.NewInt(int64(bh))
	blockHash := statedb.Hash() //실제 나중의 blockhash가 아닐 수도 있음
	usedGas := new(uint64)

	msg, err := tx.AsMessage(types.MakeSigner(statedb.ChainID(), bh), nil)
	if err != nil {
		return nil, err
	}
	config := &params.ChainConfig{
		ChainID: statedb.ChainID(),
	}

	vmenv, _ := defaultevm.DefaultEVMWithConfig(statedb, config, false)
	return applyTransaction(msg, config, bc, &author, gp, statedb, blockNumber, blockHash, tx, usedGas, vmenv)

}

// ethereum.go와 일치
func applyTransaction(msg etypes.Message, config *params.ChainConfig, bc ChainContext, author *common.Address, gp *core.GasPool, statedb istate.IStateDB, blockNumber *big.Int, blockHash common.Hash, tx *etypes.Transaction, usedGas *uint64, evm *vm.EVM) (*etypes.Receipt, error) {
	// Create a new context to be used in the EVM environment.
	txContext := NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)

	// Apply the transaction to the current state (included in the env).
	result, err := ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, err
	} else if result.Err != nil {
		return nil, result.Err
	}

	// Update the state with pending changes.
	var root []byte
	if config.IsByzantium(blockNumber) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &etypes.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	if result.Failed() {
		receipt.Status = etypes.ReceiptStatusFailed
	} else {
		receipt.Status = etypes.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas

	// If the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}

	// Set the receipt logs and create the bloom filter.
	receipt.Logs = statedb.GetLogs(tx.Hash(), blockHash)
	receipt.Bloom = etypes.CreateBloom(etypes.Receipts{receipt})
	receipt.BlockHash = blockHash
	receipt.BlockNumber = blockNumber
	receipt.TransactionIndex = uint(statedb.TxIndex())
	return receipt, err
}
