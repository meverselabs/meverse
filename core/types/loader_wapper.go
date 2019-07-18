package types

import "github.com/fletaio/fleta/common"

// SwitchLoaderWrapper returns a SwitchLoaderWrapper of the pid
func SwitchLoaderWrapper(pid uint8, lw LoaderWrapper) LoaderWrapper {
	return SwitchContextWrapper(pid, lw.(*ContextWrapper))
}

// LoaderWrapper is an interface to load state data from the target chain
type LoaderWrapper interface {
	Loader
	AccountData(addr common.Address, name []byte) []byte
	AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error)
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
