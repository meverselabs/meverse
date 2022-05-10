package pool

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

type PoolContractConstruction struct {
	Gov  common.Address
	Farm common.Address
	Want common.Address

	FeeFundAddress common.Address
	RewardsAddress common.Address

	DepositFeeFactor  uint16
	WithdrawFeeFactor uint16
	RewardFeeFactor   uint16
}

func (s *PoolContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Gov); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Farm); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Want); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.FeeFundAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.RewardsAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.DepositFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.WithdrawFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.RewardFeeFactor); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *PoolContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Gov); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Farm); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Want); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.FeeFundAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.RewardsAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.DepositFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.WithdrawFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.RewardFeeFactor); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
