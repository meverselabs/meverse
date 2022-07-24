package chain

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/keydb"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// Store saves the target chain state
// All updates are executed in one transaction with FileSync option
type Store struct {
	sync.Mutex
	db             *keydb.DB
	cdb            *piledb.DB
	chainID        *big.Int
	version        uint16
	feeUnit        *amount.Amount
	cache          storecache
	closeLock      sync.RWMutex
	isClose        bool
	AddrSeqMapLock sync.Mutex
	AddrSeqMap     map[common.Address]uint64
	timeSlotMap    map[uint32]map[string]bool
	timeSlotLock   sync.Mutex
	rankTable      *RankTable
	keydbPath      string
}

type storecache struct {
	cached           bool
	height           uint32
	heightHash       hash.Hash256
	heightBlock      *types.Block
	heightTimestamp  uint64
	heightPoFSameGen uint32
	admins           []common.Address
	generators       []common.Address
	contracts        []types.Contract
}

// NewStore returns a Store
func NewStore(keydbPath string, cdb *piledb.DB, ChainID *big.Int, Version uint16) (*Store, error) {
	db, err := keydb.Open(keydbPath, func(key []byte, value []byte) (interface{}, error) {
		switch key[0] {
		case tagHeight:
			return bin.Uint32(value), nil
		case tagHeightHash:
			var h hash.Hash256
			h.SetBytes(value)
			return h, nil
		case tagPoFRankTable:
			rt := &RankTable{}
			if _, err := rt.ReadFrom(bytes.NewReader(value)); err != nil {
				return nil, err
			}
			return rt, nil
		case tagAdmin:
			return value[0] == 1, nil
		case tagAddressSeq:
			seq := binary.LittleEndian.Uint64(value)
			return seq, nil
		case tagGenerator:
			return value[0] == 1, nil
		case tagContract:
			cd := &types.ContractDefine{}
			if _, err := cd.ReadFrom(bytes.NewReader(value)); err != nil {
				return nil, err
			}
			return cd, nil
		case tagData:
			data := make([]byte, len(value))
			copy(data, value)
			return data, nil
		case tagBlockGen:
			return bin.Uint32(value), nil
		case tagMainToken:
			var addr common.Address
			copy(addr[:], value)
			return addr, nil
		default:
			panic("unknown data type")
		}
	})
	if err != nil {
		return nil, err
	}

	st := &Store{
		db:          db,
		cdb:         cdb,
		chainID:     ChainID,
		version:     Version,
		AddrSeqMap:  map[common.Address]uint64{},
		timeSlotMap: map[uint32]map[string]bool{},
		keydbPath:   keydbPath,
	}

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
func (st *Store) ChainID() *big.Int {
	return st.chainID
}

// Version returns the version of the target chain
func (st *Store) Version() uint16 {
	return st.version
}

// TargetHeight returns the target height of the target chain
func (st *Store) TargetHeight() uint32 {
	return st.Height() + 1
}

// LastStatus returns the recored target height, prev hash and timestamp
func (st *Store) LastStatus() (uint32, hash.Hash256) {
	height := st.Height()
	h, err := st.Hash(height)
	if err != nil {
		panic(err)
	}
	return height, h
}

// LastHash returns the last hash of the chain
func (st *Store) LastHash() hash.Hash256 {
	return st.PrevHash()
}

// PrevHash returns the prev hash of the chain
func (st *Store) PrevHash() hash.Hash256 {
	h, err := st.Hash(st.Height())
	if err != nil {
		if errors.Cause(err) != ErrStoreClosed {
			// should have not reabh
			panic(err)
		}
		return hash.Hash256{}
	}
	return h
}

// LastTimestamp returns the prev timestamp of the chain
func (st *Store) LastTimestamp() uint64 {
	height := st.Height()
	if st.Height() == 0 {
		return 0
	}
	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightTimestamp
		}
	}
	bh, err := st.Header(height)
	if err != nil {
		if errors.Cause(err) != ErrStoreClosed {
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
		return hash.Hash256{}, errors.WithStack(ErrStoreClosed)
	}

	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightHash, nil
		}
	}

	h, err := st.cdb.GetHash(height)
	if err != nil {
		if errors.Cause(err) == piledb.ErrInvalidHeight {
			return hash.Hash256{}, errors.WithStack(keydb.ErrNotFound)
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
		return nil, errors.WithStack(ErrStoreClosed)
	}

	if height <= st.InitHeight() {
		return nil, errors.WithStack(keydb.ErrNotFound)
	}
	if st.cache.cached {
		if st.cache.height == height {
			return &st.cache.heightBlock.Header, nil
		}
	}

	value, err := st.cdb.GetData(height, 0)
	if err != nil {
		if errors.Cause(err) == piledb.ErrInvalidHeight {
			return nil, errors.WithStack(keydb.ErrNotFound)
		} else {
			return nil, err
		}
	}
	var bh types.Header
	if _, err := bin.ReadFromBytes(&bh, value); err != nil {
		return nil, err
	}
	return &bh, nil
}

