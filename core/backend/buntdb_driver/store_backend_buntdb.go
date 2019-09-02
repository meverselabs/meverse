package buntdb_driver

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fletaio/fleta/core/backend"
	"github.com/fletaio/fleta/core/backend/buntdb_driver/buntdb"
)

func init() {
	backend.RegisterDriver("buntdb", NewStoreBackendBuntDB)
}

type StoreBackendBuntDB struct {
	sync.Mutex
	db *buntdb.DB
}

func NewStoreBackendBuntDB(path string) (backend.StoreBackend, error) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	start := time.Now()
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}
	log.Println("BuntDB is opened in", time.Now().Sub(start))
	back := &StoreBackendBuntDB{
		db: db,
	}
	return back, nil
}

func (st *StoreBackendBuntDB) Shrink() {
	st.Lock()
	defer st.Unlock()

	defer func() {
		recover()
	}()
	st.db.Shrink()
}

func (st *StoreBackendBuntDB) Backup(dst string) (ret error) {
	st.Lock()
	defer st.Unlock()

	defer func() {
		if e := recover(); e != nil {
			ret = e.(error)
		}
	}()
	if err := st.db.Backup(dst); err != nil {
		return err
	}
	return
}

func (st *StoreBackendBuntDB) Close() {
	st.Lock()
	defer st.Unlock()

	defer func() {
		recover()
		st.db.Close()
	}()

	start := time.Now()
	st.db.Shrink()
	st.db.Close()
	log.Println("BuntDB is closed in", time.Now().Sub(start))
}

func (st *StoreBackendBuntDB) View(fn func(txn backend.StoreReader) error) error {
	if err := st.db.View(func(txn *buntdb.Tx) error {
		r := &storeBackendBuntDBTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

func (st *StoreBackendBuntDB) Update(fn func(txn backend.StoreWriter) error) error {
	if err := st.db.Update(func(txn *buntdb.Tx) error {
		r := &storeBackendBuntDBTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

type storeBackendBuntDBTx struct {
	txn *buntdb.Tx
}

func (r *storeBackendBuntDBTx) Get(key []byte) ([]byte, error) {
	value, err := r.txn.Get(string(key))
	if err != nil {
		if err == buntdb.ErrNotFound {
			return nil, backend.ErrNotExistKey
		} else {
			return nil, err
		}
	}
	return []byte(value), nil
}

func (r *storeBackendBuntDBTx) Iterate(prefix []byte, fn func(key []byte, value []byte) error) error {
	if len(prefix) > 0 {
		end := make([]byte, len(prefix))
		copy(end, prefix)
		for i := len(end) - 1; i >= 0; i-- {
			end[i]++
			if end[i] != 0 {
				break
			}
		}
		if bytes.Compare(prefix, end) > 0 {
			return nil
		}
		var inErr error
		r.txn.AscendRange("", string(prefix), string(end), func(key string, value string) bool {
			if err := fn([]byte(key), []byte(value)); err != nil {
				inErr = err
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	} else {
		var inErr error
		r.txn.Ascend("", func(key string, value string) bool {
			if err := fn([]byte(key), []byte(value)); err != nil {
				inErr = err
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	}
	return nil
}

func (r *storeBackendBuntDBTx) Set(key []byte, value []byte) error {
	if _, _, err := r.txn.Set(string(key), string(value), nil); err != nil {
		return err
	}
	return nil
}

func (r *storeBackendBuntDBTx) Delete(key []byte) error {
	if _, err := r.txn.Delete(string(key)); err != nil {
		if err == buntdb.ErrNotFound {
			return nil
		} else {
			return err
		}
	}
	return nil
}
