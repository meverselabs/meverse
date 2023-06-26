package testlib

import "github.com/ethereum/go-ethereum/common"

// HaveSameElements tells whether a and b contain the same elements.
func HaveSameElements(a, b []common.Address) bool {
	if len(a) != len(b) {
		return false
	}
	for _, va := range a {
		exist := false
		for _, vb := range b {
			if va == vb {
				exist = true
				break
			}
		}
		if !exist {
			return false
		}
	}
	return true
}
