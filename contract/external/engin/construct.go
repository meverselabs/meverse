package engin

import (
	"io"
)

type EnginContractConstruction struct {
}

func (s *EnginContractConstruction) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (s *EnginContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}
