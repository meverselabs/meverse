package types

// LoaderProcess is an interface to load state data from the target chain
type LoaderProcess interface {
	Loader
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
