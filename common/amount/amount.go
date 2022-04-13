package amount

import (
	"math/big"
	"strings"

	"github.com/pkg/errors"
)

// FractionalMax represent the max value of under the float point
const FractionalMax = 1000000000000000000

// FractionalCount represent the number of under the float point
const FractionalCount = 18

var zeroInt = big.NewInt(0)

var oneCoin = NewAmount(1, 0)
var ZeroCoin = NewAmount(0, 0)

// Amount is the precision float value based on the big.Int
type Amount struct {
	*big.Int
}

// NewAmount returns the amount that is consisted of the integer and the fractional value
func NewAmount(i uint64, f uint64) *Amount {
	if i == 0 {
		return &Amount{Int: big.NewInt(int64(f))}
	} else if f == 0 {
		bi := &Amount{Int: big.NewInt(int64(i))}
		return bi.MulC(FractionalMax)
	} else {
		bi := &Amount{Int: big.NewInt(int64(i))}
		bf := &Amount{Int: big.NewInt(int64(f))}
		return bi.MulC(FractionalMax).Add(bf)
	}
}

// NewAmountFromBytes parse the amount from the byte array
func NewAmountFromBytes(bs []byte) *Amount {
	b := &Amount{Int: big.NewInt(0)}
	b.Int.SetBytes(bs)
	return b
}

// MarshalJSON is a marshaler function
func (am *Amount) MarshalJSON() ([]byte, error) {
	return []byte(`"` + am.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (am *Amount) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return errors.WithStack(ErrInvalidAmountFormat)
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return errors.WithStack(ErrInvalidAmountFormat)
	}
	v, err := ParseAmount(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	am.Int = v.Int
	return nil
}

// Clone returns the clonend value of it
func (am *Amount) Clone() *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Add(am.Int, zeroInt)
	return c
}

// Add returns a + b (*immutable)
func (am *Amount) Add(b *Amount) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Add(am.Int, b.Int)
	return c
}

// Sub returns a - b (*immutable)
func (am *Amount) Sub(b *Amount) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Sub(am.Int, b.Int)
	return c
}

// Div returns a / b (*immutable)
func (am *Amount) Div(b *Amount) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Mul(am.Int, oneCoin.Int)
	d := &Amount{Int: big.NewInt(0)}
	d.Int.Div(c.Int, b.Int)
	return d
}

// DivC returns a / b (*immutable)
func (am *Amount) DivC(b int64) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Div(am.Int, big.NewInt(b))
	return c
}

// Mul returns a * b (*immutable)
func (am *Amount) Mul(b *Amount) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Mul(am.Int, b.Int)
	d := &Amount{Int: big.NewInt(0)}
	d.Int.Div(c.Int, oneCoin.Int)
	return d
}

// MulC returns a * b (*immutable)
func (am *Amount) MulC(b int64) *Amount {
	c := &Amount{Int: big.NewInt(0)}
	c.Int.Mul(am.Int, big.NewInt(b))
	return c
}

// IsZero returns a == 0
func (am *Amount) IsZero() bool {
	return am.Int.Cmp(zeroInt) == 0
}

// Less returns a < b
func (am *Amount) Less(b *Amount) bool {
	return am.Int.Cmp(b.Int) < 0
}

// Equal checks that two values is same or not
func (am *Amount) Equal(b *Amount) bool {
	return am.Int.Cmp(b.Int) == 0
}

// Less returns a < 0
func (am *Amount) IsMinus() bool {
	return am.Int.Cmp(zeroInt) < 0
}

// Less returns a > 0
func (am *Amount) IsPlus() bool {
	return am.Int.Cmp(zeroInt) > 0
}

// String returns the float string of the amount
func (am *Amount) String() string {
	if am.IsZero() {
		return "0"
	}
	str := am.Int.String()
	if len(str) <= FractionalCount {
		return "0." + formatFractional(str)
	} else {
		si := str[:len(str)-FractionalCount]
		sf := strings.TrimRight(str[len(str)-FractionalCount:], "0")
		if len(sf) > 0 {
			return si + "." + sf
		} else {
			return si
		}
	}
}

// ParseAmount parse the amount from the float string
func ParseAmount(str string) (*Amount, error) {
	ls := strings.SplitN(str, ".", 2)
	switch len(ls) {
	case 1:
		bi, ok := big.NewInt(0).SetString(ls[0], 10)
		if !ok {
			return nil, errors.New("not parsable")
		}
		pi := &Amount{Int: bi}
		return pi.MulC(FractionalMax), nil
	case 2:
		bi, ok := big.NewInt(0).SetString(ls[0], 10)
		if !ok {
			return nil, errors.New("not parsable")
		}
		pi := &Amount{Int: bi}
		bf, ok := big.NewInt(0).SetString(padFractional(ls[1]), 10)
		if !ok {
			return nil, errors.New("not parsable")
		}
		pf := &Amount{Int: bf}
		return pi.MulC(FractionalMax).Add(pf), nil
	default:
		return nil, errors.WithStack(ErrInvalidAmountFormat)
	}
}

// MustParseAmount parse the amount from the float string
func MustParseAmount(str string) *Amount {
	am, err := ParseAmount(str)
	if err != nil {
		panic(err)
	}
	return am
}
