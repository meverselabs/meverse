package formulator

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
)

// tags
var (
	tagStakingAmount            = []byte{1, 0}
	tagStakingAmountNumber      = []byte{1, 1}
	tagStakingAmountReverse     = []byte{1, 2}
	tagStakingAmountCount       = []byte{1, 3}
	tagRewardPolicy             = []byte{2, 0}
	tagAlphaPolicy              = []byte{2, 1}
	tagSigmaPolicy              = []byte{2, 2}
	tagOmegaPolicy              = []byte{2, 3}
	tagHyperPolicy              = []byte{2, 4}
	tagTransmutePolicy          = []byte{2, 5}
	tagGenCount                 = []byte{3, 0}
	tagGenCountNumber           = []byte{3, 1}
	tagGenCountReverse          = []byte{3, 2}
	tagGenCountCount            = []byte{3, 3}
	tagStakingAmountMap         = []byte{4, 0}
	tagStakingPowerMap          = []byte{4, 1}
	tagStackRewardMap           = []byte{4, 2}
	tagLastStakingPaidHeight    = []byte{4, 3}
	tagAutoStaking              = []byte{4, 4}
	tagLastPaidHeight           = []byte{4, 5}
	tagRevokedFormulator        = []byte{5, 0}
	tagRevokedFormulatorNumber  = []byte{5, 1}
	tagRevokedFormulatorReverse = []byte{5, 2}
	tagRevokedFormulatorCount   = []byte{5, 3}
	tagRevokedHeight            = []byte{5, 4}
	tagUnstakingAmount          = []byte{6, 0}
	tagUnstakingAmountNumber    = []byte{6, 1}
	tagUnstakingAmountReverse   = []byte{6, 2}
	tagUnstakingAmountCount     = []byte{6, 3}
	tagRewardBaseUpgrade        = []byte{7, 0}
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
	binutil.BigEndian.PutUint32(bs[2:], Num)
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
	binutil.BigEndian.PutUint32(bs[2:], Num)
	return bs
}

func toRevokedFormulatorKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagRevokedFormulator)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toRevokedFormulatorNumberKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagRevokedFormulatorNumber)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toRevokedFormulatorReverseKey(height uint32, num uint32) []byte {
	bs := make([]byte, 10)
	copy(bs, tagRevokedFormulatorReverse)
	binutil.BigEndian.PutUint32(bs[2:], height)
	binutil.BigEndian.PutUint32(bs[6:], num)
	return bs
}

func toRevokedFormulatorCountKey(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagRevokedFormulatorCount)
	binutil.BigEndian.PutUint32(bs[2:], height)
	return bs
}

func toUnstakingAmountKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagUnstakingAmount)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toUnstakingAmountNumberKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagUnstakingAmountNumber)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toUnstakingAmountReverseKey(height uint32, num uint32) []byte {
	bs := make([]byte, 10)
	copy(bs, tagUnstakingAmountReverse)
	binutil.BigEndian.PutUint32(bs[2:], height)
	binutil.BigEndian.PutUint32(bs[6:], num)
	return bs
}

func toUnstakingAmountCountKey(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagUnstakingAmountCount)
	binutil.BigEndian.PutUint32(bs[2:], height)
	return bs
}
