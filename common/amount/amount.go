package amount

import (
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/fletaio/fleta/encoding"
)

// COIN is 1 coin
var COIN = NewCoinAmount(1, 0)

// FractionalMax represent the max value of under the float point
const FractionalMax = 1000000000000000000

// FractionalCount represent the number of under the float point
const FractionalCount = 18

func init() {
	if math.Pow10(FractionalCount) != FractionalMax {
		panic("Pow10(FractionalCount) is different with FractionalMax")
	}
	encoding.Register(Amount{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Amount)
		if err := enc.EncodeBytes(item.Bytes()); err != nil {
			return err
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		bs, err := dec.DecodeBytes()
		if err != nil {
			return err
		}
		item := NewAmountFromBytes(bs)
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
	encoding.Register(big.Int{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		bi := rv.Interface().(big.Int)
		if err := enc.EncodeBytes(bi.Bytes()); err != nil {
			return err
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		bs, err := dec.DecodeBytes()
		if err != nil {
			return err
		}
		bi := big.NewInt(0)
		bi.SetBytes(bs)
		rv.Set(reflect.ValueOf(bi).Elem())
		return nil
	})
}

var zeroInt = big.NewInt(0)

// Amount is the precision float value based on the big.Int
type Amount struct {
	*big.Int
}

func newAmount(value int64) *Amount {
	return &Amount{
		Int: big.NewInt(value),
	}
}

// NewCoinAmount returns the amount that is consisted of the integer and the fractional value
func NewCoinAmount(i uint64, f uint64) *Amount {
	if i == 0 {
		return newAmount(int64(f))
	} else if f == 0 {
		bi := newAmount(int64(i))
		return bi.MulC(FractionalMax)
	} else {
		bi := newAmount(int64(i))
		bf := newAmount(int64(f))
		return bi.MulC(FractionalMax).Add(bf)
	}
}

// NewAmountFromBytes parse the amount from the byte array
func NewAmountFromBytes(bs []byte) *Amount {
	b := newAmount(0)
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
		return ErrInvalidAmountFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidAmountFormat
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
	c := newAmount(0)
	c.Int.Add(am.Int, zeroInt)
	return c
}

// Add returns a + b (*immutable)
func (am *Amount) Add(b *Amount) *Amount {
	c := newAmount(0)
	c.Int.Add(am.Int, b.Int)
	return c
}

// Sub returns a - b (*immutable)
func (am *Amount) Sub(b *Amount) *Amount {
	c := newAmount(0)
	c.Int.Sub(am.Int, b.Int)
	return c
}

// Div returns a / b (*immutable)
func (am *Amount) Div(b *Amount) *Amount {
	c := newAmount(0)
	c.Int.Div(am.Int, b.Int)
	return c
}

// DivC returns a / b (*immutable)
func (am *Amount) DivC(b int64) *Amount {
	c := newAmount(0)
	c.Int.Div(am.Int, big.NewInt(b))
	return c
}

// Mul returns a * b (*immutable)
func (am *Amount) Mul(b *Amount) *Amount {
	c := newAmount(0)
	c.Int.Mul(am.Int, b.Int)
	return c
}

// MulC returns a * b (*immutable)
func (am *Amount) MulC(b int64) *Amount {
	c := newAmount(0)
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
		pi, err := strconv.ParseUint(ls[0], 10, 64)
		if err != nil {
			return nil, ErrInvalidAmountFormat
		}
		return NewCoinAmount(pi, 0), nil
	case 2:
		pi, err := strconv.ParseUint(ls[0], 10, 64)
		if err != nil {
			return nil, ErrInvalidAmountFormat
		}
		pf, err := strconv.ParseUint(padFractional(ls[1]), 10, 64)
		if err != nil {
			return nil, ErrInvalidAmountFormat
		}
		return NewCoinAmount(pi, pf), nil
	default:
		return nil, ErrInvalidAmountFormat
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
