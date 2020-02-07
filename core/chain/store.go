package chain

import (
	"bytes"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/backend"
	"github.com/fletaio/fleta/core/pile"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Store saves the target chain state
// All updates are executed in one transaction with FileSync option
type Store struct {
	sync.Mutex
	db           backend.StoreBackend
	cdb          *pile.DB
	chainID      uint8
	symbol       string
	usage        string
	magicNumber  uint64
	version      uint16
	cache        storecache
	closeLock    sync.RWMutex
	isClose      bool
	timeSlotMap  map[uint32]map[string]bool
	timeSlotLock sync.Mutex
}

type storecache struct {
	cached      bool
	height      uint32
	heightHash  hash.Hash256
	heightBlock *types.Block
}

// NewStore returns a Store
func NewStore(db backend.StoreBackend, cdb *pile.DB, ChainID uint8, symbol string, usage string, version uint16) (*Store, error) {
	st := &Store{
		db:          db,
		cdb:         cdb,
		chainID:     ChainID,
		symbol:      symbol,
		usage:       usage,
		version:     version,
		timeSlotMap: map[uint32]map[string]bool{},
	}
	st.setupMagicNumber()

	go func() {
		for !st.isClose {
			st.closeLock.RLock()
			if st.db != nil {
				st.db.Shrink()
			}
			st.closeLock.RUnlock()
			time.Sleep(5 * time.Minute)
		}
	}()

	return st, nil
}

// Close terminate and clean store
func (st *Store) Close() {
	st.closeLock.Lock()
	defer st.closeLock.Unlock()

	st.isClose = true
	if st.db != nil {
		st.db.Shrink()
		st.db.Close()
	}
	if st.cdb != nil {
		st.cdb.Close()
	}
	st.cdb = nil
	st.db = nil
}

// ChainID returns the chain id of the target chain
func (st *Store) ChainID() uint8 {
	return st.chainID
}

// Name returns the name of the target chain
func (st *Store) Name() string {
	return st.symbol + " " + st.usage
}

// Symbol returns the symbol of the target chain
func (st *Store) Symbol() string {
	return st.symbol
}

// Usage returns the usage of the target chain
func (st *Store) Usage() string {
	return st.usage
}

// Version returns the version of the target chain
func (st *Store) Version() uint16 {
	return st.version
}

// TargetHeight returns the target height of the target chain
func (st *Store) TargetHeight() uint32 {
	return st.Height() + 1
}

// NewLoaderWrapper returns the loader wrapper of the chain
func (st *Store) NewLoaderWrapper(pid uint8) types.LoaderWrapper {
	return types.NewContextWrapper(pid, types.NewContext(st))
}

// NewAddress returns the new address with the magic number of the chain
func (st *Store) NewAddress(height uint32, index uint16) common.Address {
	return common.NewAddress(height, index, st.magicNumber)
}

func (st *Store) setupMagicNumber() {
	if len(st.symbol) < 3 {
		panic("too short symbol")
	} else if len(st.symbol) > 5 {
		panic("too long symbol")
	} else if !isCapitalAndNumber(st.symbol) {
		panic("only capital alphabet and number can be used as symbol")
	}
	if len(st.usage) < 4 {
		panic("too short usage")
	} else if len(st.usage) > 16 {
		panic("too long usage")
	} else if !isAlphabetAndNumber(st.usage) {
		panic("only alphabet and number can be used as symbol")
	}
	base := []byte{243, 133}
	Salt := "PoweredByFLETABlockchain"
	ls := hash.Hash([]byte(st.symbol + "@" + st.usage + "@" + Salt))
	unit := 2
	cnt := len(ls) / unit
	if len(ls)%unit != 0 {
		cnt++
	}
	for i := 0; i < cnt; i++ {
		from := i * unit
		to := (i + 1) * unit
		if to > len(ls) {
			to = len(ls)
		}
		str := ls[from:to]
		for j := 0; j < unit; j++ {
			v := byte(0)
			if j < len(str) {
				v = byte(str[j])
			}
			base[j] = base[j] ^ v
		}
	}
	tbs := make([]byte, 6)
	if base[0] != 0 || base[1] != 0 {
		copy(tbs, []byte(st.symbol))
	}
	for i := 0; i < len(tbs); i += 2 {
		tbs[i] = tbs[i] ^ base[0]
		tbs[i+1] = tbs[i+1] ^ base[1]
	}
	st.magicNumber = binutil.BigEndian.Uint64(append(base, tbs...))
}

// LastStatus returns the recored target height, prev hash and timestamp
func (st *Store) LastStatus() (uint32, hash.Hash256) {
	height := st.Height()
	h, err := st.Hash(height)
	if err != nil {
		panic(err)
	}
	if height == 0 {
		return 0, h
	}
	return height, h
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
	height := st.Height()
	if st.Height() == 0 {
		return 0
	}
	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightBlock.Header.Timestamp
		}
	}
	bh, err := st.Header(height)
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

	h, err := st.cdb.GetHash(height)
	if err != nil {
		if err == pile.ErrInvalidHeight {
			return hash.Hash256{}, backend.ErrNotExistKey
		} else {
			return hash.Hash256{}, err
		}
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
		return nil, backend.ErrNotExistKey
	}
	if st.cache.cached {
		if st.cache.height == height {
			return &st.cache.heightBlock.Header, nil
		}
	}

	value, err := st.cdb.GetData(height, 0)
	if err != nil {
		if err == pile.ErrInvalidHeight {
			return nil, backend.ErrNotExistKey
		} else {
			return nil, err
		}
	}
	var bh types.Header
	if err := encoding.Unmarshal(value, &bh); err != nil {
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
		return nil, backend.ErrNotExistKey
	}
	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightBlock, nil
		}
	}

	value, err := st.cdb.GetDatas(height, 0, 2)
	if err != nil {
		if err == pile.ErrInvalidHeight {
			return nil, backend.ErrNotExistKey
		} else {
			return nil, err
		}
	}
	var b types.Block
	if err := encoding.Unmarshal(value, &b); err != nil {
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
	st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(tagHeight)
		if err != nil {
			return err
		}
		height = binutil.LittleEndian.Uint32(value)
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		if err := txn.Iterate(tagAccount, func(key []byte, value []byte) error {
			if len(value) > 1 {
				acc, err := fc.Create(binutil.LittleEndian.Uint16(value))
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
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toAccountKey(addr))
		if err != nil {
			return err
		}
		if len(value) == 1 && value[0] == 0 {
			return types.ErrDeletedAccount
		}
		v, err := fc.Create(binutil.LittleEndian.Uint16(value))
		if err != nil {
			return err
		}
		if err := encoding.Unmarshal(value[2:], &v); err != nil {
			return err
		}
		acc = v.(types.Account)
		return nil
	}); err != nil {
		if err == backend.ErrNotExistKey {
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toAccountNameKey(Name))
		if err != nil {
			return err
		}
		copy(addr[:], value)
		if true {
			value, err := txn.Get(toAccountKey(addr))
			if err != nil {
				return err
			}
			if len(value) == 1 && value[0] == 0 {
				return types.ErrDeletedAccount
			}
		}
		return nil
	}); err != nil {
		if err == backend.ErrNotExistKey {
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toAccountKey(addr))
		if err != nil {
			return err
		}
		if len(value) == 1 && value[0] == 0 {
			return types.ErrDeletedAccount
		}
		Has = true
		return nil
	}); err != nil {
		if err == backend.ErrNotExistKey {
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

	if _, err := common.ParseAddress(Name); err == nil {
		return false, ErrInvalidAccountName
	}

	var Has bool
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toAccountNameKey(Name))
		if err != nil {
			return err
		}
		var addr common.Address
		copy(addr[:], value)
		if true {
			value, err := txn.Get(toAccountKey(addr))
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
		if err == backend.ErrNotExistKey {
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toAccountDataKey(key))
		if err != nil {
			return err
		}
		data = make([]byte, len(value))
		copy(data, value)
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		if err := txn.Iterate(tagUTXO, func(key []byte, value []byte) error {
			utxo := &types.UTXO{
				TxIn:  types.NewTxIn(fromUTXOKey(key)),
				TxOut: types.NewTxOut(),
			}
			if err := encoding.Unmarshal(value, &(utxo.TxOut)); err != nil {
				return err
			}
			list = append(list, utxo)
			return nil
		}); err != nil {
			return err
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toUTXOKey(id))
		if err != nil {
			return err
		}
		Has = (len(value) > 0)
		return nil
	}); err != nil {
		if err == backend.ErrNotExistKey {
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toUTXOKey(id))
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
		if err == backend.ErrNotExistKey {
			return nil, types.ErrNotExistUTXO
		} else {
			return nil, err
		}
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
	if err := st.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(toProcessDataKey(key))
		if err != nil {
			return err
		}
		data = make([]byte, len(value))
		copy(data, value)
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
	for i := From; i <= To; i++ {
		value, err := st.cdb.GetData(i, 2)
		if err != nil {
			if err == pile.ErrInvalidHeight || err == pile.ErrInvalidDataIndex {
				continue
			} else {
				return nil, err
			}
		}
		dec := encoding.NewDecoder(bytes.NewReader(value))
		EvLen, err := dec.DecodeArrayLen()
		if err != nil {
			return nil, err
		}
		for j := 0; j < EvLen; j++ {
			t, err := dec.DecodeUint16()
			if err != nil {
				return nil, err
			}
			ev, err := fc.Create(t)
			if err != nil {
				return nil, err
			}
			if err := dec.Decode(&ev); err != nil {
				return nil, err
			}
			list = append(list, ev.(types.Event))
		}
	}
	return list, nil
}

