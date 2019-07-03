package formulator

import (
	"bytes"

	"github.com/fletaio/fleta/common"
)

// tags
var (
	tagStakingAmount  = []byte{1, 0}
	tagStakingPower   = []byte{1, 1}
	tagRewardPower    = []byte{1, 2}
	tagAutoStaking    = []byte{1, 3}
	tagLastPaidHeight = []byte{1, 4}
	tagRewardPolicy   = []byte{2, 0}
	tagAlphaPolicy    = []byte{2, 1}
	tagSigmaPolicy    = []byte{2, 2}
	tagOmegaPolicy    = []byte{2, 3}
	tagHyperPolicy    = []byte{2, 4}
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

func toStakingPowerKey(HyperAddress common.Address, StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize*2)
	copy(bs, tagRewardPower)
	copy(bs[2:], HyperAddress[:])
	copy(bs[2+common.AddressSize:], StakingAddrss[:])
	return bs
}

func fromStakingPowerKey(bs []byte) (common.Address, common.Address, bool) {
	if bytes.HasPrefix(bs, tagRewardPower) {
		var HyperAddress common.Address
		copy(HyperAddress[:], bs[2:])
		var StakingAddress common.Address
		copy(StakingAddress[:], bs[2+common.AddressSize:])
		return HyperAddress, StakingAddress, true
	} else {
		return common.Address{}, common.Address{}, false
	}
}

func toAutoStakingKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagAutoStaking)
	copy(bs[2:], StakingAddrss[:])
	return bs
}
