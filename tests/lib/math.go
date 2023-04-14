package testlib

import (
	"math/big"

	"github.com/meverselabs/meverse/common/amount"
)

// ToAmount converts big.Int to amount.Amount
func ToAmount(b *big.Int) *amount.Amount {
	return &amount.Amount{Int: b}
}

// Sub : sub
func Sub(a, b *big.Int) *big.Int {
	return big.NewInt(0).Sub(a, b)
}

// Exp : exponential
func Exp(a, b *big.Int) *big.Int {
	return big.NewInt(0).Exp(a, b, nil)
}
