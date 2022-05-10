package util

import (
	"math/big"

	"github.com/meverselabs/meverse/common/amount"
)

func ToAmount(b *big.Int) *amount.Amount {
	return &amount.Amount{Int: b}
}
func ToAmounts(b []*big.Int) []*amount.Amount {
	size := len(b)
	result := make([]*amount.Amount, size, size)
	for i := 0; i < size; i++ {
		result[i] = ToAmount(b[i])
	}
	return result
}
func ToBigInts(b []*amount.Amount) []*big.Int {
	size := len(b)
	result := make([]*big.Int, size, size)
	for i := 0; i < size; i++ {
		result[i] = Clone(b[i].Int)
	}
	return result
}
