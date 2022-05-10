package depositpool

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type DepositPoolContractConstruction struct {
	Owner common.Address
	Token common.Address
	Amt   *amount.Amount
}

func (s *DepositPoolContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Token); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.Amt); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *DepositPoolContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Token); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.Amt); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
