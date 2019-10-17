package binutil

import (
	"encoding/binary"
)

// LittleEndian is the little-endian implementation of ByteOrder.
var LittleEndian littleEndian

// BigEndian is the big-endian implementation of ByteOrder.
var BigEndian bigEndian

type littleEndian struct{}

func (littleEndian) Uint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

func (littleEndian) PutUint16(b []byte, v uint16) {
	binary.LittleEndian.PutUint16(b, v)
}

func (littleEndian) Uint16ToBytes(v uint16) []byte {
	BNum := make([]byte, 2)
	binary.LittleEndian.PutUint16(BNum, v)
	return BNum
}

func (littleEndian) Uint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func (littleEndian) PutUint32(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
}

func (littleEndian) Uint32ToBytes(v uint32) []byte {
	BNum := make([]byte, 4)
	binary.LittleEndian.PutUint32(BNum, v)
	return BNum
}

func (littleEndian) Uint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func (littleEndian) PutUint64(b []byte, v uint64) {
	binary.LittleEndian.PutUint64(b, v)
}

func (littleEndian) Uint64ToBytes(v uint64) []byte {
	BNum := make([]byte, 8)
	binary.LittleEndian.PutUint64(BNum, v)
	return BNum
}

type bigEndian struct{}

func (bigEndian) Uint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func (bigEndian) PutUint16(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

func (bigEndian) Uint16ToBytes(v uint16) []byte {
	BNum := make([]byte, 2)
	binary.BigEndian.PutUint16(BNum, v)
	return BNum
}

func (bigEndian) Uint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func (bigEndian) PutUint32(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

func (bigEndian) Uint32ToBytes(v uint32) []byte {
	BNum := make([]byte, 4)
	binary.BigEndian.PutUint32(BNum, v)
	return BNum
}

func (bigEndian) Uint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func (bigEndian) PutUint64(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}

func (bigEndian) Uint64ToBytes(v uint64) []byte {
	BNum := make([]byte, 8)
	binary.BigEndian.PutUint64(BNum, v)
	return BNum
}
