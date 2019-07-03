package chain

import (
	"runtime"
	"sync"

	"github.com/fletaio/fleta/common/util"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// Chain manages the chain data using processes
type Chain struct {
	sync.Mutex
	isInit          bool
	store           *Store
	consensus       Consensus
	app             types.Application
	processes       []types.Process
	processIDMap    map[string]uint8
	processIndexMap map[uint8]int
	services        []types.Service
	serviceMap      map[string]types.Service
	closeLock       sync.RWMutex
	isClose         bool
}

// NewChain returns a Chain
func NewChain(consensus Consensus, app types.Application, store *Store) *Chain {
	cn := &Chain{
		consensus:       consensus,
		app:             app,
		store:           store,
		processes:       []types.Process{},
		processIDMap:    map[string]uint8{},
		processIndexMap: map[uint8]int{},
		services:        []types.Service{},
		serviceMap:      map[string]types.Service{},
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

	// Init
	for i, p := range cn.processes {
		if err := p.Init(types.NewRegister(IDMap[i]), cn, cn.Provider()); err != nil {
			return err
		}
	}
	if err := cn.app.Init(types.NewRegister(255), cn, cn.Provider()); err != nil {
		return err
	}
	if err := cn.consensus.Init(cn, newChainCommiter(cn)); err != nil {
		return err
	}

	// InitGenesis
	genesisContext := types.NewEmptyContext()
	if err := cn.app.InitGenesis(types.NewContextWrapper(255, genesisContext)); err != nil {
		return err
	}
	if err := cn.consensus.InitGenesis(types.NewContextWrapper(0, genesisContext)); err != nil {
		return err
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

	// OnLoadChain
	ctx := types.NewContext(cn.store)
	for i, p := range cn.processes {
		if err := p.OnLoadChain(types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	if err := cn.app.OnLoadChain(types.NewContextWrapper(255, ctx)); err != nil {
		return err
	}
	if err := cn.consensus.OnLoadChain(types.NewContextWrapper(0, ctx)); err != nil {
		return err
	}

	cn.isInit = true
	return nil
}

// Provider returns a chain provider
func (cn *Chain) Provider() types.Provider {
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
func (cn *Chain) Processes() []types.Process {
	list := []types.Process{}
	for _, p := range cn.processes {
		list = append(list, p)
	}
	return list
}

// Process returns the process by the id
func (cn *Chain) Process(id uint8) (types.Process, error) {
	if id == 255 {
		return cn.app, nil
	}
	idx, has := cn.processIndexMap[id]
	if !has {
		return nil, types.ErrNotExistProcess
	}
	return cn.processes[idx], nil
}

// ProcessByName returns the process by the name
func (cn *Chain) ProcessByName(name string) (types.Process, error) {
	id, has := cn.processIDMap[name]
	if !has {
		return nil, types.ErrNotExistProcess
	}
	idx, has := cn.processIndexMap[id]
	if !has {
		return nil, types.ErrNotExistProcess
	}
	return cn.processes[idx], nil
}

// SwitchProcess changes the context process to the target process
func (cn *Chain) SwitchProcess(ctw *types.ContextWrapper, p types.Process, fn func(cts *types.ContextWrapper) error) error {
	id, has := cn.processIDMap[p.Name()]
	if !has {
		return types.ErrNotExistProcess
	}
	if err := fn(types.SwitchContextWrapper(id, ctw)); err != nil {
		return err
	}
	return nil
}

// MustAddProcess adds Process but panic when has the same name process
func (cn *Chain) MustAddProcess(id uint8, p types.Process) {
	if cn.isInit {
		panic(ErrAddBeforeChainInit)
	}
	if id == 0 || id == 255 {
		panic(ErrReservedID)
	}
	if _, has := cn.processIDMap[p.Name()]; has {
		panic(types.ErrExistProcessName)
	}
	if _, has := cn.processIndexMap[id]; has {
		panic(types.ErrExistProcessID)
	}
	idx := len(cn.processes)
	cn.processes = append(cn.processes, p)
	cn.processIDMap[p.Name()] = id
	cn.processIndexMap[id] = idx
}

// Services returns services
func (cn *Chain) Services() []types.Service {
	list := []types.Service{}
	for _, s := range cn.services {
		list = append(list, s)
	}
	return list
}

// ServiceByName returns the service by the name
func (cn *Chain) ServiceByName(name string) (types.Service, error) {
	s, has := cn.serviceMap[name]
	if !has {
		return nil, ErrNotExistService
	}
	return s, nil
}

// MustAddService adds Service but panic when has the same name service
func (cn *Chain) MustAddService(s types.Service) {
	if cn.isInit {
		panic(ErrAddBeforeChainInit)
	}
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
	if err := cn.consensus.ValidateSignature(&b.Header, b.Signatures); err != nil {
		return err
	}

	ctx := types.NewContext(cn.store)
	if err := cn.executeBlockOnContext(b, ctx); err != nil {
		return err
	}
	return cn.connectBlockWithContext(b, ctx)
}

func (cn *Chain) connectBlockWithContext(b *types.Block, ctx *types.Context) error {
	IDMap := map[int]uint8{}
	for id, idx := range cn.processIndexMap {
		IDMap[idx] = id
	}

	if b.Header.ContextHash != ctx.Hash() {
		return ErrInvalidContextHash
	}

	// OnSaveData
	for i, p := range cn.processes {
		if err := p.OnSaveData(b, types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	if err := cn.app.OnSaveData(b, types.NewContextWrapper(255, ctx)); err != nil {
		return err
	}
	if err := cn.consensus.OnSaveData(b, types.NewContextWrapper(0, ctx)); err != nil {
		return err
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
	if err := cn.validateTransactionSignatures(b, ctx); err != nil {
		return err
	}
	IDMap := map[int]uint8{}
	for id, idx := range cn.processIndexMap {
		IDMap[idx] = id
	}

	// BeforeExecuteTransactions
	for i, p := range cn.processes {
		if err := p.BeforeExecuteTransactions(types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	if err := cn.app.BeforeExecuteTransactions(types.NewContextWrapper(255, ctx)); err != nil {
		return err
	}

	// Execute Transctions
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		pid := uint8(t >> 8)
		p, err := cn.Process(pid)
		if err != nil {
			return err
		}
		if err := tx.Execute(p, types.NewContextWrapper(pid, ctx), uint16(i)); err != nil {
			return err
		}
	}

	// AfterExecuteTransactions
	for i, p := range cn.processes {
		if err := p.AfterExecuteTransactions(b, types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		}
	}
	if err := cn.app.AfterExecuteTransactions(b, types.NewContextWrapper(255, ctx)); err != nil {
		return err
	}
	if err := cn.consensus.AfterExecuteTransactions(b, types.NewContextWrapper(0, ctx)); err != nil {
		return err
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
	return nil
}

func (cn *Chain) validateTransactionSignatures(b *types.Block, ctx *types.Context) error {
	var wg sync.WaitGroup
	cpuCnt := runtime.NumCPU()
	if len(b.Transactions) < 1000 {
		cpuCnt = 1
	}
	txUnit := len(b.Transactions) / cpuCnt
	TxHashes := make([]hash.Hash256, len(b.Transactions)+1)
	TxHashes[0] = b.Header.PrevHash
	if len(b.Transactions)%cpuCnt != 0 {
		txUnit++
	}
	errs := make(chan error, cpuCnt)
	defer close(errs)
	for i := 0; i < cpuCnt; i++ {
		lastCnt := (i + 1) * txUnit
		if lastCnt > len(b.Transactions) {
			lastCnt = len(b.Transactions)
		}
		wg.Add(1)
		go func(sidx int, txs []types.Transaction) {
			defer wg.Done()
			for q, tx := range txs {
				t := b.TransactionTypes[sidx+q]
				sigs := b.TranactionSignatures[sidx+q]

				TxHash := HashTransaction(t, tx)
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
				if err := tx.Validate(types.NewContextWrapper(uint8(t>>8), ctx), signers); err != nil {
					errs <- err
					return
				}
			}
		}(i*txUnit, b.Transactions[i*txUnit:lastCnt])
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
