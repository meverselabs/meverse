package mappfarm

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type FarmContractConstruction struct {
	Owner         common.Address
	Banker        common.Address
	FarmToken     common.Address
	WantToken     common.Address
	OwnerReward   uint16
	TokenPerBlock *amount.Amount
	StartBlock    uint32
}

func (s *FarmContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Banker); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.FarmToken); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.WantToken); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.OwnerReward); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.TokenPerBlock); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.StartBlock); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *FarmContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Banker); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.FarmToken); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.WantToken); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.OwnerReward); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.TokenPerBlock); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.StartBlock); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
