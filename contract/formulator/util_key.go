package formulator

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

var (
	tagTokenContractAddress = byte(0x01)
	tagFormulatorPolicy     = byte(0x02)
	tagRewardPolicy         = byte(0x03)
	tagFormulator           = byte(0x10)
	tagFormulatorNumber     = byte(0x11)
	tagFormulatorReverse    = byte(0x12)
	tagFormulatorCount      = byte(0x13)
	tagFormulatorSaleAmount = byte(0x14)
	tagFormulatorApprove    = byte(0x15)
	tagFormulatorApproveAll = byte(0x16)
	tagStakingAmount        = byte(0x20)
	tagStakingAmountNumber  = byte(0x21)
	tagStakingAmountReverse = byte(0x22)
	tagStakingAmountCount   = byte(0x23)
	tagStakingAmountMap     = byte(0x24)
	tagStakingPowerMap      = byte(0x25)
	tagStackRewardMap       = byte(0x31)
	tagUri                  = byte(0x41)
)

func makeFormulatorKey(key byte, addr common.Address) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = key
	copy(bs[1:], addr[:])
	return bs
}

func makeSaleAmountKey(addr common.Address) []byte {
	return makeFormulatorKey(tagFormulatorSaleAmount, addr)
}
func makeApproveKey(addr common.Address) []byte {
	return makeFormulatorKey(tagFormulatorApprove, addr)
}
func makeApproveAllKey(addr common.Address) []byte {
	return makeFormulatorKey(tagFormulatorApproveAll, addr)
}

func toStakingAmountKey(addr common.Address) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = tagStakingAmount
	copy(bs[1:], addr[:])
	return bs
}

func toStakingAmountNumberKey(addr common.Address) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = tagStakingAmountNumber
	copy(bs[1:], addr[:])
	return bs
}

func toStakingAmountReverseKey(Num uint32) []byte {
	bs := make([]byte, 6)
	bs[0] = tagStakingAmountReverse
	bin.PutUint32(bs[1:], Num)
	return bs
}

func toFormulatorNumberKey(addr common.Address) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = tagFormulatorNumber
	copy(bs[1:], addr[:])
	return bs
}

func toFormulatorReverseKey(Num uint32) []byte {
	bs := make([]byte, 6)
	bs[0] = tagFormulatorReverse
	bin.PutUint32(bs[1:], Num)
	return bs
}
