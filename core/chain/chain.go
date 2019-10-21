package chain

import (
	"log"
	"runtime"
	"sync"

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
	if app.ID() != 255 {
		panic(ErrApplicationIDMustBe255)
	}
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
	for _, s := range cn.services {
		if err := s.Init(cn, cn.Provider()); err != nil {
			return err
		}
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

	GenesisHash := hash.Hashes(hash.Hash([]byte(cn.store.Name())), hash.Hash([]byte{cn.store.ChainID()}), genesisContext.Hash())
	if cn.store.Height() > 0 {
		if h, err := cn.store.Hash(0); err != nil {
			return err
		} else {
			if GenesisHash != h {
				return ErrInvalidGenesisHash
			}
		}
	} else {
		if err := cn.store.StoreGenesis(GenesisHash, top); err != nil {
			return err
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
	for _, s := range cn.services {
		if err := s.OnLoadChain(ctx); err != nil {
			return err
		}
	}

	log.Println("Chain loaded", cn.store.Height(), ctx.LastHash().String())

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

	if !cn.isClose {
		cn.store.Close()
		cn.isClose = true
	}
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

// MustAddProcess adds Process but panic when has the same name process
func (cn *Chain) MustAddProcess(p types.Process) {
	if cn.isInit {
		panic(ErrAddBeforeChainInit)
	}
	if p.ID() == 0 || p.ID() == 255 {
		panic(ErrReservedID)
	}
	if _, has := cn.processIDMap[p.Name()]; has {
		panic(types.ErrExistProcessName)
	}
	if _, has := cn.processIndexMap[p.ID()]; has {
		panic(types.ErrExistProcessID)
	}
	idx := len(cn.processes)
	cn.processes = append(cn.processes, p)
	cn.processIDMap[p.Name()] = p.ID()
	cn.processIndexMap[p.ID()] = idx
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

// NewContext returns the context of the chain
func (cn *Chain) NewContext() *types.Context {
	return types.NewContext(cn.store)
}

// ConnectBlock try to connect block to the chain
func (cn *Chain) ConnectBlock(b *types.Block, SigMap map[hash.Hash256][]common.PublicHash) error {
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
	if err := cn.executeBlockOnContext(b, ctx, SigMap); err != nil {
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
		log.Println(ctx.Dump())
		return ErrInvalidContextHash
	}

	if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}

	// OnSaveData
	for i, p := range cn.processes {
		if err := p.OnSaveData(b, types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		} else if ctx.StackSize() > 1 {
			return ErrDirtyContext
		}
	}
	if err := cn.app.OnSaveData(b, types.NewContextWrapper(255, ctx)); err != nil {
		return err
	} else if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}
	if err := cn.consensus.OnSaveData(b, types.NewContextWrapper(0, ctx)); err != nil {
		return err
	} else if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}

	top := ctx.Top()
	if err := cn.store.StoreBlock(b, top); err != nil {
		return err
	}
	for _, s := range cn.services {
		s.OnBlockConnected(b, top.Events, ctx)
	}
	return nil
}

func (cn *Chain) executeBlockOnContext(b *types.Block, ctx *types.Context, sm map[hash.Hash256][]common.PublicHash) error {
	TxSigners, err := cn.validateTransactionSignatures(b, sm)
	if err != nil {
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
		} else if ctx.StackSize() > 1 {
			return ErrDirtyContext
		}
	}
	if err := cn.app.BeforeExecuteTransactions(types.NewContextWrapper(255, ctx)); err != nil {
		return err
	} else if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}

	// Execute Transctions
	for i, tx := range b.Transactions {
		signers := TxSigners[i]
		t := b.TransactionTypes[i]
		pid := uint8(t >> 8)
		p, err := cn.Process(pid)
		if err != nil {
			return err
		}
		ctw := types.NewContextWrapper(pid, ctx)

		sn := ctw.Snapshot()
		if err := tx.Validate(p, ctw, signers); err != nil {
			ctw.Revert(sn)
			return err
		}
		if at, is := tx.(AccountTransaction); is {
			if at.Seq() != ctw.Seq(at.From())+1 {
				ctw.Revert(sn)
				return err
			}
			ctw.AddSeq(at.From())
			Result := uint8(0)
			if err := tx.Execute(p, ctw, uint16(i)); err != nil {
				Result = 0
			} else {
				Result = 1
			}
			if Result != b.TransactionResults[i] {
				return ErrInvalidResult
			}
		} else {
			if err := tx.Execute(p, ctw, uint16(i)); err != nil {
				ctw.Revert(sn)
				return err
			}
			if 1 != b.TransactionResults[i] {
				return ErrInvalidResult
			}
		}
		if Has, err := ctw.HasAccount(b.Header.Generator); err != nil {
			ctw.Revert(sn)
			if err == types.ErrDeletedAccount {
				return ErrCannotDeleteGeneratorAccount
			} else {
				return err
			}
		} else if !Has {
			ctw.Revert(sn)
			return ErrCannotDeleteGeneratorAccount
		}
		ctw.Commit(sn)
	}

	if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}

	// AfterExecuteTransactions
	for i, p := range cn.processes {
		if err := p.AfterExecuteTransactions(b, types.NewContextWrapper(IDMap[i], ctx)); err != nil {
			return err
		} else if ctx.StackSize() > 1 {
			return ErrDirtyContext
		}
	}
	if err := cn.app.AfterExecuteTransactions(b, types.NewContextWrapper(255, ctx)); err != nil {
		return err
	} else if ctx.StackSize() > 1 {
		return ErrDirtyContext
	}
	return nil
}

