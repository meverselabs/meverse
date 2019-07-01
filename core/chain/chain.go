package chain

import (
	"runtime"
	"sync"

	"github.com/fletaio/fleta/common/util"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Chain manages the chain data using processes
type Chain struct {
	sync.Mutex
	store           *Store
	consensus       Consensus
	processes       []Process
	processIDMap    map[string]uint8
	processIndexMap map[uint8]int
	services        []Service
	serviceMap      map[string]Service
	closeLock       sync.RWMutex
	isClose         bool
}

// NewChain returns a Chain
func NewChain(consensus Consensus, store *Store) *Chain {
	cn := &Chain{
		consensus:       consensus,
		store:           store,
		processes:       []Process{},
		processIDMap:    map[string]uint8{},
		processIndexMap: map[uint8]int{},
		services:        []Service{},
		serviceMap:      map[string]Service{},
	}
	return cn
}

// Init initializes the chain
func (cn *Chain) Init() error {
	cn.Lock()
	defer cn.Unlock()

	IDMap := map[int]uint8{}
	for id, idx := range cn.processIndexMap {
		IDMap[idx] = id
	}
	ctx := types.NewContext(cn.store)
	if err := cn.consensus.Init(NewRegister(0), cn, newChainCommiter(cn), NewContextProcess(0, ctx)); err != nil {
		return err
	}
	for i, p := range cn.processes {
		if err := p.Init(NewRegister(IDMap[i]), cn, NewContextProcess(IDMap[i], ctx)); err != nil {
			return err
		}
	}

	genesisContext := types.NewContext(types.NewEmptyLoader())
	if err := cn.consensus.InitGenesis(NewContextProcess(0, genesisContext)); err != nil {
		return err
	}
	for i, p := range cn.processes {
		if err := p.InitGenesis(NewContextProcess(IDMap[i], genesisContext)); err != nil {
			return err
		}
	}
	if genesisContext.StackSize() > 1 {
		return ErrDirtyContext
	}
	top := genesisContext.Top()

	GenesisHash := hash.Hashes(hash.Hash([]byte(cn.store.Name())), hash.Hash(util.Uint16ToBytes(cn.store.Version())), genesisContext.Hash())
	if h, err := cn.Provider().Hash(0); err != nil {
		if err != ErrNotExistKey {
			return err
		} else {
			if err := cn.store.StoreGenesis(GenesisHash, top); err != nil {
				return err
			}
		}
	} else {
		if GenesisHash != h {
			return ErrInvalidGenesisHash
		}
	}
	return nil
}

// Provider returns a chain provider
func (cn *Chain) Provider() Provider {
	return cn.store
}

// Close terminates and cleans the chain
func (cn *Chain) Close() {
	cn.closeLock.Lock()
	defer cn.closeLock.Unlock()

	cn.Lock()
	defer cn.Unlock()

	cn.isClose = true
	cn.store.Close()
}

// Processes returns processes
func (cn *Chain) Processes() []Process {
	cn.Lock()
	defer cn.Unlock()

	list := []Process{}
	for _, p := range cn.processes {
		list = append(list, p)
	}
	return list
}

// Process returns the process by the id
func (cn *Chain) Process(id uint8) (Process, error) {
	cn.Lock()
	defer cn.Unlock()

	idx, has := cn.processIndexMap[id]
	if !has {
		return nil, ErrNotExistProcess
	}
	return cn.processes[idx], nil
}

// ProcessByName returns the process by the name
func (cn *Chain) ProcessByName(name string) (Process, error) {
	cn.Lock()
	defer cn.Unlock()

	id, has := cn.processIDMap[name]
	if !has {
		return nil, ErrNotExistProcess
	}
	idx, has := cn.processIndexMap[id]
	if !has {
		return nil, ErrNotExistProcess
	}
	return cn.processes[idx], nil
}

// MustAddProcess adds Process but panic when has the same name process
func (cn *Chain) MustAddProcess(id uint8, p Process) {
	cn.Lock()
	defer cn.Unlock()

	if id == 0 {
		panic(ErrZeroIDIsReservedForConsensus)
	}
	if _, has := cn.processIDMap[p.Name()]; has {
		panic(ErrExistProcessName)
	}
	if _, has := cn.processIndexMap[id]; has {
		panic(ErrExistProcessID)
	}
	idx := len(cn.processes)
	cn.processes = append(cn.processes, p)
	cn.processIDMap[p.Name()] = id
	cn.processIndexMap[id] = idx
}

// Services returns services
func (cn *Chain) Services() []Service {
	cn.Lock()
	defer cn.Unlock()

	list := []Service{}
	for _, s := range cn.services {
		list = append(list, s)
	}
	return list
}

// ServiceByName returns the service by the name
func (cn *Chain) ServiceByName(name string) (Service, error) {
	cn.Lock()
	defer cn.Unlock()

	s, has := cn.serviceMap[name]
	if !has {
		return nil, ErrNotExistService
	}
	return s, nil
}

// MustAddService adds Service but panic when has the same name service
func (cn *Chain) MustAddService(s Service) {
	cn.Lock()
	defer cn.Unlock()

	if _, has := cn.serviceMap[s.Name()]; has {
		panic(ErrExistServiceName)
	}
	cn.services = append(cn.services, s)
	cn.serviceMap[s.Name()] = s
}

// ConnectBlock try to connect block to the chain
func (cn *Chain) ConnectBlock(b *types.Block) error {
	cn.closeLock.RLock()
	defer cn.closeLock.RUnlock()
	if cn.isClose {
		return ErrChainClosed
	}

	cn.Lock()
	defer cn.Unlock()

	if err := cn.validateHeader(&b.Header); err != nil {
		return err
	}

	ctx := types.NewContext(cn.store)
	if err := cn.executeBlockOnContext(b, ctx); err != nil {
		return err
	}
	return cn.connectBlockWithContext(b, ctx)
}

func (cn *Chain) connectBlockWithContext(b *types.Block, ctx *types.Context) error {
	if b.Header.ContextHash != ctx.Hash() {
		return ErrInvalidContextHash
	}
	if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}
	top := ctx.Top()
	if err := cn.store.StoreBlock(b, top); err != nil {
		return err
	}
	for _, s := range cn.services {
		if err := s.OnBlockConnected(b, top.Events, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (cn *Chain) executeBlockOnContext(b *types.Block, ctx *types.Context) error {
	if err := cn.validateSignatures(b, ctx); err != nil {
		return err
	}
	IDMap := map[int]uint8{}
	for id, idx := range cn.processIndexMap {
		IDMap[idx] = id
	}
	for i, p := range cn.processes {
		if err := p.BeforeExecuteTransactions(b, NewContextProcess(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	for i, tx := range b.Transactions {
		if err := tx.Execute(ctx, uint16(i)); err != nil {
			return err
		}
	}
	for i, p := range cn.processes {
		if err := p.AfterExecuteTransactions(b, NewContextProcess(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	if err := cn.consensus.ProcessReward(b, NewContextProcess(0, ctx)); err != nil {
		return err
	}
	for i, p := range cn.processes {
		if err := p.ProcessReward(b, NewContextProcess(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	return nil
}

func (cn *Chain) validateHeader(bh *types.Header) error {
	provider := cn.Provider()
	if bh.Version > provider.Version() {
		return ErrInvalidVersion
	}
	if bh.Timestamp <= provider.LastTimestamp() {
		return ErrInvalidTimestamp
	}

	height := provider.Height()
	if bh.Height != height+1 {
		return ErrInvalidHeight
	}
	if bh.Height == 1 {
		if bh.Version <= 0 {
			return ErrInvalidVersion
		}
	} else {
		TargetHeader, err := provider.Header(height)
		if err != nil {
			return err
		}
		if bh.Version < TargetHeader.Version {
			return ErrInvalidVersion
		}
	}
	if bh.PrevHash != provider.LastHash() {
		return ErrInvalidPrevHash
	}
	if err := cn.consensus.ValidateHeader(bh); err != nil {
		return err
	}
	for _, p := range cn.processes {
		if err := p.ValidateHeader(bh); err != nil {
			return err
		}
	}
	return nil
}

func (cn *Chain) validateSignatures(b *types.Block, loader types.Loader) error {
	var wg sync.WaitGroup
	cpuCnt := runtime.NumCPU()
	if len(b.Transactions) < 1000 {
		cpuCnt = 1
	}
	txCnt := len(b.Transactions) / cpuCnt
	TxHashes := make([]hash.Hash256, len(b.Transactions)+1)
	TxHashes[0] = b.Header.PrevHash
	if len(b.Transactions)%cpuCnt != 0 {
		txCnt++
	}
	errs := make(chan error, cpuCnt)
	defer close(errs)
	for i := 0; i < cpuCnt; i++ {
		lastCnt := (i + 1) * txCnt
		if lastCnt > len(b.Transactions) {
			lastCnt = len(b.Transactions)
		}
		wg.Add(1)
		go func(sidx int, txs []types.Transaction) {
			defer wg.Done()
			for q, tx := range txs {
				sigs := b.Signatures[sidx+q]
				TxHash := encoding.Hash(tx)
				TxHashes[sidx+q+1] = TxHash

				signers := make([]common.PublicHash, 0, len(sigs))
				for _, sig := range sigs {
					pubkey, err := common.RecoverPubkey(TxHash, sig)
					if err != nil {
						errs <- err
						return
					}
					signers = append(signers, common.NewPublicHash(pubkey))
				}
				if err := tx.Validate(loader, signers); err != nil {
					errs <- err
					return
				}
			}
		}(i*txCnt, b.Transactions[i*txCnt:lastCnt])
	}
	wg.Wait()
	if len(errs) > 0 {
		err := <-errs
		return err
	}
	if h, err := BuildLevelRoot(TxHashes); err != nil {
		return err
	} else if b.Header.LevelRootHash != h {
		return ErrInvalidLevelRootHash
	}
	return nil
}
