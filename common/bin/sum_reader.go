package bin

import (
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
)

type SumReader struct {
	sum int64
}

func NewSumReader() *SumReader {
	return &SumReader{
		sum: 0,
	}
}

func (sr *SumReader) Uint8(r io.Reader, p *uint8) (int64, error) {
	if v, n, err := ReadUint8(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) GetUint8(r io.Reader) (uint8, int64, error) {
	if v, n, err := ReadUint8(r); err != nil {
		return 0, sr.sum, err
	} else {
		sr.sum += int64(n)
		return v, sr.sum, err
	}
}

func (sr *SumReader) Uint16(r io.Reader, p *uint16) (int64, error) {
	if v, n, err := ReadUint16(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) GetUint16(r io.Reader) (uint16, int64, error) {
	if v, n, err := ReadUint16(r); err != nil {
		return 0, sr.sum, err
	} else {
		sr.sum += int64(n)
		return v, sr.sum, nil
	}
}

func (sr *SumReader) Uint32(r io.Reader, p *uint32) (int64, error) {
	if v, n, err := ReadUint32(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) GetUint32(r io.Reader) (uint32, int64, error) {
	if v, n, err := ReadUint32(r); err != nil {
		return 0, sr.sum, err
	} else {
		sr.sum += int64(n)
		return v, sr.sum, nil
	}
}

func (sr *SumReader) Uint64(r io.Reader, p *uint64) (int64, error) {
	if v, n, err := ReadUint64(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) GetUint64(r io.Reader) (uint64, int64, error) {
	if v, n, err := ReadUint64(r); err != nil {
		return 0, sr.sum, err
	} else {
		sr.sum += int64(n)
		return v, sr.sum, nil
	}
}

func (sr *SumReader) Bytes(r io.Reader, p *[]byte) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) String(r io.Reader, p *string) (int64, error) {
	if v, n, err := ReadString(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) Bool(r io.Reader, p *bool) (int64, error) {
	if v, n, err := ReadBool(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) Hash256(r io.Reader, p *hash.Hash256) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p).SetBytes(v)
		return sr.sum, nil
	}
}

func (sr *SumReader) Signature(r io.Reader, p *common.Signature) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = v
		return sr.sum, nil
	}
}

func (sr *SumReader) Address(r io.Reader, p *common.Address) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		copy((*p)[:], v)
		return sr.sum, nil
	}
}

func (sr *SumReader) PublicKey(r io.Reader, p *common.PublicKey) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		copy((*p)[:], v)
		return sr.sum, nil
	}
}

func (sr *SumReader) Amount(r io.Reader, p **amount.Amount) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = amount.NewAmountFromBytes(v)
		return sr.sum, nil
	}
}

func (sr *SumReader) BigInt(r io.Reader, p **big.Int) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		(*p) = big.NewInt(0).SetBytes(v)
		return sr.sum, nil
	}
}

func (sr *SumReader) ReaderFrom(r io.Reader, p io.ReaderFrom) (int64, error) {
	if n, err := p.ReadFrom(r); err != nil {
		return sr.sum, err
	} else {
		sr.sum += int64(n)
		return sr.sum, nil
	}
}

func (sr *SumReader) Sum() int64 {
	return sr.sum
}