func (cn *Chain) validateHeader(bh *types.Header) error {
	provider := cn.Provider()
	height, lastHash := provider.LastStatus()
	if bh.ChainID != provider.ChainID() {
		return ErrInvalidChainID
	}
	if bh.Version > provider.Version() {
		return ErrInvalidVersion
	}
	if bh.PrevHash != lastHash {
		return ErrInvalidPrevHash
	}
	if bh.Timestamp <= provider.LastTimestamp() {
		return ErrInvalidTimestamp
	}
	if bh.Generator == common.NewAddress(0, 0, 0) {
		return ErrInvalidGenerator
	}

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
		if bh.ChainID != TargetHeader.ChainID {
			return ErrInvalidChainID
		}
	}
	return nil
}

func (cn *Chain) validateTransactionSignatures(b *types.Block, sm map[hash.Hash256][]common.PublicHash) ([][]common.PublicHash, error) {
	var wg sync.WaitGroup
	cpuCnt := runtime.NumCPU()
	if len(b.Transactions) < 1000 {
		cpuCnt = 1
	}
	txUnit := len(b.Transactions) / cpuCnt
	TxHashes := make([]hash.Hash256, len(b.Transactions)+1)
	TxSigners := make([][]common.PublicHash, len(b.Transactions))
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
				sigs := b.TransactionSignatures[sidx+q]

				TxHash := HashTransactionByType(cn.store.chainID, t, tx)
				TxHashes[sidx+q+1] = TxHash
				var signers []common.PublicHash
				if sm != nil {
					s, has := sm[TxHash]
					if !has {
						s = make([]common.PublicHash, 0, len(sigs))
						for _, sig := range sigs {
							pubkey, err := common.RecoverPubkey(TxHash, sig)
							if err != nil {
								errs <- err
								return
							}
							s = append(s, common.NewPublicHash(pubkey))
						}
					}
					signers = s
				}
				TxSigners[sidx+q] = signers
			}
		}(i*txUnit, b.Transactions[i*txUnit:lastCnt])
	}
	wg.Wait()
	if len(errs) > 0 {
		err := <-errs
		return nil, err
	}
	if h, err := BuildLevelRoot(TxHashes); err != nil {
		return nil, err
	} else if b.Header.LevelRootHash != h {
		return nil, ErrInvalidLevelRootHash
	}
	return TxSigners, nil
}
