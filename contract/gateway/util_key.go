package gateway

import (
	"github.com/meverselabs/meverse/common/hash"
)

var (
	tagTokenContractAddress = byte(0x01)
	tagTokenIn              = byte(0x02)
	tagTokenOut             = byte(0x03)
	tagTokenSender          = byte(0x04)
	tagPlatform             = byte(0x05)
	tagPlatformFee          = byte(0x06)
	tagTokenInRevert        = byte(0x07)
	tagFeeOwner             = byte(0x08)
)

func makeGatewayKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

func makeTokenInKey(h hash.Hash256, Platform string) []byte {
	return makeGatewayKey(tagTokenIn, append(h[:], []byte(Platform)...))
}
func makeTokenOutKey(txid []byte, Platform string) []byte {
	bs := append(txid, []byte(Platform)...)
	return makeGatewayKey(tagTokenOut, bs)
}
func makeTokenInRevertKey(txids []byte) []byte {
	return makeGatewayKey(tagTokenInRevert, txids)
}

func makePlatformKey(Platform string) []byte {
	return makeGatewayKey(tagPlatform, []byte(Platform))
}
func makePlatformFeeKey(Platform string) []byte {
	return makeGatewayKey(tagPlatformFee, []byte(Platform))
}
