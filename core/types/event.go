package types

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

const (
	EventTagTxMsg       = EventType(0x00)
	EventTagReward      = EventType(0x01)
	EventTagCallHistory = EventType(0x02)
)

type EventType uint8

func (e EventType) String() string {
	switch e {
	case EventTagTxMsg:
		return "EventTxMsg"
	case EventTagReward:
		return "EventReward"
	case EventTagCallHistory:
		return "EventCallHistory"
	}
	return "Unknow"
}

type Event struct {
	Index  uint16
	Type   EventType
	Result []byte
}

func NewEvent(Index uint16, Type EventType, result []byte) *Event {
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
	if sum, err := sw.Uint8(w, uint8(s.Type)); err != nil {
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
	var Type uint8
	sr.Uint8(r, &Type)
	s.Type = EventType(Type)
	return sr.Sum(), nil
}

type MethodCallEvent struct {
	From   common.Address
	To     common.Address
	Method string
	Args   []interface{}
	Result []interface{}
	Error  string
}

func (s *MethodCallEvent) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.From); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.To); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Method); err != nil {
		return sum, err
	}
	bs := bin.TypeWriteAll(s.Args...)
	if sum, err := sw.Bytes(w, bs); err != nil {
		return sum, err
	}
	bs = bin.TypeWriteAll(s.Result...)
	if sum, err := sw.Bytes(w, bs); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Error); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *MethodCallEvent) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.From); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.To); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Method); err != nil {
		return sum, err
	}
	var bs []byte
	if sum, err := sr.Bytes(r, &bs); err != nil {
		return sum, err
	}
	var err error
	s.Args, err = bin.TypeReadAll(bs, -1)
	if err != nil {
		return sr.Sum(), err
	}
	if sum, err := sr.Bytes(r, &bs); err != nil {
		if err == io.EOF {
			s.Result = []interface{}{}
			return sum, nil
		}
		return sum, err
	}
	s.Result, err = bin.TypeReadAll(bs, -1)
	if err != nil {
		return sr.Sum(), err
	}
	if sum, err := sr.String(r, &s.Error); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
