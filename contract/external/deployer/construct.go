package deployer

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type DeployerContractConstruction struct {
	EnginAddress common.Address
	EnginName    string
	EnginVersion uint32
	Binary       []byte
	Owner        common.Address
	Updateable   bool
}

func (s *DeployerContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.EnginAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.EnginName); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.EnginVersion); err != nil {
		return sum, err
	}
	if sum, err := sw.Bytes(w, s.Binary); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.Updateable); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *DeployerContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.EnginAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.EnginName); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.EnginVersion); err != nil {
		return sum, err
	}
	if sum, err := sr.Bytes(r, &s.Binary); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.Updateable); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
