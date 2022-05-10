package nft1155

import (
	"io"

	"github.com/meverselabs/meverse/common/bin"
)

type NFT1155ContractConstruction struct {
	Name   string
	Symbol string
}

func (s *NFT1155ContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.String(w, s.Name); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Symbol); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *NFT1155ContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.String(r, &s.Name); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Symbol); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
