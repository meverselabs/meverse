package goplugin

import (
	"io"

	"github.com/meverselabs/meverse/common/bin"
)

type PluginContractConstruction struct {
	Bin []byte
}

func (s *PluginContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Bytes(w, s.Bin); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *PluginContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Bytes(r, &s.Bin); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
