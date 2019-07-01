package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	ecrypto "github.com/fletaio/fleta/common/crypto/ethereum/crypto"
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return ecrypto.Ecrecover(hash, sig)
}

// VerifySignature checks that the given public key created signature over hash.
// The public key should be in compressed (33 bytes) or uncompressed (65 bytes) format.
// The signature should have the 64 byte [R || S] format.
func VerifySignature(pubkey, hash, signature []byte) bool {
	return ecrypto.VerifySignature(pubkey, hash, signature)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return ecrypto.CompressPubkey(pubkey)
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return ecrypto.S256()
}

// GenerateKey generate a key
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecrypto.GenerateKey()
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return ecrypto.Sign(hash, prv)
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	return ecrypto.Keccak256(data...)
}
