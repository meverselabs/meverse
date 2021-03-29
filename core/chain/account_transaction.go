package chain

import "github.com/fletaio/fleta/common"

// AccountTransaction defines common functions of account model based transactions
type AccountTransaction interface {
	Seq() uint64
	From() common.Address
}
