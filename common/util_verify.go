package common

import (
	"crypto/elliptic"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common/hash"
)

// RecoverPubkey recover the public key using the hash value and the signature
func RecoverPubkey(chainid *big.Int, h hash.Hash256, s Signature) (PublicKey, error) {
	if len(s) < 64 {
		return PublicKey{}, errors.New("invalid signature size")
	}
	sig := make([]byte, 64)
	copy(sig[:], s[:64])
	vs := make([]byte, len(s[64:]))
	copy(vs, s[64:])

	ChainCap := GetChainCap(chainid)

	v := big.NewInt(0).SetBytes(vs)
	legacySignV := big.NewInt(0).SetInt64(28)
	if legacySignV.Cmp(v) <= 0 {
		v.Sub(v, ChainCap)
	}

	var vb byte
	if v.Int64() == 0 {
		vb = 0
	} else if v.Int64() == 1 {
		vb = 1
	} else {
		return PublicKey{}, errors.New("invalid recid")
	}

	bs, err := crypto.Ecrecover(h[:], append(sig, vb))
	// if err != nil && errors.Cause(err) == secp256k1.ErrInvalidRecoveryID {
	// 	s[64] = s[64] - 27
	// 	bs, err = ecrypto.Ecrecover(h[:], s[:])
	// 	if err != nil {
	// 		return PublicKey{}, errors.WithStack(err)
	// 	}
	// }
	if err != nil {
		return PublicKey{}, errors.WithStack(err)
	}
	X, Y := elliptic.Unmarshal(crypto.S256(), bs)
	var pubkey PublicKey
	copy(pubkey[:], elliptic.Marshal(crypto.S256(), X, Y))
	return pubkey, nil
}

// VerifySignature checks the signature with the public key and the hash value
func VerifySignature(pubkey PublicKey, h hash.Hash256, sig Signature) error {
	if !crypto.VerifySignature(pubkey[:], h[:], sig[:64]) {
		return errors.WithStack(ErrInvalidSignature)
	}
	return nil
}

// ValidateSignaturesMajority validates signatures with the signed hash and checks majority
func ValidateSignaturesMajority(chainID *big.Int, signedHash hash.Hash256, sigs []Signature, KeyMap map[PublicKey]bool) error {
	if len(sigs) != len(KeyMap)/2+1 {
		return errors.WithStack(ErrInsufficientSignature)
	}
	sigMap := map[Address]bool{}
	for _, sig := range sigs {
		pubkey, err := RecoverPubkey(chainID, signedHash, sig)
		if err != nil {
			return err
		}
		sigMap[pubkey.Address()] = true
		if !KeyMap[pubkey] {
			return errors.WithStack(ErrInvalidSignature)
		}
	}
	if len(sigMap) != len(sigs) {
		return errors.WithStack(ErrDuplicatedSignature)
	}
	return nil
}
