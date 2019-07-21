package gateway

import (
	"encoding/hex"
	"strings"
)

// ERC20AddressSize is 20 bytes
const ERC20AddressSize = 20

// ERC20Address is the [ERC20AddressSize]byte with methods
type ERC20Address [ERC20AddressSize]byte

// MarshalJSON is a marshaler function
func (addr ERC20Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + addr.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (addr *ERC20Address) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidERC20AddressFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidERC20AddressFormat
	}
	v, err := ParseERC20Address(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(addr[:], v[:])
	return nil
}

// String returns a base58 value of the ERC20 address
func (addr ERC20Address) String() string {
	return "0x" + hex.EncodeToString(addr[:])
}

// ParseERC20Address parse the ERC20 address from the string
func ParseERC20Address(str string) (ERC20Address, error) {
	if !strings.HasPrefix(str, "0x") {
		return ERC20Address{}, ErrInvalidERC20AddressFormat
	}
	if len(str) != ERC20AddressSize*2+2 {
		return ERC20Address{}, ErrInvalidERC20AddressFormat
	}
	bs, err := hex.DecodeString(str[2:])
	if err != nil {
		return ERC20Address{}, err
	}
	if len(bs) != ERC20AddressSize {
		return ERC20Address{}, ErrInvalidERC20AddressFormat
	}
	var addr ERC20Address
	copy(addr[:], bs)
	return addr, nil
}

// MustParseERC20Address panic when error occurred
func MustParseERC20Address(str string) ERC20Address {
	addr, err := ParseERC20Address(str)
	if err != nil {
		panic(err)
	}
	return addr
}
