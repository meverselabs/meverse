package types

import (
	"log"
	"sync"

	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
)

var lock sync.Mutex
var gDefineMap = map[uint16]string{}

// DefineHashedType is return string hashed type
func DefineHashedType(Name string) uint16 {
	lock.Lock()
	defer lock.Unlock()

	h := hash.DoubleHash([]byte(Name))
	t := binutil.LittleEndian.Uint16(h[:2])
	old, has := gDefineMap[t]
	if old == Name {
		return t
	}
	if has {
		panic("Type is collapsed (" + old + ", " + Name + ")")
	}
	gDefineMap[t] = Name
	log.Println("Type Defined", t, Name)
	return t
}

// NameOfHashedType returns the name of the hashed type
func NameOfHashedType(t uint16) string {
	lock.Lock()
	defer lock.Unlock()

	return gDefineMap[t]
}
