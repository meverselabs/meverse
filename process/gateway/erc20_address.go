package gateway

import (
	"encoding/hex"
	"strings"
)

// TokenAddressSize is 20 bytes
const TokenAddressSize = 20

// TokenAddress is the [TokenAddressSize]byte with methods
type TokenAddress [TokenAddressSize]byte

// MarshalJSON is a marshaler function
func (addr TokenAddress) MarshalJSON() ([]byte, error) {
	return []byte(`"` + addr.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (addr *TokenAddress) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidTokenAddressFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidTokenAddressFormat
	}
	v, err := ParseTokenAddress(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(addr[:], v[:])
	return nil
}

// String returns a base58 value of the Token address
func (addr TokenAddress) String() string {
	return "0x" + hex.EncodeToString(addr[:])
}

// ParseTokenAddress parse the Token address from the string
func ParseTokenAddress(str string) (TokenAddress, error) {
	if !strings.HasPrefix(str, "0x") {
		return TokenAddress{}, ErrInvalidTokenAddressFormat
	}
	if len(str) != TokenAddressSize*2+2 {
		return TokenAddress{}, ErrInvalidTokenAddressFormat
	}
	bs, err := hex.DecodeString(str[2:])
	if err != nil {
		return TokenAddress{}, err
	}
	if len(bs) != TokenAddressSize {
		return TokenAddress{}, ErrInvalidTokenAddressFormat
	}
	var addr TokenAddress
	copy(addr[:], bs)
	return addr, nil
}

// MustParseTokenAddress panic when error occurred
func MustParseTokenAddress(str string) TokenAddress {
	addr, err := ParseTokenAddress(str)
	if err != nil {
		panic(err)
	}
	return addr
}
