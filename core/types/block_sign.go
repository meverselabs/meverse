package types

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
)

// BlockSign is the generator signature of the block
type BlockSign struct {
	HeaderHash         hash.Hash256
	GeneratorSignature common.Signature
}

func (s *BlockSign) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Hash256(w, s.HeaderHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Signature(w, s.GeneratorSignature); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BlockSign) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Hash256(r, &s.HeaderHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Signature(r, &s.GeneratorSignature); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
