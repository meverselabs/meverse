package types

import (
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
)

var gDefineMap = map[uint16]string{}

// DefineHashedType is return string hashed type
func DefineHashedType(Name string) uint16 {
	h := hash.DoubleHash([]byte(Name))
	t := util.BytesToUint16(h[:2])
	old, has := gDefineMap[t]
	if has {
		panic("Type is collapsed (" + old + ", " + Name + ")")
	}
	gDefineMap[t] = Name
	return t
}

// NameOfHashedType returns the name of the hashed type
func NameOfHashedType(t uint16) string {
	return gDefineMap[t]
}
