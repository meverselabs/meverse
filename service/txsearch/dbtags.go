package txsearch

import (
	"github.com/meverselabs/meverse/common/hash"
)

// tags
var (
	//init
	tagInitDB     = byte(0x00)
	tagInitHeight = byte(0x03)

	//process
	tagHeight = byte(0x10)

	//block
	tagBlockHash   = byte(0x20)
	tagEvent       = byte(0x21)
	tagEventReward = byte(0x22)

	//transaction
	tagID       = byte(0x30)
	tagTxHash   = byte(0x31)
	tagDefault  = byte(0x32)
	tagTransfer = byte(0x33)
	tagFail     = byte(0x34)
	tagAddress  = byte(0x35)
	// tagContractUsers = byte(0x36)

	//gateway
	tagTokenOut   = byte(0x40)
	tagTokenLeave = byte(0x41)

	//reward
	tagDailyReward = byte(0x50)

	//bridge
	tagBridge = byte(0x60)

	//etc
)

// func toKey(tag []byte, addr common.Address) []byte {
// 	bs := make([]byte, 2+common.AddressLength)
// 	copy(bs[:2], tag[:])
// 	copy(bs[2:], addr[:])
// 	return bs
// }

func toTxFailKey(hash hash.Hash256) []byte {
	bs := make([]byte, 1+len(hash))
	bs[0] = tagFail
	copy(bs[1:], hash[:])
	return bs
}
