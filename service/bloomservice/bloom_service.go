package bloomservice

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core/bloombits"
	"github.com/meverselabs/meverse/ethereum/params"
)

type BloomBitService struct {
	cn         *chain.Chain
	initHeight uint32
	indexer    *ChainIndexer

	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	closeBloomHandler chan struct{}
}

// NewBloomBitService starts a bloom service that performs bloom bit indexer and returns it
func NewBloomBitService(cn *chain.Chain, path string, size, confirms uint64) (*BloomBitService, error) {

	provider := cn.Provider()
	initHeight := provider.InitHeight()
	indexer, err := NewBloomIndexer(cn, path, size, confirms)
	if err != nil {
		return nil, err
	}

	b := &BloomBitService{
		cn:                cn,
		initHeight:        initHeight,
		indexer:           indexer,
		closeBloomHandler: make(chan struct{}),
		bloomRequests:     make(chan chan *bloombits.Retrieval),
	}

	// when initHeight > 0
	// 	1. zeroblock을 만들어서 1 section을 완성할 수 있도록 한다.
	// 	2. checkpoint = section - 1
	//  3. if initheight = section 의 마지막인경우 ex.  sectionSize = 24, height 23
	//   		h = init hash
	// 	   else h = common.Hash{}

	section := initHeight / uint32(indexer.sectionSize)
	var h common.Hash
	if initHeight > 0 && (initHeight+1)%uint32(indexer.sectionSize) == 0 {
		h, err = provider.Hash(initHeight)
		if err != nil {
			return nil, err
		}
		indexer.AddCheckpoint(uint64(section), h)
	} else {
		h = common.Hash{}
		if section > 0 {
			indexer.AddCheckpoint(uint64(section-1), h)
		}
	}

	b.startBloomHandlers(params.BloomBitsBlocks)
	return b, nil
}

// Name returns the name of the service
func (b *BloomBitService) Name() string {
	return "fleta.bloombit"
}

// OnLoadChain called when the chain loaded
func (b *BloomBitService) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (b *BloomBitService) OnBlockConnected(block *types.Block, loader types.Loader) {

	b.indexer.newHead(uint64(block.Header.Height))
	c := b.indexer

	if c.knownSections > c.storedSections {
		// Cache the current section count and head to allow unlocking the mutex
		c.verifyLastHead()
		section := c.storedSections
		var oldHead common.Hash
		if section > 0 {
			oldHead = c.SectionHead(section - 1)
		}
		// Process the newly defined section in the background
		// c.lock.Unlock()
		newHead, err := c.processSection(section, oldHead)
		if err != nil {
			return
		}
		// c.lock.Lock()

		// If processing succeeded and no reorgs occurred, mark the section completed
		if err == nil && (section == 0 || oldHead == c.SectionHead(section-1)) {
			c.setSectionHead(section, newHead)
			c.setValidSections(section + 1)
		}
	}

}

// OnTransactionInPoolExpired called when a transaction in pool is expired
func (b *BloomBitService) OnTransactionInPoolExpired(txs []*types.Transaction) {}

// OnTransactionFail called when a transaction in pool is expired
func (b *BloomBitService) OnTransactionFail(height uint32, txs []*types.Transaction, err []error) {}

// Close closes Database
func (b *BloomBitService) Close() error {
	close(b.closeBloomHandler)
	return b.indexer.Close()
}

// BloomStatus retrives the bloom size and sections
func (b *BloomBitService) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.indexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *BloomBitService) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.bloomRequests)
	}
}
