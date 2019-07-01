package pof

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

var gConsensusPolicyMap = map[uint64]*ConsensusPolicy{}

func SetConsensusPolicy(chainCoord *common.Coordinate, pc *ConsensusPolicy) {
	gConsensusPolicyMap[chainCoord.ID()] = pc
}

func GetConsensusPolicy(chainCoord *common.Coordinate) (*ConsensusPolicy, error) {
	pc, has := gConsensusPolicyMap[chainCoord.ID()]
	if !has {
		return nil, ErrNotExistConsensusPolicy
	}
	return pc, nil
}

// ConsensusPolicy defines a staking policy user
type ConsensusPolicy struct {
	RewardPerBlock                *amount.Amount
	PayRewardEveryBlocks          uint32
	FormulatorCreationLimitHeight uint32
	AlphaCreationAmount           *amount.Amount
	AlphaEfficiency1000           uint32
	AlphaUnlockRequiredBlocks     uint32
	SigmaRequiredAlphaBlocks      uint32
	SigmaRequiredAlphaCount       uint32
	SigmaEfficiency1000           uint32
	SigmaUnlockRequiredBlocks     uint32
	OmegaRequiredSigmaBlocks      uint32
	OmegaRequiredSigmaCount       uint32
	OmegaEfficiency1000           uint32
	OmegaUnlockRequiredBlocks     uint32
	HyperCreationAmount           *amount.Amount
	HyperEfficiency1000           uint32
	HyperUnlockRequiredBlocks     uint32
	StakingEfficiency1000         uint32
	StakingUnlockRequiredBlocks   uint32
}

// MarshalJSON is a marshaler function
func (pc *ConsensusPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"reward_per_block":`)
	if bs, err := pc.RewardPerBlock.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"pay_reward_every_blocks":`)
	if bs, err := json.Marshal(pc.PayRewardEveryBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"formulator_creation_limit_height":`)
	if bs, err := json.Marshal(pc.FormulatorCreationLimitHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"alpha_formulation_amount":`)
	if bs, err := pc.AlphaCreationAmount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"alpha_efficiency_1000":`)
	if bs, err := json.Marshal(pc.AlphaEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"alpha_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.AlphaUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"sigma_required_alpha_blocks":`)
	if bs, err := json.Marshal(pc.SigmaRequiredAlphaBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"sigma_required_alpha_count":`)
	if bs, err := json.Marshal(pc.SigmaRequiredAlphaCount); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"sigma_efficiency_1000":`)
	if bs, err := json.Marshal(pc.SigmaEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"sigma_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.SigmaUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"omega_required_sigma_blocks":`)
	if bs, err := json.Marshal(pc.OmegaRequiredSigmaBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"omega_required_sigma_count":`)
	if bs, err := json.Marshal(pc.OmegaRequiredSigmaCount); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"omega_efficiency_1000":`)
	if bs, err := json.Marshal(pc.OmegaEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"omega_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.OmegaUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_formulation_amount":`)
	if bs, err := pc.HyperCreationAmount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_efficiency_1000":`)
	if bs, err := json.Marshal(pc.HyperEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.HyperUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"staking_efficiency_1000":`)
	if bs, err := json.Marshal(pc.StakingEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"staking_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.StakingUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// HyperPolicy defines a policy of Hyper formulator
type HyperPolicy struct {
	CommissionRatio1000 uint32
	MinimumStaking      *amount.Amount
	MaximumStaking      *amount.Amount
}

// Clone returns the clonend value of it
func (pc *HyperPolicy) Clone() *HyperPolicy {
	return &HyperPolicy{
		CommissionRatio1000: pc.CommissionRatio1000,
		MinimumStaking:      pc.MinimumStaking.Clone(),
		MaximumStaking:      pc.MaximumStaking.Clone(),
	}
}

// MarshalJSON is a marshaler function
func (pc *HyperPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"commission_ratio_1000":`)
	if bs, err := json.Marshal(pc.CommissionRatio1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"minimum_staking":`)
	if bs, err := json.Marshal(pc.MinimumStaking); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"maximum_staking":`)
	if bs, err := json.Marshal(pc.MaximumStaking); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// StakingPolicy defines a staking policy user
type StakingPolicy struct {
	AutoStaking bool
}

// MarshalJSON is a marshaler function
func (pc *StakingPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"auto_staking":`)
	if bs, err := json.Marshal(pc.AutoStaking); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
