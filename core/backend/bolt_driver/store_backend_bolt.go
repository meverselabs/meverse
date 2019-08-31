package bolt_driver

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"

	"github.com/fletaio/fleta/core/backend"
)

func init() {
	backend.RegisterDriver("bolt", NewStoreBackendBolt)
}

type StoreBackendBolt struct {
	db *bolt.DB
}

func NewStoreBackendBolt(path string) (backend.StoreBackend, error) {
	os.MkdirAll(path, os.ModePerm)

	start := time.Now()
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(txn *bolt.Tx) error {
		txn.CreateBucketIfNotExists([]byte{0})
		return nil
	}); err != nil {
		return nil, err
	}
	log.Println("Bolt is opened in", time.Now().Sub(start))
	back := &StoreBackendBolt{
		db: db,
	}
	return back, nil
}

func (st *StoreBackendBolt) Shrink() {
}

func (st *StoreBackendBolt) Close() {
	start := time.Now()
	st.db.Close()
	log.Println("Bolt is closed in", time.Now().Sub(start))
}

func (st *StoreBackendBolt) View(fn func(txn backend.StoreReader) error) error {
	if err := st.db.View(func(txn *bolt.Tx) error {
		r := &StoreBackendBoltTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

func (st *StoreBackendBolt) Update(fn func(txn backend.StoreWriter) error) error {
	if err := st.db.Update(func(txn *bolt.Tx) error {
		r := &StoreBackendBoltTx{
			txn: txn,
		}
		return fn(r)
	}); err != nil {
		return err
	}
	return nil
}

type StoreBackendBoltTx struct {
	txn *bolt.Tx
}

func (r *StoreBackendBoltTx) Get(key []byte) ([]byte, error) {
	bucket := r.txn.Bucket([]byte{0})
	value := bucket.Get(key)
	if len(value) == 0 {
		return nil, backend.ErrNotExistKey
	}
	return value, nil
}

func (r *StoreBackendBoltTx) Iterate(prefix []byte, fn func(key []byte, value []byte) error) error {
	bucket := r.txn.Bucket([]byte{0})
	c := bucket.Cursor()
	if len(prefix) > 0 {
		for key, value := c.Seek(prefix); key != nil && bytes.HasPrefix(key, prefix); key, value = c.Next() {
			if err := fn([]byte(key), []byte(value)); err != nil {
				return err
			}
		}
	} else {
		for key, value := c.First(); key != nil; key, value = c.Next() {
			if err := fn([]byte(key), []byte(value)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *StoreBackendBoltTx) Set(key []byte, value []byte) error {
	bucket := r.txn.Bucket([]byte{0})
	if err := bucket.Put(key, value); err != nil {
		return err
	}
	return nil
}

func (r *StoreBackendBoltTx) Delete(key []byte) error {
	bucket := r.txn.Bucket([]byte{0})
	if err := bucket.Delete(key); err != nil {
		return err
	}
	return nil
}
