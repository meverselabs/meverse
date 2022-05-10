package common

import (
	"encoding/hex"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// PublicKeySize is 65 bytes
const PublicKeySize = 65

// PublicKey is the [PublicKeySize]byte with methods
type PublicKey [PublicKeySize]byte

// MarshalJSON is a marshaler function
func (pubkey PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pubkey.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (pubkey *PublicKey) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return errors.WithStack(ErrInvalidPublicKeyFormat)
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return errors.WithStack(ErrInvalidPublicKeyFormat)
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

// String returns the hex string of the public key
func (pubkey PublicKey) Address() Address {
	h := hash.Hash(pubkey[1:])
	return ecommon.BytesToAddress(h[12:])
}

// Clone returns the clonend value of it
func (pubkey PublicKey) Clone() PublicKey {
	var cp PublicKey
	copy(cp[:], pubkey[:])
	return cp
}

// ParsePublicKey parse the public hash from the string
func ParsePublicKey(str string) (PublicKey, error) {
	if len(str) != PublicKeySize*2 {
		return PublicKey{}, errors.WithStack(ErrInvalidPublicKeyFormat)
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
