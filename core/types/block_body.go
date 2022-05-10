package types

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type Body struct {
	Transactions          []*Transaction
	Events                []*Event
	TransactionSignatures []common.Signature
	BlockSignatures       []common.Signature
}

func (s *Body) Clone() Body {
	return Body{
		Transactions:          s.Transactions,
		Events:                s.Events,
		TransactionSignatures: s.TransactionSignatures,
		BlockSignatures:       s.BlockSignatures,
	}
}

func (s *Body) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, uint16(len(s.Transactions))); err != nil {
		return sum, err
	}
	for _, v := range s.Transactions {
		if sum, err := sw.WriterTo(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Uint16(w, uint16(len(s.TransactionSignatures))); err != nil {
		return sum, err
	}
	for _, v := range s.TransactionSignatures {
		if sum, err := sw.Signature(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Uint16(w, uint16(len(s.Events))); err != nil {
		return sum, err
	}
	for _, v := range s.Events {
		if sum, err := sw.WriterTo(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Uint16(w, uint16(len(s.BlockSignatures))); err != nil {
		return sum, err
	}
	for _, v := range s.BlockSignatures {
		if sum, err := sw.Signature(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *Body) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Transactions = make([]*Transaction, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v Transaction
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			s.Transactions = append(s.Transactions, &v)
		}
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.TransactionSignatures = make([]common.Signature, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v common.Signature
			if sum, err := sr.Signature(r, &v); err != nil {
				return sum, err
			}
			s.TransactionSignatures = append(s.TransactionSignatures, v)
		}
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Events = make([]*Event, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v Event
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			s.Events = append(s.Events, &v)
		}
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.BlockSignatures = make([]common.Signature, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v common.Signature
			if sum, err := sr.Signature(r, &v); err != nil {
				return sum, err
			}
			s.BlockSignatures = append(s.BlockSignatures, v)
		}
	}
	return sr.Sum(), nil
}
