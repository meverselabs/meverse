package pile

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fletaio/fleta/common/hash"
)

// DB provides stack like value store using piles
type DB struct {
	sync.Mutex
	path         string
	piles        []*Pile
	genHash      hash.Hash256
	syncMode     bool
	hasDirty     bool
	lastSyncTime time.Time
}

// Open creates a DB that includes loaded piles
func Open(path string) (*DB, error) {
	os.MkdirAll(path, os.ModePerm)

	start := time.Now()
	var MaxHeight uint32
	pileMap := map[uint32]*Pile{}
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			if filepath.Ext(p) == ".pile" {
				p, err := LoadPile(p)
				if err != nil {
					return err
				}
				pileMap[p.BeginHeight] = p
				if MaxHeight < p.HeadHeight {
					MaxHeight = p.HeadHeight
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	Count := MaxHeight/ChunkUnit + 1
	piles := make([]*Pile, 0, Count)
	if MaxHeight > 0 {
		for i := uint32(0); i < Count; i++ {
			if p, has := pileMap[i*ChunkUnit]; !has {
				return nil, ErrMissingPile
			} else {
				piles = append(piles, p)
			}
		}
	}
	log.Println("PileDB is opened in", time.Now().Sub(start))
	db := &DB{
		path:         path,
		piles:        piles,
		lastSyncTime: time.Now(),
	}
	if len(piles) > 0 {
		copy(db.genHash[:], db.piles[0].GenHash[:])
	}

	timer := time.NewTimer(time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
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
				timer.Reset(time.Second)
			}
		}
	}()

	return db, nil
}

// Init initialize database when not initialized
func (db *DB) Init(genHash hash.Hash256) error {
	db.Lock()
	defer db.Unlock()

	if len(db.piles) > 0 {
		return ErrAlreadyInitialized
	}

	p, err := NewPile(filepath.Join(db.path, "chain_"+strconv.Itoa(len(db.piles)+1)+".pile"), genHash, 0)
	if err != nil {
		return err
	}
	db.piles = append(db.piles, p)
	db.genHash = genHash
	return nil
}

// Close closes pile DB
func (db *DB) Close() {
	db.Lock()
	defer db.Unlock()

	start := time.Now()
	for _, p := range db.piles {
		p.Close()
	}
	log.Println("PileDB is closed in", time.Now().Sub(start))
	db.piles = []*Pile{}
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
		return ErrExeedMaximumDataArrayLength
	}

	p := db.piles[len(db.piles)-1]
	if Height != p.HeadHeight+1 {
		return ErrInvalidAppendHeight
	}
	if p.HeadHeight-p.BeginHeight >= ChunkUnit {
		if len(db.piles) > 0 {
			db.piles[len(db.piles)-1].file.Sync()
		}
		v, err := NewPile(filepath.Join(db.path, "chain_"+strconv.Itoa(len(db.piles)+1)+".pile"), db.genHash, p.BeginHeight+ChunkUnit)
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
			return hash.Hash256{}, ErrInvalidHeight
		}
	}

	idx := (Height - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return hash.Hash256{}, ErrInvalidHeight
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

	if Height == 0 {
		return nil, ErrInvalidHeight
	}

	idx := (Height - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return nil, ErrInvalidHeight
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

	if Height == 0 {
		return nil, ErrInvalidHeight
	}

	idx := (Height - 1) / ChunkUnit
	if len(db.piles) <= int(idx) {
		return nil, ErrInvalidHeight
	}
	p := db.piles[idx]

	data, err := p.GetDatas(Height, from, count)
	if err != nil {
		return nil, err
	}
	return data, nil
}
