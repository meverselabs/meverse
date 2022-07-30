package token

import "github.com/meverselabs/meverse/common"

var (
	tagTokenName        = byte(0x01)
	tagTokenSymbol      = byte(0x02)
	tagTokenMinter      = byte(0x03)
	tagTokenTotalSupply = byte(0x04)
	tagTokenGateway     = byte(0x05)
	tagTokenAmount      = byte(0x10)
	tagCollectedFee     = byte(0x11)
	tagTokenApprove     = byte(0x12)
	tagRouterAddress    = byte(0x13)
	tagRouterPaths      = byte(0x14)
	tagPause            = byte(0x15)
)

func MakeAllowanceTokenKey(sender common.Address) []byte {
	return makeTokenKey(sender, tagTokenApprove)
}
func makeTokenKey(sender common.Address, key byte) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = key
	copy(bs[1:], sender[:])
	return bs
}
