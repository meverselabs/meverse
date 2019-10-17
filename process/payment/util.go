package payment

import (
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
)

// Topic returns the topic of the name
func Topic(Name string) uint64 {
	h := hash.Hash([]byte("fleta.payment#Topic#" + Name))
	return binutil.LittleEndian.Uint64(h[:])
}
