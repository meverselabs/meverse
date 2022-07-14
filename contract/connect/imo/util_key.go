package imo

import "github.com/meverselabs/meverse/common"

var (
	tagProjectOwner     = byte(0x01)
	tagPayToken         = byte(0x02)
	tagProjectToken     = byte(0x03)
	tagProjectOffering  = byte(0x04)
	tagProjectRaising   = byte(0x05)
	tagPayLimit         = byte(0x06)
	tagStartBlock       = byte(0x07)
	tagEndBlock         = byte(0x08)
	tagHarvestFeeFactor = byte(0x09)
	tagWhiteListAddress = byte(0x10)
	tagWhiteListGroupId = byte(0x11)
	tagTotalAmount      = byte(0x12)
	tagUserInfo         = byte(0x13)
	tagAddressList      = byte(0x14)
)

func makeImoKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

func makeUserInfoKey(addr common.Address) []byte {
	return makeImoKey(tagUserInfo, addr[:])
}
