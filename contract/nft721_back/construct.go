package nft721

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type NFT721ContractConstruction struct {
	Owner  common.Address
	Name   string
	Symbol string
}

func (s *NFT721ContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Name); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Symbol); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *NFT721ContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Name); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Symbol); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
