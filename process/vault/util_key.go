package vault

import (
	"bytes"

	"github.com/fletaio/fleta/common/util"

	"github.com/fletaio/fleta/common"
)

// tags
var (
	tagBalance          = []byte{1, 1}
	tagLockedBalance    = []byte{1, 2}
	tagLockedBalanceSum = []byte{1, 3}
)

func toLockedBalanceKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagLockedBalance)
	copy(bs[2:], util.Uint32ToBytes(height))
	copy(bs[6:], addr[:])
	return bs
}

func toLockedBalancePrefix(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagLockedBalance)
	copy(bs[2:], util.Uint32ToBytes(height))
	return bs
}

func fromLockedBalancePrefix(bs []byte) (common.Address, bool) {
	if bytes.HasPrefix(bs, tagLockedBalance) {
		var addr common.Address
		copy(addr[:], bs[6:])
		return addr, true
	} else {
		return common.Address{}, false
	}
}

func toLockedBalanceSumKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagLockedBalanceSum)
	copy(bs[2:], addr[:])
	return bs
}
