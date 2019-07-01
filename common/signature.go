package common

import (
	"encoding/hex"
)

// SignatureSize is 65 bytes
const SignatureSize = 65

// Signature is the [SignatureSize]byte with methods
type Signature [SignatureSize]byte

// MarshalJSON is a marshaler function
func (sig Signature) MarshalJSON() ([]byte, error) {
	return []byte(`"` + sig.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (sig *Signature) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidSignatureFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidSignatureFormat
	}
	v, err := ParseSignature(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(sig[:], v[:])
	return nil
}

// String returns the hex string of the signature
func (sig Signature) String() string {
	return hex.EncodeToString(sig[:])
}

// Clone returns the clonend value of it
func (sig Signature) Clone() Signature {
	var cp Signature
	copy(cp[:], sig[:])
	return cp
}

// ParseSignature parse the public hash from the string
func ParseSignature(str string) (Signature, error) {
	if len(str) != SignatureSize*2 {
		return Signature{}, ErrInvalidSignatureFormat
	}
	bs, err := hex.DecodeString(str)
	if err != nil {
		return Signature{}, err
	}
	var sig Signature
	copy(sig[:], bs)
	return sig, nil
}

// MustParseSignature panic when error occurred
func MustParseSignature(str string) Signature {
	sig, err := ParseSignature(str)
	if err != nil {
		panic(err)
	}
	return sig
}
