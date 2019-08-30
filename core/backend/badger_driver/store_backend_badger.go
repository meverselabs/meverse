package badger_driver

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger"

	"github.com/fletaio/fleta/core/backend"
)

func init() {
	backend.RegisterDriver("badger", NewStoreBackendBadger)
}

type StoreBackendBadger struct {
	db *badger.DB
}

func NewStoreBackendBadger(path string) (backend.StoreBackend, error) {
	opts := badger.DefaultOptions(path)
	opts.Truncate = true
	opts.SyncWrites = true
	lockfilePath := filepath.Join(opts.Dir, "LOCK")
	os.MkdirAll(path, os.ModePerm)

	os.Remove(lockfilePath)

	start := time.Now()
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	log.Println("Badger is opened in", time.Now().Sub(start))
	back := &StoreBackendBadger{
		db: db,
	}
	return back, nil
}

func (st *StoreBackendBadger) Shrink() {
	MaxCount := 10
	Count := 0
	if st.db != nil {
	again:
		if err := st.db.RunValueLogGC(0.5); err != nil {
		} else {
			Count++
			if Count < MaxCount {
				goto again
			}
		}
	}
}

func (st *StoreBackendBadger) Close() {
	start := time.Now()
	MaxCount := 10
	Count := 0
again:
	if err := st.db.RunValueLogGC(0.9); err != nil {
	} else {
		Count++
		if Count < MaxCount {
			goto again
		}
	}
	st.db.Close()
	log.Println("Badger is closed in", time.Now().Sub(start))
}

func (st *StoreBackendBadger) View(fn func(txn backend.StoreReader) error) error {
	if err := st.db.View(func(txn *badger.Txn) error {
		r := &storeBackendBadgerTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

func (st *StoreBackendBadger) Update(fn func(txn backend.StoreWriter) error) error {
	if err := st.db.Update(func(txn *badger.Txn) error {
		r := &storeBackendBadgerTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

type storeBackendBadgerTx struct {
	txn *badger.Txn
}

func (r *storeBackendBadgerTx) Get(key []byte) ([]byte, error) {
	item, err := r.txn.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, backend.ErrNotExistKey
		} else {
			return nil, err
		}
	}
	if item.IsDeletedOrExpired() {
		return nil, backend.ErrNotExistKey
	}
	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (r *storeBackendBadgerTx) Iterate(prefix []byte, fn func(key []byte, value []byte) error) error {
	it := r.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	if len(prefix) > 0 {
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if !item.IsDeletedOrExpired() {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				if err := fn(item.KeyCopy(nil), value); err != nil {
					return err
				}
			}
		}
	} else {
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if !item.IsDeletedOrExpired() {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				if err := fn(item.KeyCopy(nil), value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *storeBackendBadgerTx) Set(key []byte, value []byte) error {
	if err := r.txn.Set(key, value); err != nil {
		return err
	}
	return nil
}

func (r *storeBackendBadgerTx) Delete(key []byte) error {
	if err := r.txn.Delete(key); err != nil {
		return err
	}
	return nil
}
