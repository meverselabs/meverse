package formulator

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
)

type FormulatorPolicy struct {
	AlphaAmount    *amount.Amount
	SigmaCount     uint8
	SigmaBlocks    uint32
	OmegaCount     uint8
	OmegaBlocks    uint32
	HyperAmount    *amount.Amount
	MinStakeAmount *amount.Amount
}

func (s *FormulatorPolicy) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.AlphaAmount); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint8(w, s.SigmaCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.SigmaBlocks); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint8(w, s.OmegaCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.OmegaBlocks); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.HyperAmount); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.MinStakeAmount); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *FormulatorPolicy) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Amount(r, &s.AlphaAmount); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint8(r, &s.SigmaCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.SigmaBlocks); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint8(r, &s.OmegaCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.OmegaBlocks); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.HyperAmount); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.MinStakeAmount); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

type RewardPolicy struct {
	RewardPerBlock        *amount.Amount
	AlphaEfficiency1000   uint32
	SigmaEfficiency1000   uint32
	OmegaEfficiency1000   uint32
	HyperEfficiency1000   uint32
	StakingEfficiency1000 uint32
	CommissionRatio1000   uint32
	MiningFeeAddress      common.Address
	MiningFee1000         uint32
}

func (s *RewardPolicy) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Amount(w, s.RewardPerBlock); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.AlphaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.SigmaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.OmegaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.HyperEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.StakingEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.CommissionRatio1000); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.MiningFeeAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.MiningFee1000); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *RewardPolicy) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Amount(r, &s.RewardPerBlock); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.AlphaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.SigmaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.OmegaEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.HyperEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.StakingEfficiency1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.CommissionRatio1000); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.MiningFeeAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.MiningFee1000); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