// Block returns the block by height
func (st *Store) Block(height uint32) (*types.Block, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	if height <= st.InitHeight() {
		return nil, errors.WithStack(keydb.ErrNotFound)
	}
	if st.cache.cached {
		if st.cache.height == height {
			return st.cache.heightBlock, nil
		}
	}

	value, err := st.cdb.GetDatas(height, 0, 2)
	if err != nil {
		if errors.Cause(err) == piledb.ErrInvalidHeight {
			return nil, errors.WithStack(keydb.ErrNotFound)
		} else {
			return nil, err
		}
	}
	var b types.Block
	if _, err := bin.ReadFromBytes(&b, value); err != nil {
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
	st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get([]byte{tagHeight})
		if err != nil {
			return err
		}
		height = value.(uint32)
		return nil
	})
	return height
}

// InitHeight returns the initial height of the target chain
func (st *Store) InitHeight() uint32 {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return 0
	}

	return st.cdb.InitHeight()
}

// InitTimestamp returns the initial timestamp of the target chain
func (st *Store) InitTimestamp() uint64 {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return 0
	}

	return st.cdb.InitTimestamp()
}

// TopGenerator returns current top generator
func (st *Store) TopGenerator(TimeoutCount uint32) (common.Address, error) {
	top, err := st.rankTable.TopRank(TimeoutCount)
	if err != nil {
		return common.Address{}, nil
	}
	return top.Address, nil
}

// RanksInMap returns current top generator
func (st *Store) GeneratorsInMap(GeneratorMap map[common.Address]bool, Limit int) ([]common.Address, error) {
	return st.rankTable.GeneratorsInMap(GeneratorMap, Limit)
}

// TopRankInMap returns current top generator
func (st *Store) TopGeneratorInMap(GeneratorMap map[common.Address]bool) (common.Address, uint32, error) {
	return st.rankTable.TopGeneratorInMap(GeneratorMap)
}

// Admins returns all Admins of the target chain
func (st *Store) Admins() ([]common.Address, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	if st.cache.cached {
		return st.cache.admins, nil
	}
	admins := []common.Address{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		return txn.Iterate([]byte{tagAdmin}, func(key []byte, value interface{}) error {
			if value.(bool) {
				var addr common.Address
				copy(addr[:], key[1:])
				admins = append(admins, addr)
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return admins, nil
}

// IsAdmin returns the account is Admin or not
func (st *Store) IsAdmin(addr common.Address) bool {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false
	}

	var is bool
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toAdminKey(addr))
		if err != nil {
			return errors.WithStack(err)
		}
		is = value.(bool)
		return nil
	}); err != nil {
		return false
	}
	return is
}

// Seq returns the sequence of the transaction
func (st *Store) AddrSeq(addr common.Address) uint64 {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return 0
	}

	st.AddrSeqMapLock.Lock()
	defer st.AddrSeqMapLock.Unlock()

	if seq, has := st.AddrSeqMap[addr]; has {
		return seq
	} else {
		var seq uint64
		if err := st.db.View(func(txn *keydb.Tx) error {
			value, err := txn.Get(toAddressSeqKey(addr))
			if err != nil {
				return err
			}
			var ok bool
			seq, ok = value.(uint64)
			if !ok {
				return errors.WithStack(ErrInvalidAdminAddress)
			}
			return nil
		}); err != nil {
			return 0
		}
		st.AddrSeqMap[addr] = seq
		return seq
	}
}

// Generators returns all generators of the target chain
func (st *Store) Generators() ([]common.Address, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	if st.cache.cached {
		return st.cache.generators, nil
	}
	generators := []common.Address{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		return txn.Iterate([]byte{tagGenerator}, func(key []byte, value interface{}) error {
			if value.(bool) {
				var addr common.Address
				copy(addr[:], key[1:])
				generators = append(generators, addr)
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return generators, nil
}

// IsGenerator returns the account is generator or not
func (st *Store) IsGenerator(addr common.Address) bool {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false
	}

	var is bool
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toGeneratorKey(addr))
		if err != nil {
			return errors.WithStack(err)
		}
		is = value.(bool)
		return nil
	}); err != nil {
		return false
	}
	return is
}

