package vault

import (
	"encoding/binary"

	"github.com/fletaio/fleta/common"
)

// tags
var (
	tagBalance              = []byte{1, 1}
	tagLockedBalance        = []byte{2, 2}
	tagLockedBalanceNumber  = []byte{2, 3}
	tagLockedBalanceReverse = []byte{2, 4}
	tagLockedBalanceCount   = []byte{2, 5}
	tagLockedBalanceSum     = []byte{2, 6}
	tagCollectedFee         = []byte{3, 1}
)

func toLockedBalanceKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagLockedBalance)
	binary.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toLockedBalanceNumberKey(height uint32, addr common.Address) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagLockedBalanceNumber)
	binary.BigEndian.PutUint32(bs[2:], height)
	copy(bs[6:], addr[:])
	return bs
}

func toLockedBalanceReverseKey(height uint32, num uint32) []byte {
	bs := make([]byte, 10)
	copy(bs, tagLockedBalanceReverse)
	binary.BigEndian.PutUint32(bs[2:], height)
	binary.BigEndian.PutUint32(bs[6:], num)
	return bs
}

func toLockedBalanceCountKey(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagLockedBalanceCount)
	binary.BigEndian.PutUint32(bs[2:], height)
	return bs
}

func toLockedBalanceSumKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagLockedBalanceSum)
	copy(bs[2:], addr[:])
	return bs
}
