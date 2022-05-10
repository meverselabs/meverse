package p2p

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/bluele/gcache"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
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
	return bin.Uint64(BNum), int64(read), nil
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
	return bin.Uint32(BNum), int64(read), nil
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
	return bin.Uint16(BNum), int64(read), nil
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
				return int64(read), errors.WithStack(ErrInvalidLength)
			}
		}
	}
	if read != len(bs) {
		return int64(read), errors.WithStack(ErrInvalidLength)
	}
	return int64(read), nil
}

var compressed bool = true

// MessageToPacket returns packet of the message
func MessageToPacket(m Serializable) []byte {
	var buffer bytes.Buffer
	buffer.Write(bin.Uint32Bytes(m.TypeID()))
	buffer.Write(make([]byte, 4))
	if compressed {
		zw := gzip.NewWriter(&buffer)
		if _, err := m.WriteTo(zw); err != nil {
			panic(err)
		}
		zw.Flush()
		zw.Close()
	} else {
		m.WriteTo(&buffer)
	}

	temp := buffer.Bytes()
	length := len(temp)
	var body []byte //are we wasting more than 10% space?
	if cap(temp) > (length + length/10) {
		body = make([]byte, length)
		copy(body, temp)
	} else {
		body = temp
	}
	temp = nil
	buffer.Reset()

	bin.PutUint32(body[4:], uint32(len(body)-8))
	return body
}

func PacketMessageType(bs []byte) uint32 {
	return bin.Uint32(bs)
}

func PacketToMessage(bs []byte) (Serializable, error) {
	t := PacketMessageType(bs)

	m, err := CreateSerializable(t)
	if err != nil {
		return nil, err
	}

	var mbs []byte
	if compressed {
		zr, err := gzip.NewReader(bytes.NewReader(bs[8:]))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer zr.Close()

		v, err := ioutil.ReadAll(zr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		mbs = v
	} else {
		mbs = bs[8:]
	}

	rd := bytes.NewReader(mbs)
	if _, err := m.ReadFrom(rd); err != nil {
		return nil, errors.WithStack(err)
	}
	mbs = nil
	return m, nil
}

func BlockPacketWithCache(msg *RequestMessage, provider types.Provider, batchCache gcache.Cache, singleCache gcache.Cache) ([]byte, error) {
	Height := provider.Height()
	var bs []byte
	if msg.Height%10 == 0 && msg.Count == 10 && msg.Height+uint32(msg.Count) <= Height {
		value, err := batchCache.Get(msg.Height)
		if err != nil {
			list := make([]*types.Block, 0, 10)
			for i := uint32(0); i < uint32(msg.Count); i++ {
				if msg.Height+i > Height {
					break
				}
				b, err := provider.Block(msg.Height + i)
				if err != nil {
					return nil, err
				}
				list = append(list, b)
			}
			sm := &BlockMessage{
				Blocks: list,
			}
			bs = MessageToPacket(sm)
			batchCache.Set(msg.Height, bs)
		} else {
			bs = value.([]byte)
		}
	} else if msg.Count == 1 {
		value, err := singleCache.Get(msg.Height)
		if err != nil {
			b, err := provider.Block(msg.Height)
			if err != nil {
				return nil, err
			}
			sm := &BlockMessage{
				Blocks: []*types.Block{b},
			}
			bs = MessageToPacket(sm)
			singleCache.Set(msg.Height, bs)
		} else {
			bs = value.([]byte)
		}
	} else {
		list := make([]*types.Block, 0, 10)
		for i := uint32(0); i < uint32(msg.Count); i++ {
			if msg.Height+i > Height {
				break
			}
			b, err := provider.Block(msg.Height + i)
			if err != nil {
				return nil, err
			}
			list = append(list, b)
		}
		sm := &BlockMessage{
			Blocks: list,
		}
		bs = MessageToPacket(sm)
	}
	return bs, nil
}

func MakeHandSharkBs(chainID *big.Int, rn uint64, ti uint64) ([]byte, error) {
	req := bytes.Buffer{}
	sw := bin.NewSumWriter()
	if _, err := sw.Bytes(&req, chainID.Bytes()); err != nil {
		return nil, err
	}
	if _, err := sw.Uint64(&req, rn); err != nil {
		return nil, err
	}
	if _, err := sw.Uint64(&req, ti); err != nil {
		return nil, err
	}
	return req.Bytes(), nil
}

func RecoveryHandSharkBs(bs []byte) (chainID *big.Int, rn uint64, ti uint64, err error) {
	req := bytes.NewBuffer(bs)
	sr := bin.NewSumReader()
	chainID = big.NewInt(0)
	ci := []byte{}
	if _, err = sr.Bytes(req, &ci); err != nil {
		return
	}
	chainID.SetBytes(ci)
	if _, err = sr.Uint64(req, &rn); err != nil {
		return
	}
	if _, err = sr.Uint64(req, &ti); err != nil {
		return
	}
	return
}
