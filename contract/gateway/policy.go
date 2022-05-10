package gateway

import (
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type GatewayPolicy struct {
	Fee       *amount.Amount
	ChainID   *big.Int
	ChainName string
}

func (s *GatewayPolicy) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.Fee); err != nil {
		return sum, err
	}
	if sum, err := sw.BigInt(w, s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.ChainName); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *GatewayPolicy) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Amount(r, &s.Fee); err != nil {
		return sum, err
	}
	if sum, err := sr.BigInt(r, &s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.ChainName); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
