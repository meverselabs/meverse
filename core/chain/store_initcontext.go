package chain

import (
	"io"
	"os"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/keydb"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// Prepare loads the initial data
func (st *Store) UpdatePrepare() error {
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
		if b.Header.Version > 1 {
			receipts, err := st.Receipts(st.cache.height)
			if err != nil {
				return err
			}
			st.cache.heightReceipts = append(types.Receipts{}, receipts...)
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

func (st *Store) CopyContext(zipContextPath string) error {
	if err := st.db.Shrink(); err != nil {
		return err
	}
	if in, err := os.Open(st.keydbPath); err != nil {
		return err
	} else {
		defer in.Close()

		st.closeLock.RLock()
		defer st.closeLock.RUnlock()
		if tempContextFile, err := os.Create(zipContextPath); err != nil {
			return err
		} else {
			if _, err = io.Copy(tempContextFile, in); err != nil {
				return err
			}
			if err = tempContextFile.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// StoreBlock stores the block
func (st *Store) UpdateContext(b *types.Block, ctx *types.Context, receipts types.Receipts) error {
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
	// Datas := [][]byte{bsHeader}
	// {
	// 	data, _, err := bin.WriterToBytes(&b.Body)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	Datas = append(Datas, data)
	// }
	// {
	// 	// if len(receipts) > 0 {
	// 	// 	log.Println(receipts)
	// 	// }
	// 	data, _, err := bin.WriterToBytes(&receipts)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	Datas = append(Datas, data)
	// }
	// if err := st.cdb.AppendData(b.Header.Height, HeaderHash, Datas); err != nil {
	// 	if errors.Cause(err) != piledb.ErrInvalidAppendHeight {
	// 		return err
	// 	}
	// }

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
	st.cache.heightReceipts = append(types.Receipts{}, receipts...)
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

// StoreGenesis stores the genesis data
func (st *Store) UpdateStoreGenesis(genHash hash.Hash256, ctd *types.ContextData) error {
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
	st.cache.heightReceipts = nil
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
