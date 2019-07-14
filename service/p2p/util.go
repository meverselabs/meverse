package p2p

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/encoding"
)

// ReadUint64 reads a uint64 number from the reader
func ReadUint64(r io.Reader) (uint64, int64, error) {
	var read int64
	BNum := make([]byte, 8)
	if n, err := FillBytes(r, BNum); err != nil {
		return 0, read, err
	} else {
		read += n
	}
	return binary.LittleEndian.Uint64(BNum), int64(read), nil
}

// ReadUint32 reads a uint32 number from the reader
func ReadUint32(r io.Reader) (uint32, int64, error) {
	var read int64
	BNum := make([]byte, 4)
	if n, err := FillBytes(r, BNum); err != nil {
		return 0, read, err
	} else {
		read += n
	}
	return binary.LittleEndian.Uint32(BNum), int64(read), nil
}

// ReadUint16 reads a uint16 number from the reader
func ReadUint16(r io.Reader) (uint16, int64, error) {
	var read int64
	BNum := make([]byte, 2)
	if n, err := FillBytes(r, BNum); err != nil {
		return 0, read, err
	} else {
		read += n
	}
	return binary.LittleEndian.Uint16(BNum), int64(read), nil
}

// ReadUint8 reads a uint8 number from the reader
func ReadUint8(r io.Reader) (uint8, int64, error) {
	var read int64
	BNum := make([]byte, 1)
	if n, err := FillBytes(r, BNum); err != nil {
		return 0, read, err
	} else {
		read += n
	}
	return uint8(BNum[0]), int64(read), nil
}

// ReadBytes reads a byte array from the reader
func ReadBytes(r io.Reader) ([]byte, int64, error) {
	var bs []byte
	var read int64
	if Len, n, err := ReadUint8(r); err != nil {
		return nil, read, err
	} else if Len < 254 {
		read += n
		bs = make([]byte, Len)
		if n, err := FillBytes(r, bs); err != nil {
			return nil, read, err
		} else {
			read += n
		}
		return bs, read, nil
	} else if Len == 254 {
		if Len, n, err := ReadUint16(r); err != nil {
			return nil, read, err
		} else {
			read += n
			bs = make([]byte, Len)
			if n, err := FillBytes(r, bs); err != nil {
				return nil, read, err
			} else {
				read += n
			}
		}
		return bs, read, nil
	} else {
		if Len, n, err := ReadUint32(r); err != nil {
			return nil, read, err
		} else {
			read += n
			bs = make([]byte, Len)
			if n, err := FillBytes(r, bs); err != nil {
				return nil, read, err
			} else {
				read += n
			}
		}
		return bs, read, nil
	}
}

// ReadString reads a string array from the reader
func ReadString(r io.Reader) (string, int64, error) {
	if bs, n, err := ReadBytes(r); err != nil {
		return "", n, err
	} else {
		return string(bs), n, err
	}
}

// ReadBool reads a bool using a uint8 from the reader
func ReadBool(r io.Reader) (bool, int64, error) {
	if v, n, err := ReadUint8(r); err != nil {
		return false, n, err
	} else {
		return (v == 1), n, err
	}
}

// FillBytes reads bytes from the reader until the given bytes array is filled
func FillBytes(r io.Reader, bs []byte) (int64, error) {
	var read int
	for read < len(bs) {
		if n, err := r.Read(bs[read:]); err != nil {
			return int64(read), err
		} else {
			read += n
			if read >= len(bs) {
				break
			}
			if n <= 0 {
				return int64(read), ErrInvalidLength
			}
		}
	}
	if read != len(bs) {
		return int64(read), ErrInvalidLength
	}
	return int64(read), nil
}

// MessageToBytes returns bytes to the message
func MessageToBytes(m interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	fc := encoding.Factory("message")
	t, err := fc.TypeOf(m)
	if err != nil {
		return nil, err
	}
	buffer.Write(util.Uint16ToBytes(t))
	enc := encoding.NewEncoder(&buffer)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
