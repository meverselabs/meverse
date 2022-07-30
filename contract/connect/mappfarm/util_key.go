package mappfarm

import (
	"github.com/meverselabs/meverse/common"
)

var (
	tagOwner         = byte(0x01)
	tagBanker        = byte(0x02)
	tagFarmToken     = byte(0x03)
	tagOwnerReward   = byte(0x04)
	tagTokenPerBlock = byte(0x06)
	tagStartBlock    = byte(0x07)

	// tagTotalAllocPoint = byte(0x08)

	tagPoolInfo = byte(0x09)
	// tagPoolLength = byte(0x10)
	tagUserInfo = byte(0x11)
)

func makeFarmKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}
func makeUserInfoKey(user common.Address) []byte {
	return makeFarmKey(tagUserInfo, user[:])
}
