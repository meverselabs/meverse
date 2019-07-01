package types

import (
	"bytes"
	"strings"

	"github.com/fletaio/fleta/common"
	"github.com/petar/GoLLRB/llrb"
)

var (
	ninf = nInf{}
	pinf = pInf{}
)

type nInf struct{}

func (nInf) Less(llrb.Item) bool {
	return true
}

type pInf struct{}

func (pInf) Less(llrb.Item) bool {
	return false
}

func cmpAddressASC(a interface{}, b interface{}) bool {
	ai := a.(common.Address)
	bi := b.(common.Address)
	return bytes.Compare(ai[:], bi[:]) < 0
}

func cmpStringASC(a interface{}, b interface{}) bool {
	ai := a.(string)
	bi := b.(string)
	return strings.Compare(ai[:], bi[:]) < 0
}

func cmpUint64ASC(a interface{}, b interface{}) bool {
	ai := a.(uint64)
	bi := b.(uint64)
	return ai < bi
}
