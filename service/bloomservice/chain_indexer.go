// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bloomservice

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	mtypes "github.com/meverselabs/meverse/core/types"
)

// ChainIndexerBackend defines the methods needed to process chain segments in
// the background and write the segment results into the database. These can be
// used to create filter blooms or CHTs.
type ChainIndexerBackend interface {
	// Reset initiates the processing of a new chain segment, potentially terminating
	// any partially completed operations (in case of a reorg).
	Reset(ctx context.Context, section uint64, prevHead common.Hash) error

	// Process crunches through the next header in the chain segment. The caller
	// will ensure a sequential order of headers.
	Process(ctx context.Context, block *mtypes.Block) error

	// Commit finalizes the section metadata and stores it into the database.
	Commit() error
}

// ChainIndexer does a post-processing job for equally sized sections of the
// canonical chain (like BlooomBits and CHT structures). A ChainIndexer is
// connected to the blockchain through the event system by starting a
// ChainHeadEventLoop in a goroutine.
//
// Further child ChainIndexers can be added which use the output of the parent
// section indexer. These child indexers receive new head notifications only
// after an entire section has been finished or in case of rollbacks that might
// affect already finished sections.
type ChainIndexer struct {
	cn         *chain.Chain        // chain to get ctx, block, receipts
	initHeight uint32              // initHeight
	chainDb    ethdb.Database      // Chain database to index the data from
	indexDb    ethdb.Database      // Prefixed table-view of the db to write index metadata into
	backend    ChainIndexerBackend // Background processor generating the index data content
	//children []*ChainIndexer     // Child indexers to cascade chain updates to

	active    uint32          // Flag whether the event loop was started
	update    chan struct{}   // Notification channel that headers should be processed
	quit      chan chan error // Quit channel to tear down running goroutines
	ctx       context.Context
	ctxCancel func()

	sectionSize uint64 // Number of blocks in a single chain segment to process
	confirmsReq uint64 // Number of confirmations before processing a completed segment

	storedSections uint64 // Number of sections successfully indexed into the database
	knownSections  uint64 // Number of sections known to be complete (block wise)
	// cascadedHead   uint64 // Block number of the last completed section cascaded to subindexers

	checkpointSections uint64      // Number of sections covered by the checkpoint
	checkpointHead     common.Hash // Section head belonging to the checkpoint

	// throttling time.Duration // Disk throttling to prevent a heavy upgrade from hogging resources

	log  log.Logger
	lock sync.Mutex
}

// NewChainIndexer creates a new chain indexer to do background processing on
// chain segments of a given size after certain number of confirmations passed.
// The throttling parameter might be used to prevent database thrashing.
func NewChainIndexer(cn *chain.Chain, chainDb ethdb.Database, indexDb ethdb.Database, backend ChainIndexerBackend, section, confirm uint64, throttling time.Duration, kind string) *ChainIndexer {
	c := &ChainIndexer{
		cn:          cn,
		initHeight:  cn.Provider().InitHeight(),
		chainDb:     chainDb,
		indexDb:     indexDb,
		backend:     backend,
		update:      make(chan struct{}, 1),
		quit:        make(chan chan error),
		sectionSize: section,
		confirmsReq: confirm,
		//throttling:  throttling,
		log: log.New("type", kind),
	}
	// Initialize database dependent fields and start the updater
	// c.loadValidSections()
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())

	//	go c.updateLoop()

	return c
}

// AddCheckpoint adds a checkpoint. Sections are never processed and the chain
// is not expected to be available before this point. The indexer assumes that
// the backend has sufficient information available to process subsequent sections.
//
// Note: knownSections == 0 and storedSections == checkpointSections until
// syncing reaches the checkpoint
func (c *ChainIndexer) AddCheckpoint(section uint64, shead common.Hash) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Short circuit if the given checkpoint is below than local's.
	if c.checkpointSections >= section+1 || section < c.storedSections {
		return
	}
	c.checkpointSections = section + 1
	c.checkpointHead = shead

	c.setSectionHead(section, shead)
	c.setValidSections(section + 1)
}

// Close tears down all goroutines belonging to the indexer and returns any error
// that might have occurred internally.
func (c *ChainIndexer) Close() error {
	var errs []error

	c.ctxCancel()

	// Tear down the primary update loop
	errc := make(chan error)
	c.quit <- errc
	if err := <-errc; err != nil {
		errs = append(errs, err)
	}
	// If needed, tear down the secondary event loop
	if atomic.LoadUint32(&c.active) != 0 {
		c.quit <- errc
		if err := <-errc; err != nil {
			errs = append(errs, err)
		}
	}
	switch {
	case len(errs) == 0:
		return nil

	case len(errs) == 1:
		return errs[0]

	default:
		return fmt.Errorf("%v", errs)
	}
}

