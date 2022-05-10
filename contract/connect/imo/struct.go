package imo

import (
	"io"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type UserInfo struct {
	Amt     *amount.Amount
	Claimed bool
}

func (s *UserInfo) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.Amt); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.Claimed); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *UserInfo) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Amount(r, &s.Amt); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.Claimed); err != nil {
		return sum, err
	}

	return sr.Sum(), nil
}
