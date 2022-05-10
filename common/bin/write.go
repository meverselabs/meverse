package bin

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// WriteUint64 writes the uint64 number to the writer
func WriteUint64(w io.Writer, num uint64) (int64, error) {
	BNum := make([]byte, 8)
	defer func() {
		BNum = nil
	}()
	binary.LittleEndian.PutUint64(BNum, num)
	if n, err := w.Write(BNum); err != nil {
		return int64(n), errors.WithStack(err)
	} else if n != 8 {
		return int64(n), errors.WithStack(ErrInvalidLength)
	} else {
		return 8, nil
	}
}

// WriteUint32 writes the uint32 number to the writer
func WriteUint32(w io.Writer, num uint32) (int64, error) {
	BNum := make([]byte, 4)
	defer func() {
		BNum = nil
	}()
	binary.LittleEndian.PutUint32(BNum, num)
	if n, err := w.Write(BNum); err != nil {
		return int64(n), errors.WithStack(err)
	} else if n != 4 {
		return int64(n), errors.WithStack(ErrInvalidLength)
	} else {
		return 4, nil
	}
}

// WriteUint16 writes the uint16 number to the writer
func WriteUint16(w io.Writer, num uint16) (int64, error) {
	BNum := make([]byte, 2)
	defer func() {
		BNum = nil
	}()
	binary.LittleEndian.PutUint16(BNum, num)
	if n, err := w.Write(BNum); err != nil {
		return int64(n), errors.WithStack(err)
	} else if n != 2 {
		return int64(n), errors.WithStack(ErrInvalidLength)
	} else {
		return 2, nil
	}
}

// WriteUint8 writes the uint8 number to the writer
func WriteUint8(w io.Writer, num uint8) (int64, error) {
	if n, err := w.Write([]byte{byte(num)}); err != nil {
		return int64(n), errors.WithStack(err)
	} else if n != 1 {
		return int64(n), errors.WithStack(ErrInvalidLength)
	} else {
		return 1, nil
	}
}

// WriteBytes writes the byte array bytes with the var-length-bytes to the writer
func WriteBytes(w io.Writer, bs []byte) (int64, error) {
	var wrote int64
	if len(bs) < 254 {
		if n, err := WriteUint8(w, uint8(len(bs))); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := w.Write(bs); err != nil {
			return wrote, err
		} else {
			wrote += int64(n)
		}
	} else if len(bs) < 65536 {
		if n, err := WriteUint8(w, 254); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := WriteUint16(w, uint16(len(bs))); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := w.Write(bs); err != nil {
			return wrote, err
		} else {
			wrote += int64(n)
		}
	} else {
		if n, err := WriteUint8(w, 255); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := WriteUint32(w, uint32(len(bs))); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := w.Write(bs); err != nil {
			return wrote, err
		} else {
			wrote += int64(n)
		}
	}
	bs = nil
	return wrote, nil
}

// WriteString writes the string with the var-length-byte to the writer
func WriteString(w io.Writer, str string) (int64, error) {
	return WriteBytes(w, []byte(str))
}

// WriteBool writes the bool using a uint8 to the writer
func WriteBool(w io.Writer, b bool) (int64, error) {
	if b {
		return WriteUint8(w, 1)
	} else {
		return WriteUint8(w, 0)
	}
}

// WriterToBytes return bytes from writer to
func WriterToBytes(w io.WriterTo) ([]byte, int64, error) {
	var buffer bytes.Buffer
	if n, err := w.WriteTo(&buffer); err != nil {
		return nil, n, errors.WithStack(err)
	} else {
		return buffer.Bytes(), n, nil
	}
}

// WriterToHash return bytes from writer to
func WriterToHash(w io.WriterTo) (hash.Hash256, int64, error) {
	bs, n, err := WriterToBytes(w)
	if err != nil {
		return hash.Hash256{}, n, errors.WithStack(err)
	}
	return hash.Hash(bs), n, nil
}

// MustWriterToHash return bytes from writer to
func MustWriterToHash(w io.WriterTo) hash.Hash256 {
	bs, _, err := WriterToBytes(w)
	if err != nil {
		panic(err)
	}
	return hash.Hash(bs)
}
