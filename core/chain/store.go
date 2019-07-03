package chain

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fletaio/fleta/encoding"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/types"
)

// Store saves the target chain state
// All updates are executed in one transaction with FileSync option
type Store struct {
	sync.Mutex
	db         *badger.DB
	name       string
	version    uint16
	SeqMapLock sync.Mutex
	SeqMap     map[common.Address]uint64
	cache      storecache
	ticker     *time.Ticker
	closeLock  sync.RWMutex
	isClose    bool
}

type storecache struct {
	cached      bool
	height      uint32
	heightHash  hash.Hash256
	heightBlock *types.Block
}

// NewStore returns a Store
func NewStore(path string, name string, version uint16, bRecover bool) (*Store, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.Truncate = bRecover
	opts.SyncWrites = true
	lockfilePath := filepath.Join(opts.Dir, "LOCK")
	os.MkdirAll(path, os.ModeDir)

	os.Remove(lockfilePath)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	{
	again:
		if err := db.RunValueLogGC(0.7); err != nil {
		} else {
			goto again
		}
	}

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
		again:
			if err := db.RunValueLogGC(0.7); err != nil {
			} else {
				goto again
			}
		}
	}()

	return &Store{
		db:      db,
		ticker:  ticker,
		name:    name,
		version: version,
		SeqMap:  map[common.Address]uint64{},
	}, nil
}

// Close terminate and clean store
func (st *Store) Close() {
	st.closeLock.Lock()
	defer st.closeLock.Unlock()

	st.isClose = true
	st.db.Close()
	st.ticker.Stop()
	st.db = nil
	st.ticker = nil
}

// Name returns the name of the target chain
func (st *Store) Name() string {
	return st.name
}

// Version returns the version of the target chain
func (st *Store) Version() uint16 {
	return st.version
}

// TargetHeight returns the target height of the target chain
func (st *Store) TargetHeight() uint32 {
	return st.Height() + 1
}

// LastHash returns the last hash of the chain
func (st *Store) LastHash() hash.Hash256 {
	h, err := st.Hash(st.Height())
	if err != nil {
		if err != ErrStoreClosed {
			// should have not reabh
			panic(err)
		}
		return hash.Hash256{}
	}
	return h
}

// LastTimestamp returns the last timestamp of the chain
func (st *Store) LastTimestamp() uint64 {
	if st.Height() == 0 {
		return 0
	}
	bh, err := st.Header(st.Height())
	if err != nil {
		if err != ErrStoreClosed {
			// should have not reabh
			panic(err)
		}
		return 0
	}
	return bh.Timestamp
}

// Hash returns the hash of the data by height
func (st *Store) Hash(height uint32) (hash.Hash256, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return hash.Hash256{}, ErrStoreClosed
	}

	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightHash, nil
		}
	}

	var h hash.Hash256
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toHeightHashKey(height))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		copy(h[:], value)
		return nil
	}); err != nil {
		return hash.Hash256{}, err
	}
	return h, nil
}

// Header returns the header of the data by height
func (st *Store) Header(height uint32) (*types.Header, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	if height < 1 {
		return nil, ErrNotExistKey
	}
	if st.cache.cached {
		if st.cache.height == height {
			return &st.cache.heightBlock.Header, nil
		}
	}

	var bh types.Header
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toHeightHeaderKey(height))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if err := encoding.Unmarshal(value, &bh); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &bh, nil
}

// Block returns the block by height
func (st *Store) Block(height uint32) (*types.Block, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	if height < 1 {
		return nil, ErrNotExistKey
	}
	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightBlock, nil
		}
	}

	var b types.Block
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toHeightBlockKey(height))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if err := encoding.Unmarshal(value, &b); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &b, nil
}

// Height returns the current height of the target chain
func (st *Store) Height() uint32 {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return 0
	}

	if st.cache.cached {
		return st.cache.height
	}

	var height uint32
	st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("height"))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		height = util.BytesToUint32(value)
		return nil
	})
	return height
}

