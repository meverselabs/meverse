package backend

type StoreBackend interface {
	Shrink()
	Close()
	View(fn func(txn StoreReader) error) error
	Update(fn func(txn StoreWriter) error) error
}

type StoreReader interface {
	Get(key []byte) ([]byte, error)
	Iterate(prefix []byte, fn func(key []byte, value []byte) error) error
}

type StoreWriter interface {
	StoreReader
	Set(key []byte, value []byte) error
	Delete(key []byte) error
}

type CreateBackend func(Paht string) (StoreBackend, error)

var gDriverMap = map[string]CreateBackend{}

func RegisterDriver(Name string, fn CreateBackend) {
	gDriverMap[Name] = fn
}

func Create(Name string, Path string) (StoreBackend, error) {
	fn, has := gDriverMap[Name]
	if !has {
		return nil, ErrNotExistDriver
	}
	return fn(Path)
}
