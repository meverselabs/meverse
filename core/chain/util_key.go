package chain

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
)

var (
	tagHeight       = byte(0x01)
	tagHeightHash   = byte(0x02)
	tagPoFRankTable = byte(0x12)
	tagAdmin        = byte(0x20)
	tagAddressSeq   = byte(0x21)
	tagGenerator    = byte(0x30)
	tagContract     = byte(0x40)
	tagData         = byte(0x50)
	tagBlockGen     = byte(0x60)
	tagMainToken    = byte(0x70)
	tagBasicFee     = byte(0x80)
)

func toHeightHashKey(height uint32) []byte {
	bs := make([]byte, 5)
	bs[0] = tagHeightHash
	bin.PutUint32(bs[1:], height)
	return bs
}

func toAdminKey(addr common.Address) []byte {
	bs := make([]byte, 1+len(addr[:]))
	bs[0] = tagAdmin
	copy(bs[1:], addr[:])
	return bs
}

func toAddressSeqKey(addr common.Address) []byte {
	bs := make([]byte, 1+len(addr[:]))
	bs[0] = tagAddressSeq
	copy(bs[1:], addr[:])
	return bs
}

func toGeneratorKey(addr common.Address) []byte {
	bs := make([]byte, 1+len(addr[:]))
	bs[0] = tagGenerator
	copy(bs[1:], addr[:])
	return bs
}

func toContractKey(addr common.Address) []byte {
	bs := make([]byte, 1+len(addr[:]))
	bs[0] = tagContract
	copy(bs[1:], addr[:])
	return bs
}

func toDataKey(key string) []byte {
	bs := make([]byte, 1+len(key))
	bs[0] = tagData
	copy(bs[1:], []byte(key))
	return bs
}

func toBlockGenKey(addr common.Address) []byte {
	bs := make([]byte, 1+len(addr[:]))
	bs[0] = tagBlockGen
	copy(bs[1:], addr[:])
	return bs
}

func fromBlockGenKey(bs []byte) common.Address {
	var addr common.Address
	copy(addr[:], bs[1:])
	return addr
}
