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
	chainID    uint8
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
func NewStore(path string, ChainID uint8, name string, version uint16, bRecover bool) (*Store, error) {
	opts := badger.DefaultOptions(path)
	opts.Truncate = bRecover
	opts.SyncWrites = true
	lockfilePath := filepath.Join(opts.Dir, "LOCK")
	os.MkdirAll(path, os.ModePerm)

	os.Remove(lockfilePath)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			MaxCount := 10
			Count := 0
		again:
			if err := db.RunValueLogGC(0.5); err != nil {
			} else {
				Count++
				if Count < MaxCount {
					goto again
				}
			}
		}
	}()

	return &Store{
		db:      db,
		ticker:  ticker,
		chainID: ChainID,
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
	MaxCount := 10
	Count := 0
again:
	if err := st.db.RunValueLogGC(1); err != nil {
	} else {
		Count++
		if Count < MaxCount {
			goto again
		}
	}
	st.db.Close()
	st.ticker.Stop()
	st.db = nil
	st.ticker = nil
}

// ChainID returns the chain id of the target chain
func (st *Store) ChainID() uint8 {
	return st.chainID
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

// NewContextWrapper returns the context wrapper of the chain
func (st *Store) NewContextWrapper(pid uint8) *types.ContextWrapper {
	return types.NewContextWrapper(pid, types.NewContext(st))
}

// LastStatus returns the recored target height, prev hash and timestamp
func (st *Store) LastStatus() (uint32, hash.Hash256, uint64) {
	height := st.Height()
	h, err := st.Hash(height)
	if err != nil {
		panic(err)
	}
	if height == 0 {
		return 0, h, 0
	}
	bh, err := st.Header(height)
	if err != nil {
		if err != ErrStoreClosed {
			// should have not reabh
			panic(err)
		}
		return 0, hash.Hash256{}, 0
	}
	return bh.Height, h, bh.Timestamp
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
		item, err := txn.Get(tagHeight)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotExistKey
			} else {
				return err
			}
		}
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
			if !item.IsDeletedOrExpired() {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				if len(value) > 1 {
					acc, err := fc.Create(util.BytesToUint16(value))
					if err != nil {
						return err
					}
					if err := encoding.Unmarshal(value[2:], &acc); err != nil {
						return err
					}
					list = append(list, acc.(types.Account))
				}
			}
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
			if item.IsDeletedOrExpired() {
				return ErrNotExistKey
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if len(value) == 1 && value[0] == 0 {
			return types.ErrDeletedAccount
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		copy(addr[:], value)
		if true {
			item, err := txn.Get(toAccountKey(addr))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return ErrNotExistKey
				} else {
					return err
				}
			}
			if item.IsDeletedOrExpired() {
				return ErrNotExistKey
			}
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(value) == 1 && value[0] == 0 {
				return types.ErrDeletedAccount
			}
		}
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if len(value) == 1 && value[0] == 0 {
			return types.ErrDeletedAccount
		}
		Has = true
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
		}
		var addr common.Address
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		copy(addr[:], value)
		if true {
			item, err := txn.Get(toAccountKey(addr))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return ErrNotExistKey
				} else {
					return err
				}
			}
			if item.IsDeletedOrExpired() {
				return ErrNotExistKey
			}
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(value) == 1 && value[0] == 0 {
				return types.ErrDeletedAccount
			}
		}
		Has = true
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

// AccountData returns the account data from the store
func (st *Store) AccountData(addr common.Address, pid uint8, name []byte) []byte {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil
	}

	key := string(addr[:]) + string(pid) + string(name)
	var data []byte
	if err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(toAccountDataKey(key))
		if err != nil {
			return err
		}
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
			if !item.IsDeletedOrExpired() {
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
		}
		Has = true
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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
		if item.IsDeletedOrExpired() {
			return ErrNotExistKey
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

	Height := st.Height()
	if To > Height {
		To = Height
	}

	fc := encoding.Factory("event")
	list := []types.Event{}
	if err := st.db.View(func(txn *badger.Txn) error {
		for i := From; i <= To; i++ {
			item, err := txn.Get(toEventKey(i))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					continue
				} else {
					return err
				}
			}
			if !item.IsDeletedOrExpired() {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				dec := encoding.NewDecoder(bytes.NewReader(value))
				EvLen, err := dec.DecodeArrayLen()
				if err != nil {
					return err
				}
				for i := 0; i < EvLen; i++ {
					t, err := dec.DecodeUint16()
					if err != nil {
						return err
					}
					ev, err := fc.Create(t)
					if err != nil {
						return err
					}
					if err := dec.Decode(&ev); err != nil {
						return err
					}
					list = append(list, ev.(types.Event))
				}
			}
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
			if err := txn.Set(tagHeight, bsHeight); err != nil {
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
			if err := txn.Set(tagHeight, bsHeight); err != nil {
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
		if err := txn.Set(toAccountKey(addr), []byte{0}); err != nil {
			inErr = err
			return false
		}
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
		})
		defer it.Close()
		prefix := toAccountDataKey(string(addr[:]))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if !item.IsDeletedOrExpired() {
				if err := txn.Delete(item.Key()); err != nil {
					inErr = err
					return false
				}
			}
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

	if len(ctd.Events) > 0 {
		efc := encoding.Factory("event")

		var buffer bytes.Buffer
		enc := encoding.NewEncoder(&buffer)
		if err := enc.EncodeArrayLen(len(ctd.Events)); err != nil {
			return err
		}
		for _, ev := range ctd.Events {
			t, err := efc.TypeOf(ev)
			if err != nil {
				return err
			}
			if err := enc.EncodeUint16(t); err != nil {
				return err
			}
			if err := enc.Encode(ev); err != nil {
				return err
			}
		}
		if err := txn.Set(toEventKey(ctd.Events[0].Height()), buffer.Bytes()); err != nil {
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
