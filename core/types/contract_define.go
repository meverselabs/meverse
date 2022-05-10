package types

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

// ContractDefine defines chain Contract functions
type ContractDefine struct {
	Address common.Address
	Owner   common.Address
	ClassID uint64
}

func (s *ContractDefine) Clone() *ContractDefine {
	return &ContractDefine{
		Address: s.Address,
		Owner:   s.Owner,
		ClassID: s.ClassID,
	}
}

func (s *ContractDefine) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Address); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.ClassID); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *ContractDefine) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Address); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.ClassID); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