// IsUsedTimeSlot returns timeslot is used or not
func (st *Store) IsUsedTimeSlot(slot uint32, key string) bool {
	st.timeSlotLock.Lock()
	defer st.timeSlotLock.Unlock()

	tm, has := st.timeSlotMap[slot]
	if !has {
		return false
	}
	return tm[key]
}

// StoreGenesis stores the genesis data
func (st *Store) StoreGenesis(genHash hash.Hash256, ctd *types.ContextData) error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return ErrStoreClosed
	}

	st.Lock()
	defer st.Unlock()

	if st.Height() > 0 {
		return ErrAlreadyGenesised
	}
	if err := st.cdb.Init(genHash); err != nil {
		if err != pile.ErrAlreadyInitialized {
			return err
		}
	}
	if err := st.db.Update(func(txn backend.StoreWriter) error {
		{
			if err := txn.Set(toHeightHashKey(0), genHash[:]); err != nil {
				return err
			}
			bsHeight := binutil.LittleEndian.Uint32ToBytes(0)
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

	st.Lock()
	defer st.Unlock()

	DataHash := encoding.Hash(b.Header)
	Datas := [][]byte{}
	{
		data, err := encoding.Marshal(b.Header)
		if err != nil {
			return err
		}
		Datas = append(Datas, data)
	}
	{
		data, err := encoding.Marshal(b)
		if err != nil {
			return err
		}
		Datas = append(Datas, data[len(Datas[0]):]) // cut header data
	}
	if len(ctd.Events) > 0 {
		var buffer bytes.Buffer
		efc := encoding.Factory("event")
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
		Datas = append(Datas, buffer.Bytes())
	}
	if err := st.cdb.AppendData(b.Header.Height, DataHash, Datas); err != nil {
		if err != pile.ErrInvalidAppendHeight {
			return err
		}
	}
	if err := st.db.Update(func(txn backend.StoreWriter) error {
		{
			bsHeight := binutil.LittleEndian.Uint32ToBytes(b.Header.Height)
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

	st.timeSlotLock.Lock()
	ctd.TimeSlotMap.EachAll(func(key uint32, mp *types.StringBoolMap) bool {
		smp, has := st.timeSlotMap[key]
		if !has {
			smp = map[string]bool{}
			st.timeSlotMap[key] = smp
		}
		mp.EachAll(func(key string, value bool) bool {
			smp[key] = true
			return true
		})
		return true
	})
	currentSlot := types.ToTimeSlot(b.Header.Timestamp)
	deleteSlots := []uint32{}
	for slot := range st.timeSlotMap {
		if slot < currentSlot-1 {
			deleteSlots = append(deleteSlots, slot)
		}
	}
	for _, v := range deleteSlots {
		delete(st.timeSlotMap, v)
	}
	st.timeSlotLock.Unlock()

	st.cache.height = b.Header.Height
	st.cache.heightHash = DataHash
	st.cache.heightBlock = b
	st.cache.cached = true
	return nil
}

func (st *Store) IterBlockAfterContext(fn func(b *types.Block) error) error {
	for h := st.Height() + 1; ; h++ {
		b, err := st.Block(h)
		if err != nil {
			if err == backend.ErrNotExistKey {
				break
			} else {
				return err
			}
		}
		if err := fn(b); err != nil {
			return err
		}
	}
	return nil
}

func applyContextData(txn backend.StoreWriter, ctd *types.ContextData) error {
	var inErr error
	afc := encoding.Factory("account")
	ctd.AccountMap.EachAll(func(addr common.Address, acc types.Account) bool {
		t, err := afc.TypeOf(acc)
		if err != nil {
			inErr = err
			return false
		}
		var buffer bytes.Buffer
		buffer.Write(binutil.LittleEndian.Uint16ToBytes(t))
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
	ctd.DeletedAccountMap.EachAll(func(addr common.Address, acc types.Account) bool {
		if err := txn.Set(toAccountKey(addr), []byte{0}); err != nil {
			inErr = err
			return false
		}
		prefix := toAccountDataKey(string(addr[:]))
		Deletes := [][]byte{}
		if err := txn.Iterate(prefix, func(key []byte, value []byte) error {
			Deletes = append(Deletes, key)
			return nil
		}); err != nil {
			inErr = err
			return false
		}
		for _, v := range Deletes {
			if err := txn.Delete(v); err != nil {
				inErr = err
				return false
			}
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

func (st *Store) InitTimeSlot() error {
	Height := st.Height()
	if Height > 0 {
		st.timeSlotLock.Lock()
		bh, err := st.Header(Height)
		if err != nil {
			return err
		}
		lastSlot := types.ToTimeSlot(bh.Timestamp)
		for h := Height; h >= 1; h-- {
			b, err := st.Block(h)
			if err != nil {
				return err
			}
			currentSlot := types.ToTimeSlot(b.Header.Timestamp)
			if currentSlot < lastSlot-1 {
				break
			}
			for i, tx := range b.Transactions {
				slot := types.ToTimeSlot(tx.Timestamp())
				if slot >= currentSlot-1 {
					mp, has := st.timeSlotMap[slot]
					if !has {
						mp = map[string]bool{}
						st.timeSlotMap[slot] = mp
					}
					t := b.TransactionTypes[i]
					TxHash := HashTransactionByType(st.chainID, t, tx)
					mp[string(TxHash[:])] = true
				}
			}
		}
		st.timeSlotLock.Unlock()
	}
	return nil
}
