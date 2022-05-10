package piledb

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// DB provides stack like value store using piles
type DB struct {
	sync.Mutex
	path          string
	piles         []*Pile
	genHash       hash.Hash256
	initHash      hash.Hash256
	initHeight    uint32
	initTimestamp uint64
	syncMode      bool
	hasDirty      bool
	lastSyncTime  time.Time
	isClosed      bool
}

// Open creates a DB that includes loaded piles
func Open(path string, initHash hash.Hash256, InitHeight uint32, InitTimestamp uint64) (*DB, error) {
	os.MkdirAll(path, os.ModePerm)

	//start := time.Now()
	var MinHeight uint32
	MaxHeight := InitHeight
	pileMap := map[uint32]*Pile{}
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			if filepath.Ext(p) == ".pile" {
				p, err := LoadPile(p)
				if err != nil {
					return err
				}
				if len(pileMap) == 0 || MinHeight > p.BeginHeight {
					MinHeight = p.BeginHeight
				}
				if len(pileMap) == 0 || MaxHeight < p.HeadHeight {
					MaxHeight = p.HeadHeight
				}
				pileMap[p.BeginHeight] = p
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if MinHeight < InitHeight {
		MinHeight = InitHeight
	}

	Count := (MaxHeight-MinHeight+(MinHeight%ChunkUnit))/ChunkUnit + 1
	piles := make([]*Pile, 0, Count)
	if (MaxHeight - MinHeight) > 0 {
		BaseIdx := MinHeight / ChunkUnit
		for i := uint32(0); i < Count; i++ {
			if p, has := pileMap[(BaseIdx+i)*ChunkUnit]; !has {
				return nil, errors.WithStack(ErrMissingPile)
			} else {
				piles = append(piles, p)
			}
		}
	}
	//log.Println("PileDB is opened in", time.Since(start))
	db := &DB{
		path:          path,
		piles:         piles,
		lastSyncTime:  time.Now(),
		initHash:      initHash,
		initHeight:    InitHeight,
		initTimestamp: InitTimestamp,
	}
	if len(piles) > 0 {
		copy(db.genHash[:], db.piles[0].GenHash[:])
		copy(db.initHash[:], db.piles[0].InitHash[:])
		db.initTimestamp = db.piles[0].InitTimestamp
	}

	go func() {
		for !db.isClosed {
			db.Lock()
			if len(db.piles) > 0 {
				if db.hasDirty {
					now := time.Now()
					if now.Sub(db.lastSyncTime) > time.Second {
						db.piles[len(db.piles)-1].file.Sync()
						db.lastSyncTime = now
						db.hasDirty = false
					}
				}
			}
			db.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return db, nil
}

// Init initialize database when not initialized
func (db *DB) Init(genHash hash.Hash256, initHash hash.Hash256, initHeight uint32, initTimestamp uint64) error {
	db.Lock()
	defer db.Unlock()

	if len(db.piles) > 0 {
		return errors.WithStack(ErrAlreadyInitialized)
	}

	p, err := NewPile(filepath.Join(db.path, "chain_"+strconv.Itoa(len(db.piles)+1)+".pile"), genHash, initHash, initHeight, initTimestamp, (initHeight/ChunkUnit)*ChunkUnit)
	if err != nil {
		return err
	}
	db.piles = append(db.piles, p)
	db.genHash = genHash
	db.initHash = initHash
	db.initHeight = initHeight
	db.initTimestamp = initTimestamp
	return nil
}

// Close closes pile DB
func (db *DB) Close() {
	db.Lock()
	defer db.Unlock()

	db.isClosed = true
	start := time.Now()
	for _, p := range db.piles {
		p.Close()
	}
	log.Println("PileDB is closed in", time.Since(start))
	db.piles = []*Pile{}
}

// InitHeight returns init height
func (db *DB) InitHeight() uint32 {
	return db.initHeight
}

// InitTimestamp returns init timestamp
func (db *DB) InitTimestamp() uint64 {
	return db.initTimestamp
}

// SetSyncMode changes sync mode(sync every second when disabled)
func (db *DB) SetSyncMode(sync bool) {
	db.Lock()
	defer db.Unlock()

	db.syncMode = sync
}

// AppendData pushes data to top of the pile in piles
func (db *DB) AppendData(Height uint32, DataHash hash.Hash256, Datas [][]byte) error {
	db.Lock()
	defer db.Unlock()

	if len(Datas) > 255 {
		return errors.WithStack(ErrExeedMaximumDataArrayLength)
	}

	p := db.piles[len(db.piles)-1]
	if Height != p.HeadHeight+1 {
		return errors.WithStack(ErrInvalidAppendHeight)
	}
	if p.HeadHeight-p.BeginHeight >= ChunkUnit {
		if len(db.piles) > 0 {
			db.piles[len(db.piles)-1].file.Sync()
		}
		v, err := NewPile(filepath.Join(db.path, "chain_"+strconv.Itoa(len(db.piles)+1)+".pile"), db.genHash, db.initHash, db.initHeight, db.initTimestamp, p.BeginHeight+ChunkUnit)
		if err != nil {
			return err
		}
		db.piles = append(db.piles, v)
		p = v
	}

	sync := db.syncMode
	now := time.Now()
	if !sync {
		if db.lastSyncTime.Sub(now) >= time.Second {
			sync = true
		}
	}
	if err := p.AppendData(sync, Height, DataHash, Datas); err != nil {
		return err
	}
	if sync {
		db.lastSyncTime = now
		db.hasDirty = false
	} else {
		db.hasDirty = true
	}
	return nil
}

// GetHash returns a hash value of the height
func (db *DB) GetHash(Height uint32) (hash.Hash256, error) {
	db.Lock()
	defer db.Unlock()

	if Height == 0 {
		if len(db.piles) > 0 {
			return db.piles[0].GenHash, nil
		} else {
			return hash.Hash256{}, errors.WithStack(ErrInvalidHeight)
		}
	}
	if Height < db.initHeight {
		return hash.Hash256{}, errors.WithStack(ErrUnderInitHeight)
	}
	if Height == db.initHeight {
		if len(db.piles) > 0 {
			return db.piles[0].InitHash, nil
		} else {
			return hash.Hash256{}, errors.WithStack(ErrInvalidHeight)
		}
	}

	idx := (Height - db.initHeight + (db.initHeight % ChunkUnit) - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return hash.Hash256{}, errors.WithStack(ErrInvalidHeight)
	}
	p := db.piles[idx]

	h, err := p.GetHash(Height)
	if err != nil {
		return hash.Hash256{}, err
	}
	return h, nil
}

// GetData returns a data at the index of the height
func (db *DB) GetData(Height uint32, index int) ([]byte, error) {
	db.Lock()
	defer db.Unlock()

	if Height <= db.initHeight {
		return nil, errors.WithStack(ErrUnderInitHeight)
	}

	idx := (Height - db.initHeight + (db.initHeight % ChunkUnit) - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return nil, errors.WithStack(ErrInvalidHeight)
	}
	p := db.piles[idx]

	data, err := p.GetData(Height, index)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetDatas returns datas of the height between from and from + count
func (db *DB) GetDatas(Height uint32, from int, count int) ([]byte, error) {
	db.Lock()
	defer db.Unlock()

	if Height <= db.initHeight {
		return nil, errors.WithStack(ErrUnderInitHeight)
	}

	idx := (Height - db.initHeight + (db.initHeight % ChunkUnit) - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return nil, errors.WithStack(ErrInvalidHeight)
	}
	p := db.piles[idx]

	data, err := p.GetDatas(Height, from, count)
	if err != nil {
		return nil, err
	}
	return data, nil
}
