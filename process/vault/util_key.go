package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
)

// tags
var (
	tagBalance              = []byte{1, 1}
	tagLockedBalance        = []byte{2, 1}
	tagLockedBalanceNumber  = []byte{2, 2}
	tagLockedBalanceReverse = []byte{2, 3}
	tagLockedBalanceCount   = []byte{2, 4}
	tagLockedBalanceSum     = []byte{2, 5}
	tagCollectedFee         = []byte{3, 1}
	tagPolicy               = []byte{4, 0}
)

func toLockedBalanceKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagLockedBalance)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toLockedBalanceNumberKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagLockedBalanceNumber)
	binutil.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toLockedBalanceReverseKey(height uint32, num uint32) []byte {
	bs := make([]byte, 10)
	copy(bs, tagLockedBalanceReverse)
	binutil.BigEndian.PutUint32(bs[2:], height)
	binutil.BigEndian.PutUint32(bs[6:], num)
	return bs
}

func toLockedBalanceCountKey(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagLockedBalanceCount)
	binutil.BigEndian.PutUint32(bs[2:], height)
	return bs
}
