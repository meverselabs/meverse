package types

import (
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
)

type Header struct {
	ChainID       *big.Int       // ~=8byte
	Version       uint16         // 2byte
	Height        uint32         // 4byte
	PrevHash      hash.Hash256   // 32byte
	LevelRootHash hash.Hash256   // 32byte
	ContextHash   hash.Hash256   // 32byte
	TimeoutCount  uint32         // 4byte
	Timestamp     uint64         // 8byte
	Generator     common.Address // 20byte
	Gas           uint16         // 2byte
}

func (s *Header) Clone() Header {
	return Header{
		ChainID:       s.ChainID,
		Version:       s.Version,
		Height:        s.Height,
		PrevHash:      s.PrevHash,
		LevelRootHash: s.LevelRootHash,
		ContextHash:   s.ContextHash,
		TimeoutCount:  s.TimeoutCount,
		Timestamp:     s.Timestamp,
		Generator:     s.Generator,
		Gas:           s.Gas,
	}
}

func (s *Header) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.BigInt(w, s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.Version); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.Height); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.PrevHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.LevelRootHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.ContextHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.Gas); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Header) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.BigInt(r, &s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.Version); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.Height); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.PrevHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.LevelRootHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.ContextHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.Gas); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
