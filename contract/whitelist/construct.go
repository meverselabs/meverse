package whitelist

import (
	"io"
)

type WhiteListContractConstruction struct {
}

func (s *WhiteListContractConstruction) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (s *WhiteListContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}
