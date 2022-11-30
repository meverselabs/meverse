package bloomservice

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	mtypes "github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core/bloombits"
)

const (
	// bloomThrottling is the time to wait between processing two consecutive index
	// sections. It's useful during chain upgrades to prevent disk overload.
	bloomThrottling = 100 * time.Millisecond
)

type dbConfig struct {
	DatabaseHandles int `toml:"-"`
	DatabaseCache   int
}

var defaultConfig = dbConfig{
	DatabaseCache: 512,
}

// BloomIndexer implements a core.ChainIndexer, building up a rotated bloom bits index
// for the Ethereum header bloom filters, permitting blazing fast filtering.
type BloomIndexer struct {
	cn      *chain.Chain         // chain to get ctx, block, receipts
	size    uint64               // section size to generate bloombits for
	db      ethdb.Database       // database instance to write index data and metadata into
	gen     *bloombits.Generator // generator to rotate the bloom bits crating the bloom index
	section uint64               // Section is the section number being processed currently
	head    common.Hash          // Head is the hash of the last header processed
	lock    sync.Mutex
}

// NewBloomIndexer returns a bloom indexer that generates bloom bits data for the
// canonical chain for fast logs filtering.
func NewBloomIndexer(cn *chain.Chain, path string, size, confirms uint64) (*ChainIndexer, error) {

	backend := &BloomIndexer{
		cn:   cn,
		size: size,
	}

	// Assemble the Ethereum object
	db, err := backend.OpenDatabase(path, defaultConfig.DatabaseCache, defaultConfig.DatabaseHandles, "mev/db/bloombit/", false)
	if err != nil {
		return nil, err
	}

	backend.db = db
	table := rawdb.NewTable(db, string(rawdb.BloomBitsIndexPrefix))

	return NewChainIndexer(cn, db, table, backend, size, confirms, bloomThrottling, "bloombits"), nil
}

// Reset implements core.ChainIndexerBackend, starting a new bloombits index
// section.
func (b *BloomIndexer) Reset(ctx context.Context, section uint64, lastSectionHead common.Hash) error {
	gen, err := bloombits.NewGenerator(uint(b.size))
	b.gen, b.section, b.head = gen, section, common.Hash{}
	return err
}

// Process implements core.ChainIndexerBackend, adding a new header's bloom into
// the index.
func (b *BloomIndexer) Process(ctx context.Context, block *mtypes.Block) error {
	blm, err := BlockLogsBloom(b.cn, block)
	if err != nil {
		return err
	}

	header := &block.Header
	b.gen.AddBloom(uint(uint64(header.Height)-b.section*b.size), blm)
	b.head = bin.MustWriterToHash(header)
	return nil
}

// Commit implements core.ChainIndexerBackend, finalizing the bloom section and
// writing it out into the database.
func (b *BloomIndexer) Commit() error {
	batch := b.db.NewBatch()
	for i := 0; i < etypes.BloomBitLength; i++ {
		bits, err := b.gen.Bitset(uint(i))
		if err != nil {
			return err
		}
		rawdb.WriteBloomBits(batch, uint(i), b.section, b.head, bitutil.CompressBytes(bits))
		//log.Println("bloomIndexer", i, "section", b.section, b.head, bits)
	}

	return batch.Write()
}

// OpenDatabase opens an existing database with the given name (or creates one if no
// previous can be found) from within the node's instance directory. If the node is
// ephemeral, a memory database is returned.
func (b *BloomIndexer) OpenDatabase(path string, cache, handles int, namespace string, readonly bool) (ethdb.Database, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	var db ethdb.Database
	var err error
	if path == "" {
		db = rawdb.NewMemoryDatabase()
	} else {
		db, err = rawdb.NewLevelDBDatabase(path, cache, handles, namespace, readonly)
	}
	return db, err
}
