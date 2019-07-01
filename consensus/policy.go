package consensus

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/util"
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

// WriteTo is a serialization function
func (pc *ConsensusPolicy) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := pc.RewardPerBlock.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.PayRewardEveryBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.FormulatorCreationLimitHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := pc.AlphaCreationAmount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.AlphaEfficiency1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.AlphaUnlockRequiredBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.SigmaRequiredAlphaBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.SigmaRequiredAlphaCount); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.SigmaEfficiency1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.SigmaUnlockRequiredBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.OmegaRequiredSigmaBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.OmegaRequiredSigmaCount); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.OmegaEfficiency1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.OmegaUnlockRequiredBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := pc.HyperCreationAmount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.HyperEfficiency1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.HyperUnlockRequiredBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.StakingEfficiency1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, pc.StakingUnlockRequiredBlocks); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (pc *ConsensusPolicy) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := pc.RewardPerBlock.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.PayRewardEveryBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.FormulatorCreationLimitHeight = v
	}
	if n, err := pc.AlphaCreationAmount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.AlphaEfficiency1000 = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.AlphaUnlockRequiredBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.SigmaRequiredAlphaBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.SigmaRequiredAlphaCount = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.SigmaEfficiency1000 = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.SigmaUnlockRequiredBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.OmegaRequiredSigmaBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.OmegaRequiredSigmaCount = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.OmegaEfficiency1000 = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.OmegaUnlockRequiredBlocks = v
	}
	if n, err := pc.HyperCreationAmount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.HyperEfficiency1000 = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.HyperUnlockRequiredBlocks = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.StakingEfficiency1000 = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.StakingUnlockRequiredBlocks = v
	}
	return read, nil
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

// WriteTo is a serialization function
func (pc *HyperPolicy) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint32(w, pc.CommissionRatio1000); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := pc.MinimumStaking.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := pc.MaximumStaking.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (pc *HyperPolicy) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		pc.CommissionRatio1000 = v
	}
	if n, err := pc.MinimumStaking.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := pc.MaximumStaking.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
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

// WriteTo is a serialization function
func (pc *StakingPolicy) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteBool(w, pc.AutoStaking); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (pc *StakingPolicy) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadBool(r); err != nil {
		return read, err
	} else {
		read += n
		pc.AutoStaking = v
	}
	return read, nil
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
