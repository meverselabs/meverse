package util

/////////// math  ///////////
// func Clone(a *big.Int) *big.Int {
// 	return big.NewInt(0).Set(a)
// }

// func Abs(a *big.Int) *big.Int {
// 	return big.NewInt(0).Abs(a)
// }

// func Sqrt(a *big.Int) *big.Int {
// 	return big.NewInt(0).Sqrt(a)
// }

// func Exp(a, b *big.Int) *big.Int {
// 	return big.NewInt(0).Exp(a, b, nil)
// }

// func Min(a, b *big.Int) *big.Int {
// 	if a.Cmp(b) < 0 {
// 		return a
// 	}
// 	return b
// }

// func MinInArray(a []*big.Int) (int, *big.Int) {
// 	size := len(a)
// 	if size == 0 {
// 		return 0, nil
// 	}
// 	if size == 1 {
// 		return 0, Clone(a[0])
// 	}
// 	idx := 0
// 	min := Clone(a[0])
// 	for i := 1; i < len(a); i++ {
// 		if a[i].Cmp(min) < 0 {
// 			idx = i
// 			min = Clone(a[i])
// 		}
// 	}
// 	return idx, min
// }

// func MaxInArray(a []*big.Int) (int, *big.Int) {
// 	size := len(a)
// 	if size == 0 {
// 		return 0, nil
// 	}
// 	if size == 1 {
// 		return 0, Clone(a[0])
// 	}
// 	idx := 0
// 	max := Clone(a[0])
// 	for i := 1; i < len(a); i++ {
// 		if a[i].Cmp(max) > 0 {
// 			idx = i
// 			max = Clone(a[i])
// 		}
// 	}
// 	return idx, max
// }

// func Add(a, b *big.Int) *big.Int {
// 	return big.NewInt(0).Add(a, b)
// }
// func Sub(a, b *big.Int) *big.Int {
// 	return big.NewInt(0).Sub(a, b)
// }
// func Mul(a, b *big.Int) *big.Int {
// 	return big.NewInt(0).Mul(a, b)
// }
// func Div(a, b *big.Int) *big.Int {
// 	return big.NewInt(0).Div(a, b)
// }

// func AddC(a *big.Int, b int64) *big.Int {
// 	return big.NewInt(0).Add(a, big.NewInt(b))
// }
// func SubC(a *big.Int, b int64) *big.Int {
// 	return big.NewInt(0).Sub(a, big.NewInt(b))
// }
// func MulC(a *big.Int, b int64) *big.Int {
// 	return big.NewInt(0).Mul(a, big.NewInt(b))
// }
// func DivC(a *big.Int, b int64) *big.Int {
// 	return big.NewInt(0).Div(a, big.NewInt(b))
// }

// func MulF(a *big.Int, b float64) *big.Int {
// 	f := float64(a.Int64()) * b
// 	bi := int64(f / float64(amount.FractionalMax))
// 	bf := int64(f - float64(bi)*float64(amount.FractionalMax))
// 	return AddC(MulC(big.NewInt(bi), amount.FractionalMax), bf)
// }

// func MulDiv(a, b, denominator *big.Int) *big.Int {
// 	return Div(Mul(a, b), denominator)
// }

// func MulDivC(a, b *big.Int, denominator int64) *big.Int {
// 	return DivC(Mul(a, b), denominator)
// }

// func Sum(a []*big.Int) *big.Int {
// 	result := big.NewInt(0)
// 	for i := 0; i < len(a); i++ {
// 		result = Add(result, a[i])
// 	}
// 	return result
// }

// func Pow10(a int) *big.Int {
// 	return Exp(big.NewInt(10), big.NewInt(int64(a)))
// }

// //max uint256 = 115792089237316195423570985008687907853269984665640564039457584007913129639935 ~ 1.1e+77
// //max float64 = 1.7976931348623157e+308
// //max float32 = 3.4028234663852886e+38
// func ToFloat64(a *big.Int) float64 {
// 	f, _ := new(big.Float).SetInt(a).Float64()
// 	return f
// }
