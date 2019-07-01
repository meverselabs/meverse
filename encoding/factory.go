package encoding

import (
	"github.com/fletaio/fleta/common/factory"
)

var gFactoryMap = map[string]*factory.Factory{}

// Factory returns the factory of the name
func Factory(name string) *factory.Factory {
	fc, has := gFactoryMap[name]
	if !has {
		fc = factory.NewFactory()
		gFactoryMap[name] = fc
	}
	return fc
}
