package common

import (
	"encoding/hex"

	"github.com/pkg/errors"
)

// SignatureSize is 67 bytes r 32 s 32 v 1 => v 3
// const SignatureSize = 67
const MinSignatureSize = 65

// Signature is the [SignatureSize]byte with methods
type Signature []byte

// MarshalJSON is a marshaler function
func (sig Signature) MarshalJSON() ([]byte, error) {
	return []byte(`"` + sig.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (sig *Signature) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return errors.WithStack(ErrInvalidSignatureFormat)
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return errors.WithStack(ErrInvalidSignatureFormat)
	}
	v, err := ParseSignature(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	*sig = append(*sig, v...)
	return nil
}

// String returns the hex string of the signature
func (sig Signature) String() string {
	return hex.EncodeToString(sig[:])
}

// Clone returns the clonend value of it
func (sig Signature) Clone() Signature {
	bs := make([]byte, len(sig))
	copy(bs, sig[:])
	return bs
}

// ParseSignature parse the public hash from the string
func ParseSignature(str string) (Signature, error) {
	if len(str) < MinSignatureSize*2 {
		return Signature{}, errors.WithStack(ErrInvalidSignatureFormat)
	}
	bs, err := hex.DecodeString(str)
	if err != nil {
		return Signature{}, errors.WithStack(err)
	}
	return bs, nil
}

// MustParseSignature panic when error occurred
func MustParseSignature(str string) Signature {
	sig, err := ParseSignature(str)
	if err != nil {
		panic(err)
	}
	return sig
}
