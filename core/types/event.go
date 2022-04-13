package types

import (
	"io"

	"github.com/fletaio/fleta_v2/common/bin"
)

const (
	EventTagTxMsg  = byte(0x00)
	EventTagReward = byte(0x01)
)

type Event struct {
	Index  uint16
	Type   uint8
	Result []byte
}

func NewEvent(result []byte) *Event {
	return &Event{
		Result: result,
	}
}

func (s *Event) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, s.Index); err != nil {
		return sum, err
	}
	if sum, err := sw.Bytes(w, s.Result); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint8(w, s.Type); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Event) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint16(r, &s.Index); err != nil {
		return sum, err
	}
	if sum, err := sr.Bytes(r, &s.Result); err != nil {
		return sum, err
	}
	sr.Uint8(r, &s.Type)
	return sr.Sum(), nil
}
