package encoding

import (
	"sync"

	"github.com/fletaio/fleta/common/factory"
)

var lock sync.Mutex
var gFactoryMap = map[string]*factory.Factory{}

// Factory returns the factory of the name
func Factory(name string) *factory.Factory {
	lock.Lock()
	defer lock.Unlock()

	fc, has := gFactoryMap[name]
	if !has {
		fc = factory.NewFactory()
		gFactoryMap[name] = fc
	}
	return fc
}
