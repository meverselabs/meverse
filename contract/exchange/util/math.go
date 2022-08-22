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

	d1 := big.NewFloat(b * b * c * c)
	d2 := big.NewFloat(4 * b * b * b * d)
	d3 := big.NewFloat(4 * c * c * c * a)
	d4 := big.NewFloat(18 * a * b * c * d)
	d5 := big.NewFloat(27 * a * a * d * d)
	// D := d1 - d2 - d3 + d4 - d5 // 판별식
	bd := big.NewFloat(0)
	D, _ := bd.Sub(d1, d2).Sub(bd, d3).Add(bd, d4).Sub(bd, d5).Float64() // 판별식

	s1 := big.NewFloat(2 * b * b * b)
	s2 := big.NewFloat(9 * a * b * c)
	s3 := big.NewFloat(27 * a * a * d)
	bs := big.NewFloat(0)
	// s := 2*b*b*b - 9*a*b*c + 27*a*a*d
	s, _ := bs.Sub(s1, s2).Add(bs, s3).Float64()

	if D >= 0 { // 값의 결과상 x_3만 필요, x_1, x_2 < 0
		t := cmplx.Sqrt(complex(s*s-4*math.Pow((b*b-3*a*c), 3.0), 0))
		p := cmplx.Pow((complex(s, 0)+t)/2., complex(1./3., 0)) // plus
		n := cmplx.Pow((complex(s, 0)-t)/2., complex(1./3., 0)) // negative
		//x_1 := (-complex(b, 0) - p - n) / complex(3*a, 0)
		//x_2 := (-complex(b, 0) + complex(0.5, 0.5*math.Sqrt(3.))*p + complex(0.5, -0.5*math.Sqrt(3.))*n) / complex(3*a, 0)

		x1 := -complex(b, 0)
		x2 := complex(0.5, -0.5*math.Sqrt(3.))
		x3 := complex(0.5, 0.5*math.Sqrt(3.))
		// x4 := complex(3*a, 0)
		x_3 := x1
		{
			x5 := complexMul(p, x2)
			x_3 = complexAdd(x_3, x5)
			x6 := complexMul(x3, n)
			x_3 = complexAdd(x_3, x6)
			x_3 = x_3 / complex(3*a, 0)
		}

		if math.Abs(imag(x_3)) > 1e-20 {
			return 0., errors.New("NOT_REAL")
		}
		rx_3 := real(x_3)
		return rx_3, nil
	}

	// 실수해 1개인 경우
	// s := 2*b*b*b - 9*a*b*c + 27*a*a*d
	t1 := big.NewFloat(s * s)
	t2 := big.NewFloat(b * b)
	t3 := big.NewFloat(3 * a * c)
	t4 := t2.Sub(t2, t3)
	bt := big.NewFloat(0)
	ft4, _ := t4.Float64()
	t4 = big.NewFloat(4 * math.Pow(ft4, 3.0))
	bt = bt.Sub(t1, t4)
	ft, _ := bt.Float64()
	t := math.Sqrt(ft)
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

func complexMul(a, b complex128) complex128 {
	bxr, bxi := _complexMul(a, b)
	fbxr, _ := bxr.Float64()
	fbxi, _ := bxi.Float64()
	return complex(fbxr, fbxi)
}

func _complexMul(a, b complex128) (*big.Float, *big.Float) {
	pr := big.NewFloat(real(a))
	pi := big.NewFloat(imag(a))
	xr := big.NewFloat(real(b))
	xi := big.NewFloat(imag(b))

	x_1 := big.NewFloat(0).Mul(pr, xr)
	x_2 := big.NewFloat(0).Mul(pr, xi)
	x_3 := big.NewFloat(0).Mul(pi, xr)
	x_4 := big.NewFloat(0).Mul(pi, xi)

	bxr := big.NewFloat(0).Sub(x_1, x_4)
	bxi := big.NewFloat(0).Add(x_2, x_3)
	return bxr, bxi
}

func complexAdd(a, b complex128) complex128 {
	pr := big.NewFloat(real(a))
	pi := big.NewFloat(imag(a))
	x2r := big.NewFloat(real(b))
	x2i := big.NewFloat(imag(b))
	bx5 := big.NewFloat(0)
	bx5r, _ := bx5.Add(pr, x2r).Float64()
	bx5i, _ := bx5.Add(pi, x2i).Float64()
	return complex(bx5r, bx5i)
}

// 3차방정식 해 - Newtonian Method
//x0  initial
func Newtonian(a, b, c, d, x0 float64) (float64, error) {
	if Approx(a, 0., 1e-10) {
		return 0., errors.New("NOT_CUBIC_FUNCTION")
	}

	j := 0
	x_p := x0
	var x float64
	for {
		x = (2*a*x_p*x_p*x_p + b*x_p*x_p - d) / (3*a*x_p*x_p + 2*b*x_p + c)
		if math.Abs(x-x_p) < 1 {
			return x, nil
		}
		j++
		x_p = x
		if j > 256 {
			return x, nil
		}
	}
}
