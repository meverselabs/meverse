package util

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"

	"github.com/pkg/errors"
)

func IsPlus(a *big.Int) bool {
	return a.Cmp(Zero) > 0
}
func Clone(a *big.Int) *big.Int {
	return big.NewInt(0).Set(a)
}
func Abs(a *big.Int) *big.Int {
	return big.NewInt(0).Abs(a)
}
func Sqrt(a *big.Int) *big.Int {
	return big.NewInt(0).Sqrt(a)
}
func Exp(a, b *big.Int) *big.Int {
	return big.NewInt(0).Exp(a, b, nil)
}
func Min(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return a
	}
	return b
}
func MinInArray(a []*big.Int) (int, *big.Int) {
	size := len(a)
	if size == 0 {
		return 0, nil
	}
	if size == 1 {
		return 0, Clone(a[0])
	}
	idx := 0
	min := Clone(a[0])
	for i := 1; i < len(a); i++ {
		if a[i].Cmp(min) < 0 {
			idx = i
			min = Clone(a[i])
		}
	}
	return idx, min
}
func MaxInArray(a []*big.Int) (int, *big.Int) {
	size := len(a)
	if size == 0 {
		return 0, nil
	}
	if size == 1 {
		return 0, Clone(a[0])
	}
	idx := 0
	max := Clone(a[0])
	for i := 1; i < len(a); i++ {
		if a[i].Cmp(max) > 0 {
			idx = i
			max = Clone(a[i])
		}
	}
	return idx, max
}
func Add(a, b *big.Int) *big.Int {
	return big.NewInt(0).Add(a, b)
}
func Sub(a, b *big.Int) *big.Int {
	return big.NewInt(0).Sub(a, b)
}
func Mul(a, b *big.Int) *big.Int {
	return big.NewInt(0).Mul(a, b)
}
func Div(a, b *big.Int) *big.Int {
	return big.NewInt(0).Div(a, b)
}
func AddC(a *big.Int, b int64) *big.Int {
	return big.NewInt(0).Add(a, big.NewInt(b))
}
func SubC(a *big.Int, b int64) *big.Int {
	return big.NewInt(0).Sub(a, big.NewInt(b))
}
func MulC(a *big.Int, b int64) *big.Int {
	return big.NewInt(0).Mul(a, big.NewInt(b))
}
func DivC(a *big.Int, b int64) *big.Int {
	return big.NewInt(0).Div(a, big.NewInt(b))
}
func MulF(a *big.Int, b float64) (*big.Int, error) {
	f := float64(a.Int64()) * b
	return ToBigInt(f)
}
func MulDiv(a, b, denominator *big.Int) *big.Int {
	return Div(Mul(a, b), denominator)
}
func MulDivC(a, b *big.Int, denominator int64) *big.Int {
	return DivC(Mul(a, b), denominator)
}
func MulDivCC(a *big.Int, b, denominator int64) *big.Int {
	return DivC(MulC(a, b), denominator)
}
func Sum(a []*big.Int) *big.Int {
	result := big.NewInt(0)
	for i := 0; i < len(a); i++ {
		result = Add(result, a[i])
	}
	return result
}
func Pow10(a int) *big.Int {
	return Exp(big.NewInt(10), big.NewInt(int64(a)))
}

//max uint256 = 115792089237316195423570985008687907853269984665640564039457584007913129639935 ~ 1.1e+77
//max float64 = 1.7976931348623157e+308
//max float32 = 3.4028234663852886e+38
func ToFloat64(a *big.Int) float64 {
	f, _ := new(big.Float).SetInt(a).Float64()
	return f
}
func ToBigInt(a float64) (*big.Int, error) {
	s := fmt.Sprintf("%.0f", a)
	b := big.NewInt(0)
	result, err := b.SetString(s, 10)
	if err != true {
		return nil, errors.New("Util: FloatToBigInt")
	}
	return result, nil
}
func Approx(a, b, _precision float64) bool {
	if math.Abs(a) < 1e-36 && math.Abs(b) < 1e-36 {
		return true
	}
	return 2*math.Abs(a-b)/(a+b) <= _precision
}

func CubicFunc(a, b, c, d, x float64) float64 {
	return a*x*x*x + b*x*x + c*x + d
}

// 3차방정식 일반해 :
func CubicRoot(a, b, c, d float64) (float64, error) {
	if Approx(a, 0., 1e-10) {
		return 0., errors.New("NOT_CUBIC_FUNCTION")
	}
	D := b*b*c*c - 4*b*b*b*d - 4*c*c*c*a + 18*a*b*c*d - 27*a*a*d*d // 판별식
	if D >= 0 {                                                    // 값의 결과상 x_3만 필요, x_1, x_2 < 0
		s := 2*b*b*b - 9*a*b*c + 27*a*a*d
		t := cmplx.Sqrt(complex(s*s-4*math.Pow((b*b-3*a*c), 3.0), 0))
		p := cmplx.Pow((complex(s, 0)+t)/2., complex(1./3., 0)) // plus
		n := cmplx.Pow((complex(s, 0)-t)/2., complex(1./3., 0)) // negative
		//x_1 := (-complex(b, 0) - p - n) / complex(3*a, 0)
		//x_2 := (-complex(b, 0) + complex(0.5, 0.5*math.Sqrt(3.))*p + complex(0.5, -0.5*math.Sqrt(3.))*n) / complex(3*a, 0)
		x_3 := (-complex(b, 0) + complex(0.5, -0.5*math.Sqrt(3.))*p + complex(0.5, 0.5*math.Sqrt(3.))*n) / complex(3*a, 0)

		if math.Abs(imag(x_3)) > 1e-20 {
			return 0., errors.New("NOT_REAL")
		}
		return real(x_3), nil

	}

	// 실수해 1개인 경우
	s := 2*b*b*b - 9*a*b*c + 27*a*a*d
	t := math.Sqrt(s*s - 4*math.Pow((b*b-3*a*c), 3.0))
	var p, n float64
	if s+t > 0 {
		p = math.Pow((s+t)/2., 1./3)
	} else {
		p = -math.Pow(-(s+t)/2., 1./3)
	}
	if s-t > 0 {
		n = math.Pow((s-t)/2., 1./3)
	} else {
		n = -math.Pow(-(s-t)/2., 1./3)
	}

	x := (-b - p - n) / (3 * a)

	if x < 0 {
		return 0., errors.New("NOT_POSITIVE")
	}
	return x, nil
}
