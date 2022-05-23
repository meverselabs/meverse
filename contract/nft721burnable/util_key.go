package nft721

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

var (
	tagOwner              = byte(0x01)
	tagName               = byte(0x02)
	tagSymbol             = byte(0x03)
	tagNFTMakeIndex       = byte(0x04)
	tagNFTCount           = byte(0x05)
	tagNFTIndex           = byte(0x06)
	tagIndexNFT           = byte(0x07)
	tagNFTOwner           = byte(0x08)
	tagTokenApprove       = byte(0x09)
	tagTokenApproveForAll = byte(0x10)
	tagBaseURI            = byte(0x11)
	tagTokenURI           = byte(0x12)
)

func makeNFTKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

func makeNFTIndexKey(k hash.Hash256) []byte {
	return makeNFTKey(tagNFTIndex, k[:])
}
func makeIndexNFTKey(i *big.Int) []byte {
	return makeNFTKey(tagIndexNFT, i.Bytes())
}
func makeNFTOwnerKey(i hash.Hash256) []byte {
	return makeNFTKey(tagNFTOwner, i.Bytes())
}
func makeTokenApproveKey(i hash.Hash256) []byte {
	return makeNFTKey(tagTokenApprove, i.Bytes())
}
func makeTokenApproveForAllKey(_owner common.Address, _operator common.Address) []byte {
	return makeNFTKey(tagTokenApproveForAll, append(_owner[:], _operator[:]...))
}
func makeTokenURIKey(tokenID hash.Hash256) []byte {
	return makeNFTKey(tagTokenURI, tokenID[:])
}
