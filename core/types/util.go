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

// IsAllowedAccountName returns it is allowed account name or not
func IsAllowedAccountName(Name string) bool {
	if len(Name) < 8 || len(Name) > 16 {
		return false
	}
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) && c != '.' && c != '-' && c != '_' {
			return false
		}
	}
	return true
}
