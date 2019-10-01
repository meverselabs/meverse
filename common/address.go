package common

import (
	"bytes"
	"encoding/binary"

	"github.com/mr-tron/base58/base58"
)

// AddressSize is 14 bytes
const AddressSize = 14

// Address is the [AddressSize]byte with methods
type Address [AddressSize]byte

// NewAddress returns a Address by the AccountCoordinate and the magic
func NewAddress(height uint32, index uint16, magic uint64) Address {
	var addr Address
	binary.BigEndian.PutUint32(addr[:], height)
	binary.BigEndian.PutUint16(addr[4:], index)
	binary.BigEndian.PutUint64(addr[6:], magic)
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
	} else if len(result) < 15 {
		bs = make([]byte, 15)
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

// Checksum returns the checksum byte
func (addr Address) Checksum() byte {
	var cs byte
	for _, c := range addr {
		cs = cs ^ c
	}
	return cs
}

// Height returns the height of the address created
func (addr Address) Height() uint32 {
	return binary.BigEndian.Uint32(addr[:])
}

// Index returns the index of the address created
func (addr Address) Index() uint16 {
	return binary.BigEndian.Uint16(addr[4:])
}

// Nonce returns the nonce of the address created
func (addr Address) Nonce() uint64 {
	return binary.BigEndian.Uint64(addr[6:])
}

// ParseAddress parse the address from the string
func ParseAddress(str string) (Address, error) {
	bs, err := base58.Decode(str)
	if err != nil {
		return Address{}, err
	}
	if len(bs) != 7 && len(bs) != 15 {
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

// TickerUsageToMagicNumber convert a name and a usage to the magic number
func TickerUsageToMagicNumber(Ticker string, Usage string) uint64 {
	base := []byte{95, 65, 69, 84, 122, 64, 11, 43}
	Salt := "PoweredByFLETABlockchain"
	ls := Ticker + " " + Usage + " " + Salt
	cnt := len(ls) / 8
	if len(ls)%8 != 0 {
		cnt++
	}
	for i := 0; i < cnt; i++ {
		from := i * 8
		to := (i + 1) * 8
		if to > len(ls) {
			to = len(ls)
		}
		str := ls[from:to]
		for j := 0; j < 8; j++ {
			v := byte(0)
			if j < len(str) {
				v = byte(str[j])
			}
			base[j] = base[j] ^ v
		}
	}
	return binary.BigEndian.Uint64(base)
}
