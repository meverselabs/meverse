package util

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
)

var (
	Zero       = big.NewInt(0)
	ZeroAmount = amount.NewAmount(0, 0)
	MaxUint256 = ToAmount(Sub(Exp(big.NewInt(2), big.NewInt(256)), big.NewInt(1)))

	ZeroAddress = common.Address{}
)
