package bin

import (
	"bytes"
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

type TypeReader struct {
	sum int64
}

func TypeReadAll(bs []byte, count int) ([]interface{}, error) {
	if bs == nil && len(bs) == 0 {
		return []interface{}{}, nil
	}
	tr := &TypeReader{
		sum: 0,
	}
	data, _, err := tr.readAll(bytes.NewReader(bs))
	if count < 0 {
		return data, err
	}
	if len(data) < count {
		return nil, errors.Errorf("invalid output count less then, %v", count)
	}
	return data, err
}

func (tr *TypeReader) readAll(r io.Reader) ([]interface{}, int64, error) {
	var data []interface{}
	for {
		d, _, err := tr.read(r)
		if err != nil {
			if errors.Cause(err) == io.EOF {
				err = nil
			}
			return data, tr.sum, err
		}
		data = append(data, d)
	}
}

func (tr *TypeReader) read(r io.Reader) (interface{}, int64, error) {
	if v, n, err := tr.getByte(r); err != nil {
		tr.sum += n
		return nil, tr.sum, err
	} else {
		switch v {
		case tagUint8:
			return tr.getUint8(r)
		case tagUint16:
			return tr.getUint16(r)
		case tagUint32:
			return tr.getUint32(r)
		case tagUint64:
			return tr.getUint64(r)
		case tagBytes:
			var data []byte
			if _, err := tr.bytes(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagString:
			var data string
			if _, err := tr.string(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagBool:
			var data bool
			if _, err := tr.bool(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagHash256:
			var data hash.Hash256
			if _, err := tr.hash256(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagSignature:
			var data common.Signature
			if _, err := tr.signature(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagAddress:
			var data common.Address
			if _, err := tr.address(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagPublicKey:
			var data common.PublicKey
			if _, err := tr.publicKey(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagAmount:
			var data *amount.Amount
			if _, err := tr.amount(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagBigInt:
			var data *big.Int
			if _, err := tr.bigInt(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagAddressArr:
			var data []common.Address
			if _, err := tr.addrs(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		case tagSlice:
			data := []interface{}{}
			if Len, n, err := ReadUint8(r); err != nil {
				tr.sum += n
				return nil, tr.sum, err
			} else {
				for i := 0; i < int(Len); i++ {
					if inter, _, err := tr.read(r); err != nil {
						return nil, tr.sum, err
					} else {
						data = append(data, inter)
						tr.sum += int64(n)
					}
				}
			}
			return data, tr.sum, nil
		case tagAmountArr:
			var data []*amount.Amount
			if _, err := tr.amounts(r, &data); err != nil {
				return nil, tr.sum, errors.New("Transaction hash read Value error")
			}
			return data, tr.sum, nil
		}
	}
	return nil, tr.sum, errors.New("not defined tag")
}

func (tr *TypeReader) getByte(r io.Reader) (byte, int64, error) {
	BNum := make([]byte, 1)
	if n, err := FillBytes(r, BNum); err != nil {
		return 0, tr.sum, err
	} else {
		tr.sum += n
	}
	return BNum[0], tr.sum, nil
}

func (tr *TypeReader) uint8(r io.Reader, p *uint8) (int64, error) {
	if v, n, err := ReadUint8(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) getUint8(r io.Reader) (uint8, int64, error) {
	if v, n, err := ReadUint8(r); err != nil {
		return 0, tr.sum, err
	} else {
		tr.sum += int64(n)
		return v, tr.sum, err
	}
}

func (tr *TypeReader) uint16(r io.Reader, p *uint16) (int64, error) {
	if v, n, err := ReadUint16(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) getUint16(r io.Reader) (uint16, int64, error) {
	if v, n, err := ReadUint16(r); err != nil {
		return 0, tr.sum, err
	} else {
		tr.sum += int64(n)
		return v, tr.sum, err
	}
}

func (tr *TypeReader) uint32(r io.Reader, p *uint32) (int64, error) {
	if v, n, err := ReadUint32(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) getUint32(r io.Reader) (uint32, int64, error) {
	if v, n, err := ReadUint32(r); err != nil {
		return 0, tr.sum, err
	} else {
		tr.sum += int64(n)
		return v, tr.sum, err
	}
}

func (tr *TypeReader) uint64(r io.Reader, p *uint64) (int64, error) {
	if v, n, err := ReadUint64(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) getUint64(r io.Reader) (uint64, int64, error) {
	if v, n, err := ReadUint64(r); err != nil {
		return 0, tr.sum, err
	} else {
		tr.sum += int64(n)
		return v, tr.sum, err
	}
}

func (tr *TypeReader) bytes(r io.Reader, p *[]byte) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) string(r io.Reader, p *string) (int64, error) {
	if v, n, err := ReadString(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) bool(r io.Reader, p *bool) (int64, error) {
	if v, n, err := ReadBool(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, err
	}
}

func (tr *TypeReader) hash256(r io.Reader, p *hash.Hash256) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p).SetBytes(v)
		return tr.sum, nil
	}
}

func (tr *TypeReader) signature(r io.Reader, p *common.Signature) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = v
		return tr.sum, nil
	}
}

func (tr *TypeReader) address(r io.Reader, p *common.Address) (int64, error) {
	bs := make([]byte, len(*p))
	if n, err := FillBytes(r, bs); err != nil {
		return tr.sum, err
	} else {
		tr.sum += n
		copy((*p)[:], bs)
		return tr.sum, nil
	}
}

func (tr *TypeReader) publicKey(r io.Reader, p *common.PublicKey) (int64, error) {
	bs := make([]byte, len(*p))
	if n, err := FillBytes(r, bs); err != nil {
		return tr.sum, err
	} else {
		tr.sum += n
		copy((*p)[:], bs)
		return tr.sum, nil
	}
}

func (tr *TypeReader) amount(r io.Reader, p **amount.Amount) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = amount.NewAmountFromBytes(v)
		return tr.sum, nil
	}
}

func (tr *TypeReader) bigInt(r io.Reader, p **big.Int) (int64, error) {
	if v, n, err := ReadBytes(r); err != nil {
		return tr.sum, err
	} else {
		tr.sum += int64(n)
		(*p) = big.NewInt(0).SetBytes(v)
		return tr.sum, nil
	}
}

func (tr *TypeReader) addrs(r io.Reader, p *[]common.Address) (int64, error) {
	if Len, n, err := ReadUint8(r); err != nil {
		tr.sum += n
		return tr.sum, err
	} else {
		(*p) = make([]common.Address, Len)
		for i := 0; i < int(Len); i++ {
			if n, err := FillBytes(r, (*p)[i][:]); err != nil {
				return tr.sum, err
			} else {
				tr.sum += int64(n)
			}
		}
	}
	return tr.sum, nil
}

func (tr *TypeReader) amounts(r io.Reader, p *[]*amount.Amount) (int64, error) {
	if Len, n, err := ReadUint8(r); err != nil {
		tr.sum += n
		return tr.sum, err
	} else {
		(*p) = make([]*amount.Amount, Len)
		for i := 0; i < int(Len); i++ {
			if v, n, err := ReadBytes(r); err != nil {
				return tr.sum, err
			} else {
				tr.sum += int64(n)
				(*p)[i] = amount.NewAmountFromBytes(v)
			}
		}
	}
	return tr.sum, nil
}

func (tr *TypeReader) readerFrom(r io.Reader, p io.ReaderFrom) (int64, error) {
	if n, err := p.ReadFrom(r); err != nil {
		return tr.sum, errors.WithStack(err)
	} else {
		tr.sum += int64(n)
		return tr.sum, nil
	}
}

func (tr *TypeReader) Sum() int64 {
	return tr.sum
}
