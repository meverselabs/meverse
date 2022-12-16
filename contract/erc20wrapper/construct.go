package erc20wrapper

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type ERC20WrapperContractConstruction struct {
	Erc20Token common.Address
}

func (s *ERC20WrapperContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Erc20Token); err != nil {
		return sum, err
	}

	return sw.Sum(), nil
}

func (s *ERC20WrapperContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Erc20Token); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
