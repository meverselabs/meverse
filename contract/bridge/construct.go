package bridge

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type BridgeContractConstruction struct {
	Bank         common.Address
	FeeOwner     common.Address
	MeverseToken common.Address
}

func (s *BridgeContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Bank); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.FeeOwner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.MeverseToken); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BridgeContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Bank); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.FeeOwner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.MeverseToken); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
