package router

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type RouterContractConstruction struct {
	Factory common.Address
}

func (s *RouterContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Factory); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}
func (s *RouterContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Factory); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
