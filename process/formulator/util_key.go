package formulator

import (
	"encoding/binary"

	"github.com/fletaio/fleta/common"
)

// tags
var (
	tagStakingAmount         = []byte{1, 0}
	tagStakingAmountNumber   = []byte{1, 1}
	tagStakingAmountReverse  = []byte{1, 2}
	tagStakingAmountCount    = []byte{1, 3}
	tagRewardPower           = []byte{2, 0}
	tagRewardPowerNumber     = []byte{2, 1}
	tagRewardPowerReverse    = []byte{2, 2}
	tagRewardPowerCount      = []byte{2, 3}
	tagRewardPolicy          = []byte{3, 0}
	tagAlphaPolicy           = []byte{3, 1}
	tagSigmaPolicy           = []byte{3, 2}
	tagOmegaPolicy           = []byte{3, 3}
	tagHyperPolicy           = []byte{3, 4}
	tagGenCount              = []byte{4, 0}
	tagGenCountNumber        = []byte{4, 1}
	tagGenCountReverse       = []byte{4, 2}
	tagGenCountCount         = []byte{4, 3}
	tagStakingAmountMap      = []byte{5, 0}
	tagStakingPowerMap       = []byte{5, 1}
	tagStackRewardMap        = []byte{5, 2}
	tagLastStakingPaidHeight = []byte{5, 3}
	tagAutoStaking           = []byte{5, 4}
	tagLastPaidHeight        = []byte{5, 5}
)

func toStakingAmountKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagStakingAmount)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func toStakingAmountNumberKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagStakingAmountNumber)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func toStakingAmountReverseKey(Num uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagStakingAmountReverse)
	binary.BigEndian.PutUint32(bs[2:], Num)
	return bs
}

func toRewardPowerKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagRewardPower)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func toRewardPowerNumberKey(StakingAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagRewardPowerNumber)
	copy(bs[2:], StakingAddrss[:])
	return bs
}

func toRewardPowerReverseKey(Num uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagRewardPowerReverse)
	binary.BigEndian.PutUint32(bs[2:], Num)
	return bs
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

func toGenCountNumberKey(GenAddrss common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagGenCountNumber)
	copy(bs[2:], GenAddrss[:])
	return bs
}

func toGenCountReverseKey(Num uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagGenCountReverse)
	binary.BigEndian.PutUint32(bs[2:], Num)
	return bs
}
