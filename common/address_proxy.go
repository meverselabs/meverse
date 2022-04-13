package common

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Address = common.Address

var ZeroAddr = Address{}

// Lengths of hashes and addresses in bytes.
const (
	// AddressLength is the expected length of the address
	AddressLength = common.AddressLength
)

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	return common.BytesToAddress(b)
}

// BigToAddress returns Address with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAddress(b *big.Int) Address {
	return common.BigToAddress(b)
}

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Address {
	return common.HexToAddress(s)
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Ethereum address or not.
func IsHexAddress(s string) bool {
	return common.IsHexAddress(s)
}

// ParseAddress is parse address
func ParseAddress(s string) (common.Address, error) {
	if strings.Index(s, "0x") == 0 {
		s = s[2:]
	}
	h, err := hex.DecodeString(s)
	addr := common.Address{}
	if err != nil {
		return addr, errors.WithStack(err)
	}
	copy(addr[:], h[:])
	return addr, nil
}
