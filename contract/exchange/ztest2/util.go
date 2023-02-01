package test

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

func MakeSlice(size uint8) []*big.Int {
	result := make([]*big.Int, size, size)
	for i := uint8(0); i < size; i++ {
		result[i] = big.NewInt(0)
	}
	return result
}
func MakeAmountSlice(size uint8) []*amount.Amount {
	result := make([]*amount.Amount, size, size)
	for i := uint8(0); i < size; i++ {
		result[i] = ToAmount(big.NewInt(0))
	}
	return result
}
func CloneSlice(input []*big.Int) []*big.Int {
	size := len(input)
	result := make([]*big.Int, size, size)
	for i := 0; i < size; i++ {
		result[i] = big.NewInt(0).Set(input[i])
	}
	return result
}
func CloneAmountSlice(input []*amount.Amount) []*amount.Amount {
	size := len(input)
	result := make([]*amount.Amount, size, size)
	for i := 0; i < size; i++ {
		result[i] = ToAmount(big.NewInt(0).Set(input[i].Int))
	}
	return result
}
