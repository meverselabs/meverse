package trade

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
)

type UniSwapConstruction struct {
	Name      string
	Symbol    string
	Factory   common.Address
	Token0    common.Address
	Token1    common.Address
	PayToken  common.Address
	Owner     common.Address
	Winner    common.Address
	Fee       uint64
	AdminFee  uint64
	WinnerFee uint64
	WhiteList common.Address
	GroupId   hash.Hash256
}

func (s *UniSwapConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.String(w, s.Name); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Symbol); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Factory); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Token0); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Token1); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.PayToken); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Winner); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Fee); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.AdminFee); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.WinnerFee); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.WhiteList); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.GroupId); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}
func (s *UniSwapConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.String(r, &s.Name); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Symbol); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Factory); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Token0); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Token1); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.PayToken); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Owner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Winner); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Fee); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.AdminFee); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.WinnerFee); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.WhiteList); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.GroupId); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
