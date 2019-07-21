package hash

import (
	"encoding/hex"
)

// Hash256Size is 32 bytes
const Hash256Size = 32

// Hash256 is the [Hash256Size]byte with methods
type Hash256 [Hash256Size]byte

// MarshalJSON is a marshaler function
func (hash Hash256) MarshalJSON() ([]byte, error) {
	return []byte(`"` + hash.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (hash *Hash256) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidHashFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidHashFormat
	}
	v, err := ParseHash(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(hash[:], v[:])
	return nil
}

// String returns the hex string of the hash
func (hash Hash256) String() string {
	return hex.EncodeToString(hash[:])
}

// ParseHash parse the hash from the string
func ParseHash(str string) (Hash256, error) {
	if len(str) != Hash256Size*2 {
		return Hash256{}, ErrInvalidHashFormat
	}
	bs, err := hex.DecodeString(str)
	if err != nil {
		return Hash256{}, err
	}
	if len(bs) != Hash256Size {
		return Hash256{}, ErrInvalidHashSize
	}
	var hash Hash256
	copy(hash[:], bs)
	return hash, nil
}

// MustParseHash panic when error occurred
func MustParseHash(str string) Hash256 {
	h, err := ParseHash(str)
	if err != nil {
		panic(err)
	}
	return h
}
