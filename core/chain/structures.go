package chain

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

// DeployContractData defines data of deploy contract tx
type DeployContractData struct {
	Owner   common.Address
	ClassID uint64
	Args    []byte
}

func (s *DeployContractData) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.ClassID); err != nil {
		return sum, err
	}
	if sum, err := sw.Bytes(w, s.Args); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *DeployContractData) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.ClassID); err != nil {
		return sum, err
	}
	if sum, err := sr.Bytes(r, &s.Args); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
