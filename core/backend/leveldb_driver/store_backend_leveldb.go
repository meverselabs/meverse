package leveldb_drvier

import (
	"bytes"
	"log"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/fletaio/fleta/core/backend"
)

func init() {
	backend.RegisterDriver("leveldb", NewStoreBackendLevelDB)
}

type StoreBackendLevelDB struct {
	db *leveldb.DB
}

func NewStoreBackendLevelDB(path string) (backend.StoreBackend, error) {
	start := time.Now()
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	log.Println("LevelDB is opened in", time.Now().Sub(start))
	back := &StoreBackendLevelDB{
		db: db,
	}
	return back, nil
}

func (st *StoreBackendLevelDB) Shrink() {
}

func (st *StoreBackendLevelDB) Close() {
	start := time.Now()
	st.db.Close()
	log.Println("LevelDB is closed in", time.Now().Sub(start))
}

func (st *StoreBackendLevelDB) View(fn func(txn backend.StoreReader) error) error {
	txn, err := st.db.OpenTransaction()
	if err != nil {
		return err
	}
	r := &storeBackendLevelDBTx{
		txn: txn,
	}
	if err := fn(r); err != nil {
		txn.Discard()
		return err
	}
	if err := txn.Commit(); err != nil {
		txn.Discard()
		return err
	}
	return nil
}

func (st *StoreBackendLevelDB) Update(fn func(txn backend.StoreWriter) error) error {
	txn, err := st.db.OpenTransaction()
	if err != nil {
		return err
	}
	r := &storeBackendLevelDBTx{
		txn: txn,
	}
	if err := fn(r); err != nil {
		txn.Discard()
		return err
	}
	if err := txn.Commit(); err != nil {
		txn.Discard()
		return err
	}
	return nil
}

type storeBackendLevelDBTx struct {
	txn *leveldb.Transaction
}

func (r *storeBackendLevelDBTx) Get(key []byte) ([]byte, error) {
	value, err := r.txn.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, backend.ErrNotExistKey
		} else {
			return nil, err
		}
	}
	return value, nil
}

func (r *storeBackendLevelDBTx) Iterate(prefix []byte, fn func(key []byte, value []byte) error) error {
	var rg *util.Range
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
		rg = &util.Range{Start: prefix, Limit: end}
	}
	it := r.txn.NewIterator(rg, nil)
	defer it.Release()
	for it.Next() {
		if err := fn(it.Key(), it.Value()); err != nil {
			return err
		}
	}
	return nil
}

func (r *storeBackendLevelDBTx) Set(key []byte, value []byte) error {
	if err := r.txn.Put(key, value, nil); err != nil {
		return err
	}
	return nil
}

func (r *storeBackendLevelDBTx) Delete(key []byte) error {
	if err := r.txn.Delete(key, nil); err != nil {
		return err
	}
	return nil
}
