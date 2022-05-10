package bridge

import (
	"io"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type SendMaintokenInfo struct {
	overthen *amount.Amount
	amt      *amount.Amount
}

func (s *SendMaintokenInfo) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.overthen); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.amt); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *SendMaintokenInfo) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	s.overthen = &amount.Amount{}
	if sum, err := sr.Amount(r, &s.overthen); err != nil {
		return sum, err
	}
	s.amt = &amount.Amount{}
	if sum, err := sr.Amount(r, &s.amt); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