// Accounts returns all accounts in the store
func (st *Store) Accounts() ([]types.Account, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	fc := encoding.Factory("account")
	list := []types.Account{}
	if err := st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(tagAccount); it.ValidForPrefix(tagAccount); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			acc, err := fc.Create(util.BytesToUint16(value))
			if err != nil {
				return err
			}
			if err := encoding.Unmarshal(value[2:], &acc); err != nil {
				return err
			}
			list = append(list, acc.(types.Account))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// Seq returns the sequence of the transaction
func (st *Store) Seq(addr common.Address) uint64 {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return 0
	}

	st.SeqMapLock.Lock()
	defer st.SeqMapLock.Unlock()

	if seq, has := st.SeqMap[addr]; has {
		return seq
	} else {
		var seq uint64
		if err := st.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(toAccountSeqKey(addr))
			if err != nil {
				return err
			}
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			seq = util.BytesToUint64(value)
			return nil
		}); err != nil {
			return 0
		}
		st.SeqMap[addr] = seq
		return seq
	}
}

// Account returns the account instance of the address from the store
func (st *Store) Account(addr common.Address) (types.Account, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	fc := encoding.Factory("account")

	var acc types.Account
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountKey(addr))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		v, err := fc.Create(util.BytesToUint16(value))
		if err != nil {
			return err
		}
		if err := encoding.Unmarshal(value[2:], &v); err != nil {
			return err
		}
		acc = v.(types.Account)
		return nil
	}); err != nil {
		if err == ErrNotExistKey {
			return nil, types.ErrNotExistAccount
		} else {
			return nil, err
		}
	}
	return acc, nil
}

// AddressByName returns the account instance of the name from the store
func (st *Store) AddressByName(Name string) (common.Address, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return common.Address{}, ErrStoreClosed
	}

	var addr common.Address
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountNameKey(Name))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		copy(addr[:], value)
		return nil
	}); err != nil {
		if err == ErrNotExistKey {
			return common.Address{}, types.ErrNotExistAccount
		} else {
			return common.Address{}, err
		}
	}
	return addr, nil
}

// HasAccount bhecks that the account of the address is exist or not
func (st *Store) HasAccount(addr common.Address) (bool, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false, ErrStoreClosed
	}

	var Has bool
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountKey(addr))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		Has = !item.IsDeletedOrExpired()
		return nil
	}); err != nil {
		if err == ErrNotExistKey {
			return false, nil
		} else {
			return false, err
		}
	}
	return Has, nil
}

// HasAccountName bhecks that the account of the name is exist or not
func (st *Store) HasAccountName(Name string) (bool, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false, ErrStoreClosed
	}

	var Has bool
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountNameKey(Name))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		Has = !item.IsDeletedOrExpired()
		return nil
	}); err != nil {
		if err == ErrNotExistKey {
			return false, nil
		} else {
			return false, err
		}
	}
	return Has, nil
}

// AccountDataKeys returns all data keys of the account in the store
func (st *Store) AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	list := [][]byte{}
	if err := st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		pre := toAccountDataKey(string(addr[:]))
		if len(Prefix) > 0 {
			pre = append(pre, Prefix...)
		}
		for it.Seek(pre); it.ValidForPrefix(pre); it.Next() {
			item := it.Item()
			key := item.Key()
			list = append(list, key[len(pre):])
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// AccountData returns the account data from the store
func (st *Store) AccountData(addr common.Address, name []byte) []byte {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil
	}

	key := string(addr[:]) + string(name)
	var data []byte
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountDataKey(key))
		if err != nil {
			return err
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		data = value
		return nil
	}); err != nil {
		return nil
	}
	return data
}

// UTXOs returns all UTXOs in the store
func (st *Store) UTXOs() ([]*types.UTXO, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	list := []*types.UTXO{}
	if err := st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(tagUTXO); it.ValidForPrefix(tagUTXO); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			utxo := &types.UTXO{
				TxIn:  types.NewTxIn(fromUTXOKey(item.Key())),
				TxOut: types.NewTxOut(),
			}
			if err := encoding.Unmarshal(value, &(utxo.TxOut)); err != nil {
				return err
			}
			list = append(list, utxo)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// HasUTXO bhecks that the utxo of the id is exist or not
func (st *Store) HasUTXO(id uint64) (bool, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false, ErrStoreClosed
	}

	var Has bool
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toUTXOKey(id))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		Has = !item.IsDeletedOrExpired()
		return nil
	}); err != nil {
		if err == ErrNotExistKey {
			return false, nil
		} else {
			return false, err
		}
	}
	return Has, nil
}

