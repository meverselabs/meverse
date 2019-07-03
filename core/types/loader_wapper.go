package types

// LoaderWrapper is an interface to load state data from the target chain
type LoaderWrapper interface {
	Loader
	ProcessData(name []byte) []byte
	ProcessDataKeys(Prefix []byte) ([][]byte, error)
}
