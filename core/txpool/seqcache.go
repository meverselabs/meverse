package txpool

import "github.com/fletaio/fleta/common"

// SeqCache is an interface that defines the function that acquire the last sequence of the address
type SeqCache interface {
	Seq(addr common.Address) uint64
}
