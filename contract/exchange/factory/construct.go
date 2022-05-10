package factory

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type FactoryContractConstruction struct {
	Owner common.Address
}

func (s *FactoryContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}

	return sw.Sum(), nil
}
func (s *FactoryContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
