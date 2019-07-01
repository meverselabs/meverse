package common

import (
	"encoding/binary"
	"encoding/hex"
	"reflect"

	"github.com/fletaio/fleta/encoding"
)

// CoordinateSize is 6 bytes
const CoordinateSize = 6

func init() {
	encoding.Register(Coordinate{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Coordinate)
		if err := enc.EncodeUint64(item.ID()); err != nil {
			return err
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		v, err := dec.DecodeUint64()
		if err != nil {
			return err
		}
		item := NewCoordinateByID(v)
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

// Coordinate is (BlockHeight, TransactionIndexOfTheBlock)
type Coordinate struct {
	Height uint32
	Index  uint16
}

// NewCoordinate returns a Coordinate
func NewCoordinate(Height uint32, Index uint16) *Coordinate {
	return &Coordinate{
		Height: Height,
		Index:  Index,
	}
}

// NewCoordinateByID returns a Coordinate using compacted id
func NewCoordinateByID(id uint64) *Coordinate {
	return &Coordinate{
		Height: uint32(id >> 32),
		Index:  uint16(id >> 16),
	}
}

// MarshalJSON is a marshaler function
func (crd *Coordinate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + crd.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (crd *Coordinate) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidCoordinateFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidCoordinateFormat
	}
	v, err := ParseCoordinate(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	crd.Height = v.Height
	crd.Index = v.Index
	return nil
}

// Equal checks that two values is same or not
func (crd *Coordinate) Equal(b *Coordinate) bool {
	return crd.Height == b.Height && crd.Index == b.Index
}

// Clone returns the clonend value of it
func (crd *Coordinate) Clone() *Coordinate {
	return &Coordinate{
		Height: crd.Height,
		Index:  crd.Index,
	}
}

// Bytes returns a byte array
func (crd *Coordinate) Bytes() []byte {
	bs := make([]byte, CoordinateSize)
	binary.LittleEndian.PutUint32(bs, crd.Height)
	binary.LittleEndian.PutUint16(bs[4:], crd.Index)
	return bs
}

// SetBytes updates the coordinate using given bytes
func (crd *Coordinate) SetBytes(bs []byte) error {
	if len(bs) != CoordinateSize {
		return ErrInvalidCoordinateBytesLength
	}
	crd.Height = binary.LittleEndian.Uint32(bs)
	crd.Index = binary.LittleEndian.Uint16(bs[4:])
	return nil
}

// ID returns a compacted id
func (crd *Coordinate) ID() uint64 {
	return uint64(crd.Height)<<32 | uint64(crd.Index)<<16
}

// String returns a hex value of the byte array
func (crd *Coordinate) String() string {
	return hex.EncodeToString(crd.Bytes())
}

// ParseCoordinate parse the public hash from the string
func ParseCoordinate(str string) (*Coordinate, error) {
	if len(str) != CoordinateSize*2 {
		return nil, ErrInvalidCoordinateFormat
	}
	bs, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	coord := &Coordinate{
		Height: binary.LittleEndian.Uint32(bs),
		Index:  binary.LittleEndian.Uint16(bs[4:]),
	}
	return coord, nil
}

// MustParseCoordinate panic when error occurred
func MustParseCoordinate(str string) *Coordinate {
	coord, err := ParseCoordinate(str)
	if err != nil {
		panic(err)
	}
	return coord
}
