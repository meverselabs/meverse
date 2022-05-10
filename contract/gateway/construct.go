package gateway

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type GatewayContractConstruction struct {
	TokenAddress common.Address
}

func (s *GatewayContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.TokenAddress); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *GatewayContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.TokenAddress); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
