package gateway

import (
	"github.com/fletaio/fleta/common/hash"
)

// tags
var (
	tagTokenTXID   = []byte{1, 0}
	tagOutCoinTXID = []byte{1, 1}
	tagPolicy      = []byte{2, 0}
)

func toTokenTXIDKey(TokenPlatform string, h hash.Hash256) []byte {
	bs := make([]byte, 2+hash.Hash256Size+len(TokenPlatform))
	copy(bs, tagTokenTXID)
	copy(bs[2:], h[:])
	copy(bs[2+hash.Hash256Size:], []byte(TokenPlatform))
	return bs
}

func toOutCoinTXIDKey(TXID string) []byte {
	bs := make([]byte, 2+len(TXID))
	copy(bs, tagOutCoinTXID)
	copy(bs[2:], []byte(TXID))
	return bs
}

func toPolicyKey(TokenPlatform string) []byte {
	bs := make([]byte, 2+len(TokenPlatform))
	copy(bs, tagPolicy)
	copy(bs[2:], []byte(TokenPlatform))
	return bs
}
