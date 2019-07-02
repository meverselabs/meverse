package common

import (
	"bytes"
	"encoding/binary"

	"github.com/fletaio/fleta/common/util"

	"github.com/mr-tron/base58/base58"
)

// AddressSize is 14 bytes
const AddressSize = 14

// Address is the [AddressSize]byte with methods
type Address [AddressSize]byte

// NewAddress returns a Address by the AccountCoordinate, by the nonce
func NewAddress(height uint32, index uint16, nonce uint64) Address {
	var addr Address
	copy(addr[:], util.Uint32ToBytes(height))
	copy(addr[4:], util.Uint16ToBytes(index))
	if nonce > 0 {
		copy(addr[6:], util.Uint64ToBytes(nonce))
	}
	return addr
}

// MarshalJSON is a marshaler function
func (addr Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + addr.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (addr *Address) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidAddressFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidAddressFormat
	}
	v, err := ParseAddress(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(addr[:], v[:])
	return nil
}

// String returns a base58 value of the address
func (addr Address) String() string {
	var bs []byte
	checksum := addr.Checksum()
	result := bytes.TrimRight(addr[:], string([]byte{0}))
	if len(result) < 7 {
		bs = make([]byte, 7)
		copy(bs[1:], result[:])
	} else if len(result) < 13 {
		bs = make([]byte, 13)
		copy(bs[1:], result[:])
	}
	bs[0] = checksum

	return base58.Encode(bs)
}

// Clone returns the clonend value of it
func (addr Address) Clone() Address {
	var cp Address
	copy(cp[:], addr[:])
	return cp
}

// WithNonce returns derive the address using the nonce
func (addr Address) WithNonce(nonce uint64) Address {
	var cp Address
	copy(cp[:], addr[:])
	binary.LittleEndian.PutUint64(cp[6:], nonce)
	return cp
}

// Checksum returns the checksum byte
func (addr Address) Checksum() byte {
	var cs byte
	for _, c := range addr {
		cs = cs ^ c
	}
	return cs
}

// Coordinate returns the coordinate of the address
func (addr Address) Coordinate() *Coordinate {
	var coord Coordinate
	coord.SetBytes(addr[:CoordinateSize])
	return &coord
}

// ParseAddress parse the address from the string
func ParseAddress(str string) (Address, error) {
	bs, err := base58.Decode(str)
	if err != nil {
		return Address{}, err
	}
	if len(bs) != 7 && len(bs) != 13 {
		return Address{}, ErrInvalidAddressFormat
	}
	cs := bs[0]
	var addr Address
	copy(addr[:], bs[1:])
	if cs != addr.Checksum() {
		return Address{}, ErrInvalidAddressCheckSum
	}
	return addr, nil
}

// MustParseAddress panic when error occurred
func MustParseAddress(str string) Address {
	addr, err := ParseAddress(str)
	if err != nil {
		panic(err)
	}
	return addr
}
