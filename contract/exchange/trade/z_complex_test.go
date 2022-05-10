package trade

import (
	"errors"
	"math"
	"math/big"
	"math/cmplx"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	//	. "github.com/onsi/gomega"
)

func Cbrt(x complex128) complex128 {
	z := complex128(1)
	for i := 0; i < 10; i++ {
		z = z - (z*z*z-x)/(3*z*z)
	}
	return z
}

// (1-f)x^3 + 3(1-f)x_0*x^2 + ((2-f)x_0 - (1-f)A)*x_0*x - A*x_0^2
func originalOptimalCubicRoot(fee uint64, amountIn *big.Int, reserve *big.Int) (complex128, complex128, complex128, error) {
	f := float64(fee) / float64(FEE_DENOMINATOR)
	A := ToFloat64(amountIn)
	x_0 := ToFloat64(reserve)
	a := 1. - f
	b := 3. * (1 - f) * x_0
	c := ((2.-f)*x_0 - (1.-f)*A) * x_0
	d := -A * x_0 * x_0
	GPrintln(a, b, c, d)
	return orignialCubicRoot(a, b, c, d)
}

func orignialCubicRoot(a, b, c, d float64) (complex128, complex128, complex128, error) {
	if Approx(a, 0., 1e-10) {
		return 0., 0., 0., errors.New("NOT_CUBIC_FUNCTION")
	}
	s := 2*b*b*b - 9*a*b*c + 27*a*a*d
	t := cmplx.Sqrt(complex(s*s-4*math.Pow((b*b-3*a*c), 3.0), 0))

	x_1 := (-complex(b, 0) - cmplx.Pow((complex(s, 0)+t)/2, 1.0/3) - cmplx.Pow((complex(s, 0)-t)/2, 1.0/3)) / complex(3*a, 0)
	x_2 := (-complex(b, 0) + complex(0.5, 0.5*math.Sqrt(3.))*cmplx.Pow((complex(s, 0)+t)/2, 1.0/3) + complex(0.5, -0.5*math.Sqrt(3.))*cmplx.Pow((complex(s, 0)-t)/2, 1.0/3)) / complex(3*a, 0)
	x_3 := (-complex(b, 0) + complex(0.5, -0.5*math.Sqrt(3.))*cmplx.Pow((complex(s, 0)+t)/2, 1.0/3) + complex(0.5, 0.5*math.Sqrt(3.))*cmplx.Pow((complex(s, 0)-t)/2, 1.0/3)) / complex(3*a, 0)

	return x_1, x_2, x_3, nil
}

var _ = Describe("complex", func() {

	It("Complex", func() {
		var c1 complex128 = 1 + 1i
		var c2 complex128 = 2 + 2i

		c3 := c1 + c2
		GPrintln(c3)

		c4 := cmplx.Sqrt(c3)
		GPrintln(c4)

		r := cmplx.Pow(c4, 1.0/3)
		GPrintln(r)

		GPrintln(Cbrt(2+3i), cmplx.Pow(2.0+3i, 1.0/3))
	})

	It("math.Pow(x,1./3)", func() {
		GPrintln(math.Pow(27, 1./3))
	})

	It("CubicRoot 3중근", func() {

		//x ^ 3 - x = 0
		GPrintln(CubicRoot(1., 0, -1, 0))

		// (x-1)(x-2)(x-3) = x^3 - 6*x^2 + 11*x - 6
		GPrintln(CubicRoot(1, -6, 11, -6))

		// x^3 + x
		GPrintln(CubicRoot(1, 0, -1, 0))

		// x^3 - 2500x + 2
		GPrintln(CubicRoot(1, 0, -2500, 2))

		// x^3 + 150x^2 + 5000x + 2
		GPrintln(CubicRoot(1, 0, -1601, -43624))

		// x^3 + 150x^2 + 5000x + 2
		x, _ := CubicRoot(0.75, 112, 4374, -1.46)
		GPrintlnT("x", x, CubicFunc(0.75, 112, 4374, -1.46, x))

		//x, _ := CubicRoot(1, 1, -1, -1)
		GPrintlnT("x", x, CubicFunc(1, 1, -1, -1, x))

		GPrintln(CubicRoot(1, -2, 1, 0))

	})

	It("OptimalCubicRoot", func() {
		// f = 0  근 = Sqrt(x_0*(x_0 + A)) - x_0
		GPrintln(originalOptimalCubicRoot(0, big.NewInt(1), big.NewInt(10)))
		GPrintln("root=", math.Sqrt(10*(10+1))-10)

		GPrintln(originalOptimalCubicRoot(10000000000, big.NewInt(300000), big.NewInt(1000000000000000000)))
	})
})
