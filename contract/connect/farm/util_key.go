package farm

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

var (
	tagOwner          = byte(0x01)
	tagFarmToken      = byte(0x02)
	tagOwnerReward    = byte(0x03)
	tagTokenMaxSupply = byte(0x04)
	tagTokenPerBlock  = byte(0x05)
	tagStartBlock     = byte(0x06)

	tagTotalAllocPoint = byte(0x07)

	tagPoolInfo   = byte(0x08)
	tagPoolLength = byte(0x09)
	tagUserInfo   = byte(0x10)

	tagHoldShares       = byte(0x11)
	tagHoldSharesHeight = byte(0x12)
)

func makeFarmKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}
func makePoolInfoKey(pid uint64) []byte {
	return makeFarmKey(tagPoolInfo, bin.Uint64Bytes(pid))
}
func makeUserInfoKey(pid uint64, user common.Address) []byte {
	bs := append(bin.Uint64Bytes(pid), user[:]...)
	return makeFarmKey(tagUserInfo, bs)
}