// MainToken returns the account MainToken
func (st *Store) MainToken() *common.Address {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil
	}

	var addr common.Address
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get([]byte{tagMainToken})
		if err != nil {
			return errors.WithStack(err)
		}
		addr = value.(common.Address)
		return nil
	}); err != nil {
		return nil
	}
	return &addr
}

// BasicFee returns basic fee
func (st *Store) BasicFee() *amount.Amount {
	if st.feeUnit == nil {
		st.closeLock.RLock()
		defer st.closeLock.RUnlock()
		if err := st.db.View(func(txn *keydb.Tx) error {
			value, err := txn.Get([]byte{tagBasicFee})
			if err != nil {
				return errors.WithStack(err)
			}
			bs := value.([]byte)
			if len(bs) == 0 {
				st.feeUnit = amount.NewAmount(0, 0) // 0
			} else {
				st.feeUnit = amount.NewAmountFromBytes(bs)
			}
			return nil
		}); err != nil {
			st.feeUnit = amount.NewAmount(0, 100000000000000000) // 0.1
			return st.feeUnit
		}
	}
	return st.feeUnit
}

// Contracts returns the contract form the store
func (st *Store) Contracts() ([]types.Contract, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	if st.cache.cached {
		return st.cache.contracts, nil
	}
	conts := []types.Contract{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		txn.Iterate([]byte{tagContract}, func(key []byte, value interface{}) error {
			cd := value.(*types.ContractDefine)
			cont, err := types.CreateContract(cd)
			if err != nil {
				return err
			}
			conts = append(conts, cont)
			return nil
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return conts, nil
}

// BlockGenMap returns current block gen map
func (st *Store) BlockGenMap() (map[common.Address]uint32, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	mp := map[common.Address]uint32{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		txn.Iterate([]byte{tagBlockGen}, func(key []byte, value interface{}) error {
			mp[fromBlockGenKey(key)] = value.(uint32)
			return nil
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return mp, nil
}

// IsContract returns is the contract
func (st *Store) IsContract(addr common.Address) bool {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return false
	}

	var exist bool
	if err := st.db.View(func(txn *keydb.Tx) error {
		_, err := txn.Get(toContractKey(addr))
		if err != nil {
			return errors.WithStack(err)
		}
		exist = true
		return nil
	}); err != nil {
		return false
	}
	return exist
}

// Contract returns the contract form the store
func (st *Store) Contract(addr common.Address) (types.Contract, error) {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil, errors.WithStack(ErrStoreClosed)
	}

	var cd *types.ContractDefine
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toContractKey(addr))
		if err != nil {
			return errors.WithStack(err)
		}
		cd = value.(*types.ContractDefine)
		return nil
	}); err != nil {
		return nil, err
	}
	cont, err := types.CreateContract(cd)
	if err != nil {
		return nil, err
	}
	return cont, nil
}

// Data returns the account data from the store
func (st *Store) Data(cont common.Address, addr common.Address, name []byte) []byte {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return nil
	}

	key := string(cont[:]) + string(addr[:]) + string(name)
	var data []byte
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toDataKey(key))
		if err != nil {
			return errors.WithStack(err)
		}
		bs := value.([]byte)
		data = make([]byte, len(bs))
		copy(data, bs)
		return nil
	}); err != nil {
		return nil
	}
	return data
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
		return errors.WithStack(ErrStoreClosed)
	}

	st.Lock()
	defer st.Unlock()

	if st.Height() > 0 {
		return errors.WithStack(ErrAlreadyGenesised)
	}

	rt := NewRankTable()
	phase := rt.smallestPhase() + 2
	for addr := range ctd.GeneratorMap {
		if err := rt.addRank(NewRank(addr, phase, hash.DoubleHash(addr[:]))); err != nil {
			return err
		}
	}

	if err := st.cdb.Init(genHash, genHash, 0, 0); err != nil {
		if errors.Cause(err) != piledb.ErrAlreadyInitialized {
			return err
		}
	}
	if err := st.db.Update(func(txn *keydb.Tx) error {
		if err := txn.Set(toHeightHashKey(0), genHash, genHash[:]); err != nil {
			return errors.WithStack(err)
		}
		Height := uint32(0)
		if err := txn.Set([]byte{tagHeight}, Height, bin.Uint32Bytes(Height)); err != nil {
			return errors.WithStack(err)
		}
		if err := applyContextData(txn, ctd); err != nil {
			return err
		}
		{
			bsRankTable, _, err := bin.WriterToBytes(rt)
			if err != nil {
				return err
			}
			if err := txn.Set([]byte{tagPoFRankTable}, rt, bsRankTable); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	st.rankTable = rt

	st.cache.height = 0
	st.cache.heightHash = genHash
	st.cache.heightBlock = nil
	st.cache.heightTimestamp = 0
	st.cache.heightPoFSameGen = 0
	st.cache.generators = []common.Address{}
	st.cache.contracts = []types.Contract{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		if err := txn.Iterate([]byte{tagGenerator}, func(key []byte, value interface{}) error {
			if value.(bool) {
				var addr common.Address
				copy(addr[:], key[1:])
				st.cache.generators = append(st.cache.generators, addr)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := txn.Iterate([]byte{tagContract}, func(key []byte, value interface{}) error {
			cd := value.(*types.ContractDefine)
			cont, err := types.CreateContract(cd)
			if err != nil {
				return err
			}
			st.cache.contracts = append(st.cache.contracts, cont)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	st.cache.cached = true
	return nil
}

// StoreInit stores the init data
func (st *Store) StoreInit(genHash hash.Hash256, initHash hash.Hash256, initHeight uint32, initTimestamp uint64) error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return errors.WithStack(ErrStoreClosed)
	}

	if st.Height() > initHeight {
		return errors.WithStack(ErrAlreadyInitialzed)
	}
	if err := st.cdb.Init(genHash, initHash, initHeight, initTimestamp); err != nil {
		if errors.Cause(err) != piledb.ErrAlreadyInitialized {
			return err
		}
	}
	if err := st.db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toHeightHashKey(0))
		if err != nil {
			return errors.WithStack(err)
		}
		bsHash, ok := value.([]byte)
		if !ok {
			if Hash, ok := value.(hash.Hash256); ok {
				bsHash = Hash[:]
			}
		}
		if !bytes.Equal(bsHash, genHash[:]) {
			return errors.WithStack(piledb.ErrInvalidGenesisHash)
		}
		{
			v, err := txn.Get([]byte{tagPoFRankTable})
			if err != nil {
				return errors.WithStack(err)
			}
			st.rankTable = v.(*RankTable)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := st.db.Update(func(txn *keydb.Tx) error {
		if err := txn.Set(toHeightHashKey(initHeight), initHash, initHash[:]); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}); err != nil {
		return err
	}

	st.cache.height = initHeight
	st.cache.heightHash = initHash
	st.cache.heightBlock = nil
	st.cache.heightTimestamp = initTimestamp
	st.cache.generators = []common.Address{}
	st.cache.contracts = []types.Contract{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		if err := txn.Iterate([]byte{tagGenerator}, func(key []byte, value interface{}) error {
			if value.(bool) {
				var addr common.Address
				copy(addr[:], key[1:])
				st.cache.generators = append(st.cache.generators, addr)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := txn.Iterate([]byte{tagContract}, func(key []byte, value interface{}) error {
			cd := value.(*types.ContractDefine)
			cont, err := types.CreateContract(cd)
			if err != nil {
				return err
			}
			st.cache.contracts = append(st.cache.contracts, cont)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	st.cache.cached = true
	return nil
}

// Prepare loads the initial data
func (st *Store) Prepare() error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return errors.WithStack(ErrStoreClosed)
	}

	if st.rankTable == nil {
		if err := st.db.View(func(txn *keydb.Tx) error {
			{
				v, err := txn.Get([]byte{tagPoFRankTable})
				if err != nil {
					return errors.WithStack(err)
				}
				st.rankTable = v.(*RankTable)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if !st.cache.cached {
		st.cache.height, st.cache.heightHash = st.LastStatus()
		b, err := st.Block(st.cache.height)
		if err != nil {
			return err
		}
		st.cache.heightBlock = b
		st.cache.heightTimestamp = st.LastTimestamp()
		st.cache.generators = []common.Address{}
		st.cache.contracts = []types.Contract{}
		if err := st.db.View(func(txn *keydb.Tx) error {
			if err := txn.Iterate([]byte{tagGenerator}, func(key []byte, value interface{}) error {
				if value.(bool) {
					var addr common.Address
					copy(addr[:], key[1:])
					st.cache.generators = append(st.cache.generators, addr)
				}
				return nil
			}); err != nil {
				return err
			}
			if err := txn.Iterate([]byte{tagContract}, func(key []byte, value interface{}) error {
				cd := value.(*types.ContractDefine)
				cont, err := types.CreateContract(cd)
				if err != nil {
					return err
				}
				st.cache.contracts = append(st.cache.contracts, cont)
				return nil
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		st.cache.cached = true
	}
	return nil
}

func (st *Store) ProcessReward(ctx *types.Context, b *types.Block) (map[common.Address][]byte, error) {
	conts, err := st.Contracts()
	if err != nil {
		return nil, err
	}
	GenMap, err := st.BlockGenMap()
	if err != nil {
		return nil, err
	}
	var rewardMap map[common.Address][]byte
	if len(conts) > 0 {
		rewardMap = map[common.Address][]byte{}
	}
	for _, cont := range conts {
		cc := ctx.ContractContext(cont, common.Address{})
		intr := types.NewInteractor(ctx, cont, cc, "000000000000", false)
		cc.Exec = intr.Exec
		if rewardEvent, err := cont.OnReward(cc, b, GenMap); err != nil {
			return nil, err
		} else if rewardEvent != nil {
			if bs, err := types.MarshalAddressAmountMap(rewardEvent); err != nil {
				return nil, err
			} else {
				rewardMap[cont.Address()] = bs
			}
		}
		intr.Distroy()
	}
	return rewardMap, nil
}

// StoreBlock stores the block
func (st *Store) StoreBlock(b *types.Block, ctx *types.Context) error {
	st.closeLock.RLock()
	defer st.closeLock.RUnlock()
	if st.isClose {
		return errors.WithStack(ErrStoreClosed)
	}

	st.Lock()
	defer st.Unlock()

	bsHeader, _, err := bin.WriterToBytes(&b.Header)
	if err != nil {
		return err
	}
	HeaderHash := hash.Hash(bsHeader)
	Datas := [][]byte{bsHeader}
	{
		data, _, err := bin.WriterToBytes(&b.Body)
		if err != nil {
			return err
		}
		Datas = append(Datas, data)
	}
	if err := st.cdb.AppendData(b.Header.Height, HeaderHash, Datas); err != nil {
		if errors.Cause(err) != piledb.ErrInvalidAppendHeight {
			return err
		}
	}

	ctd := ctx.Top()
	if err := st.db.Update(func(txn *keydb.Tx) error {
		if err := txn.Set([]byte{tagHeight}, b.Header.Height, bin.Uint32Bytes(b.Header.Height)); err != nil {
			return err
		}

		if ctx.IsProcessReward() {
			keys := [][]byte{}
			if err := txn.Iterate([]byte{tagBlockGen}, func(key []byte, value interface{}) error {
				keys = append(keys, key)
				return nil
			}); err != nil {
				return err
			}
			for _, v := range keys {
				if err := txn.Delete(v); err != nil {
					return err
				}
			}
		}

		{
			var cnt uint32
			v, err := txn.Get(toBlockGenKey(b.Header.Generator))
			if err != nil {
				if errors.Cause(err) != keydb.ErrNotFound {
					return err
				}
			} else {
				cnt = v.(uint32)
			}
			cnt++
			if err := txn.Set(toBlockGenKey(b.Header.Generator), cnt, bin.Uint32Bytes(cnt)); err != nil {
				return err
			}
		}

		if err := applyContextData(txn, ctd); err != nil {
			return err
		}
		var rt *RankTable
		{
			v, err := txn.Get([]byte{tagPoFRankTable})
			if err != nil {
				return errors.WithStack(err)
			}
			rt = v.(*RankTable)
		}
		if b.Header.TimeoutCount > 0 {
			if err := rt.forwardCandidates(int(b.Header.TimeoutCount)); err != nil {
				return err
			}
		}

		phase := rt.smallestPhase() + 2
		for addr := range ctd.GeneratorMap {
			if err := rt.addRank(NewRank(addr, phase, hash.DoubleHash(addr[:]))); err != nil {
				return err
			}
		}
		for addr := range ctd.DeletedGeneratorMap {
			rt.removeRank(addr)
		}
		if rt.CandidateCount() == 0 {
			return errors.WithStack(ErrInsufficientCandidateCount)
		}

		bsRankTable, _, err := bin.WriterToBytes(rt)
		if err != nil {
			return err
		}
		if err := txn.Set([]byte{tagPoFRankTable}, rt, bsRankTable); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}); err != nil {
		return err
	}

	st.AddrSeqMapLock.Lock()
	types.EachAllAddressUint64(ctd.AddrSeqMap, func(key common.Address, value uint64) error {
		st.AddrSeqMap[key] = value
		return nil
	})
	st.AddrSeqMapLock.Unlock()

	st.timeSlotLock.Lock()
	types.EachAllTimeSlotMap(ctd.TimeSlotMap, func(key uint32, value map[string]bool) error {
		smp, has := st.timeSlotMap[key]
		if !has {
			smp = map[string]bool{}
			st.timeSlotMap[key] = smp
		}
		types.EachAllStringBool(value, func(key string, value bool) error {
			smp[key] = true
			return nil
		})
		return nil
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
	st.cache.heightHash = HeaderHash
	st.cache.heightBlock = b
	st.cache.heightTimestamp = b.Header.Timestamp
	st.cache.generators = []common.Address{}
	st.cache.contracts = []types.Contract{}
	if err := st.db.View(func(txn *keydb.Tx) error {
		if err := txn.Iterate([]byte{tagGenerator}, func(key []byte, value interface{}) error {
			if value.(bool) {
				var addr common.Address
				copy(addr[:], key[1:])
				st.cache.generators = append(st.cache.generators, addr)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := txn.Iterate([]byte{tagContract}, func(key []byte, value interface{}) error {
			cd := value.(*types.ContractDefine)
			cont, err := types.CreateContract(cd)
			if err != nil {
				return err
			}
			st.cache.contracts = append(st.cache.contracts, cont)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	st.cache.cached = true
	return nil
}

func (st *Store) IterBlockAfterContext(fn func(b *types.Block) error) error {
	for h := st.Height() + 1; ; h++ {
		b, err := st.Block(h)
		if err != nil {
			if errors.Cause(err) == keydb.ErrNotFound {
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

func applyContextData(txn *keydb.Tx, ctd *types.ContextData) error {
	if err := types.EachAllAddressUint64(ctd.AddrSeqMap, func(key common.Address, value uint64) error {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, value)
		return txn.Set(toAddressSeqKey(key), value, bs)
	}); err != nil {
		return err
	}
	addr := ctd.UnsafeGetMainToken()
	if addr != nil {
		txn.Set([]byte{tagMainToken}, *addr, (*addr)[:])
	}
	if err := types.EachAllAddressBool(ctd.AdminMap, func(key common.Address, value bool) error {
		return txn.Set(toAdminKey(key), true, []byte{1})
	}); err != nil {
		return err
	}
	if err := types.EachAllAddressBool(ctd.DeletedAdminMap, func(key common.Address, value bool) error {
		return txn.Delete(toAdminKey(key))
	}); err != nil {
		return err
	}
	if err := types.EachAllAddressBool(ctd.GeneratorMap, func(key common.Address, value bool) error {
		return txn.Set(toGeneratorKey(key), true, []byte{1})
	}); err != nil {
		return err
	}
	if err := types.EachAllAddressBool(ctd.DeletedGeneratorMap, func(key common.Address, value bool) error {
		return txn.Delete(toGeneratorKey(key))
	}); err != nil {
		return err
	}
	if err := types.EachAllAddressContractDefine(ctd.ContractDefineMap, func(key common.Address, cd *types.ContractDefine) error {
		bs, _, err := bin.WriterToBytes(cd)
		if err != nil {
			return err
		}
		if err := txn.Set(toContractKey(key), cd.Clone(), bs); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if err := types.EachAllStringBytes(ctd.DataMap, func(key string, value []byte) error {
		return txn.Set(toDataKey(key), value, value)
	}); err != nil {
		return err
	}
	if err := types.EachAllStringBool(ctd.DeletedDataMap, func(key string, value bool) error {
		txn.Delete(toDataKey(key))
		return nil
	}); err != nil {
		return err
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
			for _, tx := range b.Body.Transactions {
				slot := types.ToTimeSlot(tx.Timestamp)
				if slot >= currentSlot-1 {
					mp, has := st.timeSlotMap[slot]
					if !has {
						mp = map[string]bool{}
						st.timeSlotMap[slot] = mp
					}
					TxHash := tx.Hash(bh.Height)
					mp[string(TxHash[:])] = true
				}
			}
		}
		st.timeSlotLock.Unlock()
	}
	return nil
}
