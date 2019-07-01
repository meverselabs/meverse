package util

import (
	"encoding/binary"
)

// Uint64ToBytes returns a byte array of the uint64 number
func Uint64ToBytes(v uint64) []byte {
	BNum := make([]byte, 8)
	binary.LittleEndian.PutUint64(BNum, v)
	return BNum
}

// Uint32ToBytes returns a byte array of the uint32 number
func Uint32ToBytes(v uint32) []byte {
	BNum := make([]byte, 4)
	binary.LittleEndian.PutUint32(BNum, v)
	return BNum
}

// Uint16ToBytes returns a byte array of the uint16 number
func Uint16ToBytes(v uint16) []byte {
	BNum := make([]byte, 2)
	binary.LittleEndian.PutUint16(BNum, v)
	return BNum
}

// Uint48ToBytes returns a byte array of the uint32 number and the uint16 number
func Uint48ToBytes(a uint32, b uint16) []byte {
	BNum := make([]byte, 6)
	binary.LittleEndian.PutUint32(BNum, a)
	binary.LittleEndian.PutUint16(BNum[4:], b)
	return BNum
}

// BytesToUint16 returns a uint16 number of the byte array
func BytesToUint16(v []byte) uint16 {
	return binary.LittleEndian.Uint16(v)
}

// BytesToUint32 returns a uint32 number of the byte array
func BytesToUint32(v []byte) uint32 {
	return binary.LittleEndian.Uint32(v)
}

// BytesToUint64 returns a uint64 number of the byte array
func BytesToUint64(v []byte) uint64 {
	return binary.LittleEndian.Uint64(v)
}