// UTXO returns the UTXO from the top store
func (st *Store) UTXO(id uint64) (*types.UTXO, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	var utxo *types.UTXO
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toUTXOKey(id))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return types.ErrNotExistUTXO
			} else {
				return err
			}
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		utxo = &types.UTXO{
			TxIn:  types.NewTxIn(id),
			TxOut: types.NewTxOut(),
		}
		if err := encoding.Unmarshal(value, &(utxo.TxOut)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return utxo, nil
}

// ProcessDataKeys returns all data keys of the process in the store
func (st *Store) ProcessDataKeys(pid uint8, Prefix []byte) ([][]byte, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	list := [][]byte{}
	if err := st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		pre := toProcessDataKey(string(pid))
		if len(Prefix) > 0 {
			pre = append(pre, Prefix...)
		}
		for it.Seek(pre); it.ValidForPrefix(pre); it.Next() {
			item := it.Item()
			key := item.Key()
			list = append(list, key[len(pre):])
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// ProcessData returns the process data from the store
func (st *Store) ProcessData(pid uint8, name []byte) []byte {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil
	}

	key := string(pid) + string(name)
	var data []byte
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toProcessDataKey(key))
		if err != nil {
			return err
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		data = value
		return nil
	}); err != nil {
		return nil
	}
	return data
}

// Events returns all events by conditions
func (st *Store) Events(From uint32, To uint32) ([]types.Event, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, ErrStoreClosed
	}

	fc := encoding.Factory("event")

	list := []types.Event{}
	if err := st.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		tagBegin := toEventKey(types.MarshalID(From, 0, 0))
		tagEnd := toEventKey(types.MarshalID(To, 65535, 65535))
		for it.Seek(tagBegin); it.ValidForPrefix(tagEnd); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ev, err := fc.Create(util.BytesToUint16(value))
			if err != nil {
				return err
			}
			if err := encoding.Unmarshal(value[2:], &ev); err != nil {
				return err
			}
			list = append(list, ev.(types.Event))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// StoreGenesis stores the genesis data
