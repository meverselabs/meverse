package types

import "github.com/fletaio/fleta/common"

// LoaderWrapper is an interface to load state data from the target chain
type LoaderWrapper interface {
	Loader
	AccountData(addr common.Address, name []byte) []byte
	ProcessData(name []byte) []byte
}

// NewLoaderWrapper returns a LoaderWrapper
func NewLoaderWrapper(pid uint8, loader Loader) LoaderWrapper {
	if v, is := loader.(*ContextWrapper); is {
		return SwitchContextWrapper(pid, v)
	} else {
		return NewContextWrapper(pid, loader.(*Context))
	}
}
