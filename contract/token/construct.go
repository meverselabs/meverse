package token

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type TokenContractConstruction struct {
	Name   string
	Symbol string
	// FeeTokenAddress  common.Address
	InitialSupplyMap map[common.Address]*amount.Amount
}

func (s *TokenContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.String(w, s.Name); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Symbol); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, uint32(len(s.InitialSupplyMap))); err != nil {
		return sum, err
	}
	for k, v := range s.InitialSupplyMap {
		if sum, err := sw.Address(w, k); err != nil {
			return sum, err
		}
		if sum, err := sw.Amount(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *TokenContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.String(r, &s.Name); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Symbol); err != nil {
		return sum, err
	}
	if Len, sum, err := sr.GetUint32(r); err != nil {
		return sum, err
	} else {
		s.InitialSupplyMap = map[common.Address]*amount.Amount{}
		for i := uint32(0); i < Len; i++ {
			var addr common.Address
			if sum, err := sr.Address(r, &addr); err != nil {
				return sum, err
			}
			var am *amount.Amount
			if sum, err := sr.Amount(r, &am); err != nil {
				return sum, err
			}
			s.InitialSupplyMap[addr] = am
		}
	}
	return sr.Sum(), nil
}
