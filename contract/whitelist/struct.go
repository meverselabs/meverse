package whitelist

import (
	"errors"
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type GroupData struct {
	delegate    common.Address
	method      string
	params      []interface{}
	checkResult string
	result      []byte
}

func (s *GroupData) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.delegate); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.method); err != nil {
		return sum, err
	}
	plen := len(s.params)
	params := bin.TypeWriteAll(s.params...)
	if _, err := bin.TypeReadAll(params, plen); err != nil {
		return sw.Sum(), err
	}
	if sum, err := sw.Bytes(w, params); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.checkResult); err != nil {
		return sum, err
	}
	if sum, err := sw.Bytes(w, s.result); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *GroupData) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.delegate); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.method); err != nil {
		return sum, err
	}
	var params []byte
	if sum, err := sr.Bytes(r, &params); err != nil {
		return sum, err
	}
	var err error
	if s.params, err = bin.TypeReadAll(params, -1); err != nil {
		return sr.Sum(), err
	}
	if sum, err := sr.String(r, &s.checkResult); err != nil {
		return sum, err
	}
	if sum, err := sr.Bytes(r, &s.result); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

func (s *GroupData) ParseParam(m map[string]interface{}) error {
	if s.method == "" {
		return errors.New("groupdata not init")
	}
	if s.params == nil {
		return nil
	}

	for i, iv := range s.params {
		if str, ok := iv.(string); ok {
			if v, ok := m[str]; ok {
				s.params[i] = v
			}
		}
	}
	return nil
}
