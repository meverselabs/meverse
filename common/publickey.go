package common

import (
	"encoding/hex"
)

// PublicKeySize is 33 bytes
const PublicKeySize = 33

// PublicKey is the [PublicKeySize]byte with methods
type PublicKey [PublicKeySize]byte

// MarshalJSON is a marshaler function
func (pubkey PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pubkey.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (pubkey *PublicKey) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidPublicKeyFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidPublicKeyFormat
	}
	v, err := ParsePublicKey(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(pubkey[:], v[:])
	return nil
}

// String returns the hex string of the public key
func (pubkey PublicKey) String() string {
	return hex.EncodeToString(pubkey[:])
}

// Clone returns the clonend value of it
func (pubkey PublicKey) Clone() PublicKey {
	var cp PublicKey
	copy(cp[:], pubkey[:])
	return cp
}

// Checksum returns the checksum byte
func (pubkey PublicKey) Checksum() byte {
	var cs byte
	for i := 0; i < len(pubkey); i++ {
		cs = cs ^ pubkey[i]
	}
	return cs
}

// ParsePublicKey parse the public hash from the string
func ParsePublicKey(str string) (PublicKey, error) {
	if len(str) != PublicKeySize*2 {
		return PublicKey{}, ErrInvalidPublicKeyFormat
	}
	bs, err := hex.DecodeString(str)
	if err != nil {
		return PublicKey{}, err
	}
	var pubkey PublicKey
	copy(pubkey[:], bs)
	return pubkey, nil
}

// MustParsePublicKey panic when error occurred
func MustParsePublicKey(str string) PublicKey {
	pubkey, err := ParsePublicKey(str)
	if err != nil {
		panic(err)
	}
	return pubkey
}
