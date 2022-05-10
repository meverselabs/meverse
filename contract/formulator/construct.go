package formulator

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type FormulatorContractConstruction struct {
	TokenAddress     common.Address
	FormulatorPolicy FormulatorPolicy
	RewardPolicy     RewardPolicy
}

func (s *FormulatorContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.TokenAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.WriterTo(w, &s.FormulatorPolicy); err != nil {
		return sum, err
	}
	if sum, err := sw.WriterTo(w, &s.RewardPolicy); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *FormulatorContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.TokenAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.ReaderFrom(r, &s.FormulatorPolicy); err != nil {
		return sum, err
	}
	if sum, err := sr.ReaderFrom(r, &s.RewardPolicy); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
