package bin

import (
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
)

type SumWriter struct {
	sum int64
}

func NewSumWriter() *SumWriter {
	return &SumWriter{
		sum: 0,
	}
}

func (sw *SumWriter) Uint8(w io.Writer, v uint8) (int64, error) {
	if n, err := WriteUint8(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Uint16(w io.Writer, v uint16) (int64, error) {
	if n, err := WriteUint16(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Uint32(w io.Writer, v uint32) (int64, error) {
	if n, err := WriteUint32(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Uint64(w io.Writer, v uint64) (int64, error) {
	if n, err := WriteUint64(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Bytes(w io.Writer, v []byte) (int64, error) {
	if n, err := WriteBytes(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) String(w io.Writer, v string) (int64, error) {
	if n, err := WriteString(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Bool(w io.Writer, v bool) (int64, error) {
	if n, err := WriteBool(w, v); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Hash256(w io.Writer, v hash.Hash256) (int64, error) {
	if n, err := WriteBytes(w, v.Bytes()); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Signature(w io.Writer, v common.Signature) (int64, error) {
	if n, err := WriteBytes(w, v[:]); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Address(w io.Writer, v common.Address) (int64, error) {
	if n, err := WriteBytes(w, v[:]); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) PublicKey(w io.Writer, v common.PublicKey) (int64, error) {
	if n, err := WriteBytes(w, v[:]); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Amount(w io.Writer, v *amount.Amount) (int64, error) {
	if n, err := WriteBytes(w, v.Bytes()); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) BigInt(w io.Writer, v *big.Int) (int64, error) {
	bs := []byte{}
	if v != nil {
		bs = v.Bytes()
	}
	if n, err := WriteBytes(w, bs); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) WriterTo(w io.Writer, v io.WriterTo) (int64, error) {
	if n, err := v.WriteTo(w); err != nil {
		return sw.sum, err
	} else {
		sw.sum += int64(n)
		return sw.sum, nil
	}
}

func (sw *SumWriter) Sum() int64 {
	return sw.sum
}
