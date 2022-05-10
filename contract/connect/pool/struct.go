package pool

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type PoolInfo struct {
	Want             common.Address
	AllocPoint       uint32
	LastRewardBlock  uint32
	AccTokenPerShare *amount.Amount
	Strat            common.Address
}

func (s *PoolInfo) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Want); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.AllocPoint); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.LastRewardBlock); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.AccTokenPerShare); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Strat); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *PoolInfo) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Want); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.AllocPoint); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.LastRewardBlock); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.AccTokenPerShare); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Strat); err != nil {
		return sum, err
	}

	return sr.Sum(), nil
}

type UserInfo struct {
	Shares     *amount.Amount
	RewardDebt *amount.Amount
}

func (s *UserInfo) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.Shares); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.RewardDebt); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *UserInfo) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Amount(r, &s.Shares); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.RewardDebt); err != nil {
		return sum, err
	}

	return sr.Sum(), nil
}