// newHead notifies the indexer about new chain heads and/or reorgs.
func (c *ChainIndexer) newHead(head uint64) {
	// c.lock.Lock()
	// defer c.lock.Unlock()

	// calculate the number of newly known sections and update if high enough
	var sections uint64
	if head >= c.confirmsReq {
		sections = (head + 1 - c.confirmsReq) / c.sectionSize
		if sections < c.checkpointSections {
			sections = 0
		}
		if sections > c.knownSections {
			c.knownSections = sections
		}
	}
}

// processSection processes an entire section by calling backend functions while
// ensuring the continuity of the passed headers. Since the chain mutex is not
// held while processing, the continuity can be broken by a long reorg, in which
// case the function returns with an error.
func (c *ChainIndexer) processSection(section uint64, lastHead common.Hash) (common.Hash, error) {
	c.log.Trace("Processing new chain section", "section", section)

	// Reset and partial processing
	if err := c.backend.Reset(c.ctx, section, lastHead); err != nil {
		c.setValidSections(0)
		return common.Hash{}, err
	}

	provider := c.cn.Provider()
	for number := section * c.sectionSize; number < (section+1)*c.sectionSize; number++ {

		block, hash, err := blockAndHash(provider, c.initHeight, uint32(number))
		if err != nil {
			return common.Hash{}, err
		}

		if number > uint64(c.initHeight) && hash == (common.Hash{}) {
			return common.Hash{}, fmt.Errorf("canonical block #%d unknown", number)
		}
		header := block.Header
		if header.PrevHash != lastHead {
			return common.Hash{}, fmt.Errorf("chain reorged during section processing")
		}
		if err := c.backend.Process(c.ctx, block); err != nil {
			return common.Hash{}, err
		}
		lastHead = hash
	}
	if err := c.backend.Commit(); err != nil {
		return common.Hash{}, err
	}
	return lastHead, nil
}

// blockAndHash returns block and blockhash
func blockAndHash(provider mtypes.Provider, initHeight uint32, height uint32) (*mtypes.Block, common.Hash, error) {
	if height < initHeight {
		return blankBlock(height), common.Hash{}, nil
	} else if height == initHeight {
		hash, err := provider.Hash(height)
		if err != nil {
			return nil, common.Hash{}, err
		}
		return blankBlock(height), hash, nil
	}
	block, err := provider.Block(uint32(height))
	if err != nil {
		return nil, common.Hash{}, err
	}
	hash := bin.MustWriterToHash(&block.Header)
	return block, hash, nil
}

func blankBlock(height uint32) *mtypes.Block {
	return &mtypes.Block{
		Header: mtypes.Header{
			Height:   height,
			PrevHash: common.Hash{},
		},
		Body: mtypes.Body{
			// Transactions:          []*mtypes.Transaction{},
			// Events:                []*mtypes.Event{},
			// TransactionSignatures: []mcommon.Signature{},
			// BlockSignatures:       []mcommon.Signature{},
		},
	}
}

// verifyLastHead compares last stored section head with the corresponding block hash in the
// actual canonical chain and rolls back reorged sections if necessary to ensure that stored
// sections are all valid
func (c *ChainIndexer) verifyLastHead() error {
	for c.storedSections > 0 && c.storedSections > c.checkpointSections {
		hash, err := c.cn.Provider().Hash(uint32(c.storedSections*c.sectionSize - 1))
		if err != nil {
			return err
		}
		if c.SectionHead(c.storedSections-1) == hash {
			return nil
		}
		c.setValidSections(c.storedSections - 1)
	}
	return nil
}

// Sections returns the number of processed sections maintained by the indexer
// and also the information about the last header indexed for potential canonical
// verifications.
func (c *ChainIndexer) Sections() (uint64, uint64, common.Hash) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.verifyLastHead()
	return c.storedSections, c.storedSections*c.sectionSize - 1, c.SectionHead(c.storedSections - 1)
}

// setValidSections writes the number of valid sections to the index database
func (c *ChainIndexer) setValidSections(sections uint64) {
	// Set the current number of valid sections in the database
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], sections)
	c.indexDb.Put([]byte("count"), data[:])

	// Remove any reorged sections, caching the valids in the mean time
	for c.storedSections > sections {
		c.storedSections--
		c.removeSectionHead(c.storedSections)
	}
	c.storedSections = sections // needed if new > old
}

// SectionHead retrieves the last block hash of a processed section from the
// index database.
func (c *ChainIndexer) SectionHead(section uint64) common.Hash {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], section)

	hash, _ := c.indexDb.Get(append([]byte("shead"), data[:]...))
	if len(hash) == len(common.Hash{}) {
		return common.BytesToHash(hash)
	}
	return common.Hash{}
}

// setSectionHead writes the last block hash of a processed section to the index
// database.
func (c *ChainIndexer) setSectionHead(section uint64, hash common.Hash) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], section)

	c.indexDb.Put(append([]byte("shead"), data[:]...), hash.Bytes())
}

// removeSectionHead removes the reference to a processed section from the index
// database.
func (c *ChainIndexer) removeSectionHead(section uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], section)

	c.indexDb.Delete(append([]byte("shead"), data[:]...))
}
