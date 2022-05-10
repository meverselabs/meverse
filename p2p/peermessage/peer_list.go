package peermessage

import (
	"io"
	"time"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/pkg/errors"
)

//peermessage errors
var (
	ErrOverflowLength = errors.New("Overflow string length")
)

// ConnectInfo is a structure of connection information that includes ping time and score board.
type ConnectInfo struct {
	Address        string
	Hash           string
	PingTime       time.Duration
	PingScoreBoard *ScoreBoardMap
}

// NewConnectInfo is creator of ConnectInfo
func NewConnectInfo(addr string, hash string, t time.Duration) ConnectInfo {
	return ConnectInfo{
		Address:        addr,
		Hash:           hash,
		PingTime:       t,
		PingScoreBoard: &ScoreBoardMap{},
	}
}

// WriteTo is a serialization function
func (ci *ConnectInfo) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	{
		n, err := w.Write(bin.Uint64Bytes(uint64(ci.PingTime)))
		if err != nil {
			return wrote, errors.WithStack(err)
		}
		wrote += int64(n)
	}
	{
		bs := []byte(ci.Address)

		bslen := len(bs)
		if bslen > 255 {
			return wrote, errors.WithStack(ErrOverflowLength)
		}
		bsLen := uint8(bslen)
		n, err := w.Write([]byte{byte(bsLen)})
		if err != nil {
			return wrote, errors.WithStack(err)
		}
		wrote += int64(n)

		nint, err := w.Write(bs)
		if err != nil {
			return wrote, errors.WithStack(err)
		}
		wrote += int64(nint)
	}
	{
		bs := []byte(ci.Hash)

		bslen := len(bs)
		if bslen > 255 {
			return wrote, errors.WithStack(ErrOverflowLength)
		}
		bsLen := uint8(bslen)
		n, err := w.Write([]byte{byte(bsLen)})
		if err != nil {
			return wrote, errors.WithStack(err)
		}
		wrote += int64(n)

		nint, err := w.Write(bs)
		if err != nil {
			return wrote, errors.WithStack(err)
		}
		wrote += int64(nint)
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (ci *ConnectInfo) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	{
		bs := make([]byte, 8)
		n, err := r.Read(bs)
		if err != nil {
			return read, errors.WithStack(err)
		}
		read += int64(n)
		v := bin.Uint64(bs)
		ci.PingTime = time.Duration(v)
	}

	{
		bs := make([]byte, 1)
		n, err := r.Read(bs)
		if err != nil {
			return read, errors.WithStack(err)
		}
		read += int64(n)
		bsLen := uint8(bs[0])
		bsBs := make([]byte, bsLen)

		nInt, err := r.Read(bsBs)
		if err != nil {
			return read, errors.WithStack(err)
		}
		read += int64(nInt)

		ci.Address = string(bsBs)
	}
	{
		bs := make([]byte, 1)
		n, err := r.Read(bs)
		if err != nil {
			return read, errors.WithStack(err)
		}
		read += int64(n)
		bsLen := uint8(bs[0])
		bsBs := make([]byte, bsLen)

		nInt, err := r.Read(bsBs)
		if err != nil {
			return read, errors.WithStack(err)
		}
		read += int64(nInt)

		ci.Hash = string(bsBs)
	}
	return read, nil
}

// Score is calculated and returned based on the ping time.
func (ci *ConnectInfo) Score() (score int64) {
	ci.PingScoreBoard.Range(func(addr string, t time.Duration) bool {
		score += int64(t)
		return true
	})
	score /= int64(ci.PingScoreBoard.Len())
	return score
}

const (
	requestTrue  = byte('t')
	requestFalse = byte('f')
)
