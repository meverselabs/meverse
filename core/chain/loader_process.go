package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// LoaderProcess is an interface to load state data from the target chain
type LoaderProcess interface {
	types.Loader
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
