package types

import (
	"io"

	"github.com/meverselabs/meverse/common/bin"
)

type Block struct {
	Header Header
	Body   Body
}

func (s *Block) Clone() *Block {
	return &Block{
		Header: s.Header.Clone(),
		Body:   s.Body.Clone(),
	}
}

func (s *Block) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.WriterTo(w, &s.Header); err != nil {
		return sum, err
	}
	if sum, err := sw.WriterTo(w, &s.Body); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Block) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.ReaderFrom(r, &s.Header); err != nil {
		return sum, err
	}
	if sum, err := sr.ReaderFrom(r, &s.Body); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
