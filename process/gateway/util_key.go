package gateway

import (
	"github.com/fletaio/fleta/common/hash"
)

// tags
var (
	tagERC20TXID = []byte{1, 0}
	tagOutTXID   = []byte{1, 1}
	tagPolicy    = []byte{2, 0}
)

func toERC20TXIDKey(h hash.Hash256) []byte {
	bs := make([]byte, 2+hash.Hash256Size)
	copy(bs, tagERC20TXID)
	copy(bs[2:], h[:])
	return bs
}

func toOutTXIDKey(TXID string) []byte {
	bs := make([]byte, 2+len(TXID))
	copy(bs, tagOutTXID)
	copy(bs[2:], []byte(TXID))
	return bs
}
