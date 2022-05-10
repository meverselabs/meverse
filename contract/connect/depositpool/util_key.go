package depositpool

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
)

var (
	tagOwner         = byte(0x01)
	tagDepositToken  = byte(0x02)
	tagDepositAmount = byte(0x03)
	tagDepositLock   = byte(0x04)
	tagWithdrawLock  = byte(0x05)
	tagAccountLength = byte(0x06)
	tagAccountIndex  = byte(0x07)
	tagHolderList    = byte(0x08)
)

func makePoolKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

func makeAccountIndexKey(count *big.Int) []byte {
	return makePoolKey(tagAccountIndex, count.Bytes())

}
func makeHolderKey(addr common.Address) []byte {
	return makePoolKey(tagHolderList, addr[:])

}
