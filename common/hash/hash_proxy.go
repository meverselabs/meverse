package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
)

type Hash256 = ecommon.Hash

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = ecommon.HashLength
)

// BigToHash sets byte representation of b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BigToHash(b *big.Int) Hash256 {
	return Hash256(ecommon.BigToHash(b))
}

// HexToHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToHash(s string) Hash256 {
	return Hash256(ecommon.HexToHash(s))
}

// Hash calculates and returns the Hash hash of the input data.
func Hash(data ...[]byte) Hash256 {
	return Hash256(ecrypto.Keccak256Hash(data...))
}

func HashForPubKey(data []byte) Hash256 {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		panic(err)
	}
	bs := h.Sum(nil)
	var hash Hash256
	copy(hash[:], bs)
	return hash
}

// Uint64 calculates and returns uint64 from the Hash hash of the input data.
func Uint64(data ...[]byte) uint64 {
	h := Hash256(ecrypto.Keccak256Hash(data...))
	return binary.LittleEndian.Uint64(h[:])
}

// DoubleHash hashes twice
func DoubleHash(data ...[]byte) Hash256 {
	h := Hash(data...)
	return Hash(h[:])
}

func DoubleHashForPubKey(data []byte) Hash256 {
	h := HashForPubKey(data)
	return HashForPubKey(h[:])
}

// Hashes returns the result of Hash(h1+'h'+...)
func Hashes(hs ...Hash256) Hash256 {
	data := make([]byte, (HashLength+1)*len(hs)-1)
	idx := 0
	for i, h := range hs {
		copy(data, h[:])
		idx += HashLength
		if i < len(hs)-1 {
			data[idx] = 'h'
			idx++
		}
	}
	return Hash(data)
}