func (st *Store) StoreGenesis(genHash hash.Hash256, ctd *types.ContextData) error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return ErrStoreClosed
	}

	if st.Height() > 0 {
		return ErrAlreadyGenesised
	}
	if err := st.db.Update(func(txn *badger.Txn) error {
		{
			if err := txn.Set(toHeightHashKey(0), genHash[:]); err != nil {
				return err
			}
			bsHeight := util.Uint32ToBytes(0)
			if err := txn.Set(toHashHeightKey(genHash), bsHeight); err != nil {
				return err
			}
			if err := txn.Set([]byte("height"), bsHeight); err != nil {
				return err
			}
		}
		if err := applyContextData(txn, ctd); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	st.cache.height = 0
	st.cache.heightHash = genHash
	st.cache.heightBlock = nil
	st.cache.cached = true
	return nil
}

// StoreBlock stores the block
func (st *Store) StoreBlock(b *types.Block, ctd *types.ContextData) error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return ErrStoreClosed
	}

	DataHash := encoding.Hash(b.Header)
	if err := st.db.Update(func(txn *badger.Txn) error {
		{
			data, err := encoding.Marshal(b)
			if err != nil {
				return err
			}
			if err := txn.Set(toHeightBlockKey(b.Header.Height), data); err != nil {
				return err
			}
		}
		{
			data, err := encoding.Marshal(b.Header)
			if err != nil {
				return err
			}
			if err := txn.Set(toHeightHeaderKey(b.Header.Height), data); err != nil {
				return err
			}
		}
		{
			if err := txn.Set(toHeightHashKey(b.Header.Height), DataHash[:]); err != nil {
				return err
			}
			bsHeight := util.Uint32ToBytes(b.Header.Height)
			if err := txn.Set(toHashHeightKey(DataHash), bsHeight); err != nil {
				return err
			}
			if err := txn.Set([]byte("height"), bsHeight); err != nil {
				return err
			}
		}
		if err := applyContextData(txn, ctd); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	st.SeqMapLock.Lock()
	ctd.SeqMap.EachAll(func(addr common.Address, value uint64) bool {
		st.SeqMap[addr] = value
		return true
	})
	st.SeqMapLock.Unlock()
	st.cache.height = b.Header.Height
	st.cache.heightHash = DataHash
	st.cache.heightBlock = b
	st.cache.cached = true
	return nil
}

func applyContextData(txn *badger.Txn, ctd *types.ContextData) error {
	var inErr error
	ctd.SeqMap.EachAll(func(addr common.Address, value uint64) bool {
		if err := txn.Set(toAccountSeqKey(addr), util.Uint64ToBytes(value)); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	afc := encoding.Factory("account")
	ctd.AccountMap.EachAll(func(addr common.Address, acc types.Account) bool {
		t, err := afc.TypeOf(acc)
		if err != nil {
			inErr = err
			return false
		}
		var buffer bytes.Buffer
		buffer.Write(util.Uint16ToBytes(t))
		data, err := encoding.Marshal(acc)
		if err != nil {
			inErr = err
			return false
		}
		buffer.Write(data)
		if err := txn.Set(toAccountKey(addr), buffer.Bytes()); err != nil {
			inErr = err
			return false
		}
		if err := txn.Set(toAccountNameKey(acc.Name()), addr[:]); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.DeletedAccountMap.EachAll(func(addr common.Address, acc types.Account) bool {
		if err := txn.Delete(toAccountKey(addr)); err != nil {
			inErr = err
			return false
		}
		if err := txn.Delete(toAccountBalanceKey(addr)); err != nil {
			inErr = err
			return false
		}
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := toAccountDataKey(string(addr[:]))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := txn.Delete(item.Key()); err != nil {
				inErr = err
				return false
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.DeletedAccountNameMap.EachAll(func(key string, value bool) bool {
		if err := txn.Delete(toAccountNameKey(key)); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.AccountDataMap.EachAll(func(key string, value []byte) bool {
		if err := txn.Set(toAccountDataKey(key), value); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.DeletedAccountDataMap.EachAll(func(key string, value bool) bool {
		if err := txn.Delete(toAccountDataKey(key)); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.UTXOMap.EachAll(func(id uint64, utxo *types.UTXO) bool {
		if utxo.TxIn.ID() != id {
			inErr = ErrInvalidTxInKey
			return false
		}
		data, err := encoding.Marshal(utxo.TxOut)
		if err != nil {
			inErr = err
			return false
		}
		if err := txn.Set(toUTXOKey(id), data); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.CreatedUTXOMap.EachAll(func(id uint64, vout *types.TxOut) bool {
		data, err := encoding.Marshal(vout)
		if err != nil {
			inErr = err
			return false
		}
		if err := txn.Set(toUTXOKey(id), data); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.DeletedUTXOMap.EachAll(func(id uint64, utxo *types.UTXO) bool {
		if err := txn.Delete(toUTXOKey(id)); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}

	efc := encoding.Factory("event")
	for _, ev := range ctd.Events {
		t, err := efc.TypeOf(ev)
		if err != nil {
			return err
		}
		var buffer bytes.Buffer
		buffer.Write(util.Uint16ToBytes(t))
		data, err := encoding.Marshal(ev)
		if err != nil {
			return err
		}
		buffer.Write(data)
		if err := txn.Set(toEventKey(types.MarshalID(ev.Height(), ev.Index(), ev.N())), buffer.Bytes()); err != nil {
			return err
		}
	}

	ctd.ProcessDataMap.EachAll(func(key string, value []byte) bool {
		if err := txn.Set(toProcessDataKey(key), value); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctd.DeletedProcessDataMap.EachAll(func(key string, value bool) bool {
		if err := txn.Delete(toProcessDataKey(key)); err != nil {
			inErr = err
			return false
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	return nil
}
