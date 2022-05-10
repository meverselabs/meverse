package factory

import (
	"github.com/meverselabs/meverse/common"
)

var (
	tagOwner    = byte(0x00)
	tagAllPairs = byte(0x01)
)

//makePairKey two Token Address -> bytes key
func makePairKey(token0, token1 common.Address) []byte {
	base := make([]byte, common.AddressLength*2)
	copy(base[0:], token0[:])
	copy(base[common.AddressLength:], token1[:])
	return base
}
