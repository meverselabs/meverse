package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// RewardPolicy defines a reward policy
type RewardPolicy struct {
	RewardPerBlock        *amount.Amount
	PayRewardEveryBlocks  uint32
	AlphaEfficiency1000   uint32
	SigmaEfficiency1000   uint32
	OmegaEfficiency1000   uint32
	HyperEfficiency1000   uint32
	StakingEfficiency1000 uint32
}

// MarshalJSON is a marshaler function
func (pc *RewardPolicy) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"alpha_efficiency_1000":`)
	if bs, err := json.Marshal(pc.AlphaEfficiency1000); err != nil {
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
	buffer.WriteString(`"omega_efficiency_1000":`)
	if bs, err := json.Marshal(pc.OmegaEfficiency1000); err != nil {
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
	buffer.WriteString(`"staking_efficiency_1000":`)
	if bs, err := json.Marshal(pc.StakingEfficiency1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// AlphaPolicy defines a alpha policy
type AlphaPolicy struct {
	AlphaCreationAmount       *amount.Amount
	AlphaUnlockRequiredBlocks uint32
	AlphaCreationLimitHeight  uint32
}

// MarshalJSON is a marshaler function
func (pc *AlphaPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"alpha_creation_amount":`)
	if bs, err := pc.AlphaCreationAmount.MarshalJSON(); err != nil {
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
	buffer.WriteString(`"alpha_creation_limit_height":`)
	if bs, err := json.Marshal(pc.AlphaCreationLimitHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// SigmaPolicy defines a sigma policy
type SigmaPolicy struct {
	SigmaRequiredAlphaBlocks  uint32
	SigmaRequiredAlphaCount   uint32
	SigmaUnlockRequiredBlocks uint32
}

// MarshalJSON is a marshaler function
func (pc *SigmaPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
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
	buffer.WriteString(`"sigma_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.SigmaUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// OmegaPolicy defines a omega policy
type OmegaPolicy struct {
	OmegaRequiredSigmaBlocks  uint32
	OmegaRequiredSigmaCount   uint32
	OmegaUnlockRequiredBlocks uint32
}

// MarshalJSON is a marshaler function
func (pc *OmegaPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
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
	buffer.WriteString(`"omega_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.OmegaUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// HyperPolicy defines a hyper policy
type HyperPolicy struct {
	HyperCreationAmount         *amount.Amount
	HyperUnlockRequiredBlocks   uint32
	StakingUnlockRequiredBlocks uint32
}

// MarshalJSON is a marshaler function
func (pc *HyperPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"hyper_creation_amount":`)
	if bs, err := pc.HyperCreationAmount.MarshalJSON(); err != nil {
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
	buffer.WriteString(`"staking_unlock_required_blocks":`)
	if bs, err := json.Marshal(pc.StakingUnlockRequiredBlocks); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// ValidatorPolicy defines a policy of Hyper formulator
type ValidatorPolicy struct {
	CommissionRatio1000 uint32
	MinimumStaking      *amount.Amount
	PayOutInterval      uint32
}

// Clone returns the clonend value of it
func (pc *ValidatorPolicy) Clone() *ValidatorPolicy {
	return &ValidatorPolicy{
		CommissionRatio1000: pc.CommissionRatio1000,
		MinimumStaking:      pc.MinimumStaking.Clone(),
		PayOutInterval:      pc.PayOutInterval,
	}
}

// MarshalJSON is a marshaler function
func (pc *ValidatorPolicy) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"pay_out_interval":`)
	if bs, err := json.Marshal(pc.PayOutInterval); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// TransmutePolicy defines a transmute policy
type TransmutePolicy struct {
	TransmuteEnableHeightFrom uint32
	TransmuteEnableHeightTo   uint32
}

// MarshalJSON is a marshaler function
func (pc *TransmutePolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"transmute_enable_height_from":`)
	if bs, err := json.Marshal(pc.TransmuteEnableHeightFrom); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"transmute_enable_height_to":`)
	if bs, err := json.Marshal(pc.TransmuteEnableHeightTo); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// MiningFeePolicy defines a mining fee policy
type MiningFeePolicy struct {
	MiningFeeAddress common.Address
	MiningFee1000    uint32
}

// MarshalJSON is a marshaler function
func (pc *MiningFeePolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"mining_fee_address":`)
	if bs, err := pc.MiningFeeAddress.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"mining_fee_1000":`)
	if bs, err := json.Marshal(pc.MiningFee1000); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
