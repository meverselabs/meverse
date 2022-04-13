package token

import "github.com/fletaio/fleta_v2/common"

var (
	tagTokenName        = byte(0x01)
	tagTokenSymbol      = byte(0x02)
	tagTokenMinter      = byte(0x03)
	tagTokenTotalSupply = byte(0x04)
	tagTokenGateway     = byte(0x05)
	tagTokenAmount      = byte(0x10)
	tagCollectedFee     = byte(0x11)
	tagTokenApprove     = byte(0x12)
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
