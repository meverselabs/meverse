package bin

import (
	"encoding/binary"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
)

// Uint16Bytes returns a byte array of the uint16 number
func Uint16Bytes(v uint16) []byte {
	BNum := make([]byte, 2)
	binary.LittleEndian.PutUint16(BNum, v)
	return BNum
}

// Uint32Bytes returns a byte array of the uint32 number
func Uint32Bytes(v uint32) []byte {
	BNum := make([]byte, 4)
	binary.LittleEndian.PutUint32(BNum, v)
	return BNum
}

// Uint64Bytes returns a byte array of the uint64 number
func Uint64Bytes(v uint64) []byte {
	BNum := make([]byte, 8)
	binary.LittleEndian.PutUint64(BNum, v)
	return BNum
}

// PutUint16 returns a byte array of the uint16 number
func PutUint16(bs []byte, v uint16) {
	binary.LittleEndian.PutUint16(bs, v)
}

// Uint32 returns a byte array of the uint32 number
func PutUint32(bs []byte, v uint32) {
	binary.LittleEndian.PutUint32(bs, v)
}

// Uint64 returns a byte array of the uint64 number
func PutUint64(bs []byte, v uint64) {
	binary.LittleEndian.PutUint64(bs, v)
}

// Uint16 returns a uint16 number of the byte array
func Uint16(v []byte) uint16 {
	return binary.LittleEndian.Uint16(v)
}

// Uint32 returns a uint32 number of the byte array
func Uint32(v []byte) uint32 {
	return binary.LittleEndian.Uint32(v)
}

// Uint64 returns a uint64 number of the byte array
func Uint64(v []byte) uint64 {
	return binary.LittleEndian.Uint64(v)
}

// Amount returns a Amount of the byte array
func Amount(v []byte) *amount.Amount {
	return &amount.Amount{Int: big.NewInt(0).SetBytes(v)}
}

// Amount returns a Amount of the byte array
func Address(v []byte) common.Address {
	bi := big.NewInt(0).SetBytes(v)
	return common.BigToAddress(bi)
}
