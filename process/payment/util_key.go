package payment

import (
	"github.com/fletaio/fleta/common/binutil"
)

// tags
var (
	tagRequestPayment = []byte{1, 0}
	tagTopic          = []byte{2, 0}
)

func toRequestPaymentKey(TXID string) []byte {
	bs := make([]byte, 2+len(TXID))
	copy(bs, tagRequestPayment)
	copy(bs[2:], []byte(TXID))
	return bs
}

func toTopicKey(topic uint64) []byte {
	bs := make([]byte, 10)
	copy(bs, tagTopic)
	binutil.BigEndian.PutUint64(bs[2:], topic)
	return bs
}
