package testlib

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meverselabs/meverse/common/amount"
)

const GasLimit uint64 = 0x1DCD6500 // 500000000

var (
	Zero       = big.NewInt(0)
	ZeroAmount = amount.NewAmount(0, 0)
	MaxUint256 = ToAmount(Sub(Exp(big.NewInt(2), big.NewInt(256)), big.NewInt(1)))

	ZeroAddress = common.Address{}
)

const (
	ChainDataPath   = "_data"
	BloomDataPath   = "_bloombits"
	BloomBitsBlocks = 24
	BloomConfirms   = 10
)

var (
	ChainID = big.NewInt(61337) // mainnet 1,  testnet 65535, local 61337
	Version = uint16(2)

	GasPrice = big.NewInt(748488682)
)
