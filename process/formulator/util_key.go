package formulator

import (
	"bytes"

	"github.com/fletaio/fleta/common"
)

// tags
var (
	tagStakingAmount         = []byte{1, 0}
	tagRewardPower           = []byte{1, 2}
	tagAutoStaking           = []byte{1, 3}
	tagLastPaidHeight        = []byte{1, 4}
	tagRewardPolicy          = []byte{2, 0}
	tagAlphaPolicy           = []byte{2, 1}
	tagSigmaPolicy           = []byte{2, 2}
	tagOmegaPolicy           = []byte{2, 3}
	tagHyperPolicy           = []byte{2, 4}
	tagGenCount              = []byte{3, 0}
	tagStakingAmountMap      = []byte{4, 0}
	tagStakingPowerMap       = []byte{4, 1}
	tagStackRewardMap        = []byte{4, 2}
	tagLastStakingPaidHeight = []byte{4, 3}
)

func toStakingAmountKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagStakingAmount)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func fromStakingAmountKey(bs []byte) (common.Address, bool) {
	if bytes.HasPrefix(bs, tagStakingAmount) {
		var addr common.Address
		copy(addr[:], bs[2:])
		return addr, true
	} else {
		return common.Address{}, false
	}
}

func toRewardPowerKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagRewardPower)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func fromRewardPowerKey(bs []byte) (common.Address, bool) {
	if bytes.HasPrefix(bs, tagRewardPower) {
		var addr common.Address
		copy(addr[:], bs[2:])
		return addr, true
	} else {
		return common.Address{}, false
	}
}

func toAutoStakingKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagAutoStaking)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func toGenCountKey(GenAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagGenCount)
	copy(bs[2:], GenAddrss[:])
	return bs
}

func fromGenCountKey(bs []byte) (common.Address, bool) {
	if bytes.HasPrefix(bs, tagGenCount) {
		var addr common.Address
		copy(addr[:], bs[2:])
		return addr, true
	} else {
		return common.Address{}, false
	}
}
