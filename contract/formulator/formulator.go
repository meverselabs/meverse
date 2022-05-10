package formulator

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

const (
	AlphaFormulatorType = uint8(0x01)
	SigmaFormulatorType = uint8(0x02)
	OmegaFormulatorType = uint8(0x03)
	AlphaMiningPower    = 1
	SigmaMiningPower    = 1.15
	OmegaMiningPower    = 1.3
)

type Formulator struct {
	Type    uint8
	Height  uint32
	Amount  *amount.Amount
	Owner   common.Address
	TokenID common.Address
}

func (s *Formulator) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint8(w, s.Type); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.Height); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.Amount); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.TokenID); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Formulator) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint8(r, &s.Type); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.Height); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.Amount); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.TokenID); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
