package gateway

import (
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
)

// tags
var (
	tagERC20TXID     = []byte{1, 0}
	tagOutTXID       = []byte{1, 1}
	tagPolicy        = []byte{2, 0}
	tagPlatform      = []byte{3, 0}
	tagPlatformIndex = []byte{3, 1}
	tagPlatformCount = []byte{3, 2}
)

func toERC20TXIDKey(Platform string, h hash.Hash256) []byte {
	bs := make([]byte, 2+len(Platform)+hash.Hash256Size)
	copy(bs, tagERC20TXID)
	copy(bs[2:], []byte(Platform))
	copy(bs[2+len(Platform):], h[:])
	return bs
}

func toOutTXIDKey(Platform string, TXID string) []byte {
	bs := make([]byte, 2+len(Platform)+len(TXID))
	copy(bs, tagOutTXID)
	copy(bs[2:], []byte(Platform))
	copy(bs[2+len(Platform):], []byte(TXID))
	return bs
}

func toPlatformKey(Platform string) []byte {
	bs := make([]byte, 2+len(Platform))
	copy(bs, tagPlatform)
	copy(bs[2:], []byte(Platform))
	return bs
}

func toPlatformIndexKey(index uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagPlatformIndex)
	copy(bs[2:], binutil.LittleEndian.Uint32ToBytes(index))
	return bs
}

func toPolicyKey(Platform string) []byte {
	bs := make([]byte, 2+len(Platform))
	copy(bs, tagPolicy)
	copy(bs[2:], []byte(Platform))
	return bs
}
