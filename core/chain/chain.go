package chain

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"time"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"

	"github.com/meverselabs/meverse/core/chain/admin"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core"

	"github.com/pkg/errors"
)

// Chain manages the chain data
type Chain struct {
	sync.Mutex
	isInit         bool
	store          *Store
	services       []types.Service
	serviceMap     map[string]types.Service
	observerKeyMap map[common.PublicKey]bool
	closeLock      sync.RWMutex
	waitChan       map[uuid.UUID]*common.SyncChan
	waitLock       sync.Mutex
	isClose        bool
	tag            string
}

// NewChain returns a Chain
func NewChain(ObserverKeys []common.PublicKey, store *Store, tag string) *Chain {
	ObserverKeyMap := map[common.PublicKey]bool{}
	for _, v := range ObserverKeys {
		ObserverKeyMap[v] = true
	}
	cn := &Chain{
		store:          store,
		observerKeyMap: ObserverKeyMap,
		services:       []types.Service{},
		serviceMap:     map[string]types.Service{},
		waitChan:       map[uuid.UUID]*common.SyncChan{},
		tag:            tag,
	}
	return cn
}

// Init initializes the chain
func (cn *Chain) Init(genesisContextData *types.ContextData) error {
	cn.Lock()
	defer cn.Unlock()

	GenesisHash := hash.Hashes(hash.Hash(cn.store.ChainID().Bytes()), genesisContextData.Hash())
	Height := cn.store.Height()
	if Height > 0 {
		if h, err := cn.store.Hash(0); err != nil {
			return err
		} else {
			if GenesisHash != h {
				return errors.WithStack(piledb.ErrInvalidGenesisHash)
			}
		}
	} else {
		if err := cn.store.StoreGenesis(GenesisHash, genesisContextData); err != nil {
			return err
		}
	}
	if err := cn.store.Prepare(); err != nil {
		return err
	}

	// OnLoadChain
	ctx := types.NewContext(cn.store)
	for _, s := range cn.services {
		if err := s.OnLoadChain(ctx); err != nil {
			return err
		}
	}

	//log.Println("Chain loaded", cn.store.Height(), ctx.PrevHash().String())

	cn.isInit = true
	return nil
}

// InitWith initializes the chain with snapshot informations
func (cn *Chain) InitWith(InitGenesisHash hash.Hash256, initHash hash.Hash256, initHeight uint32, initTimestamp uint64) error {
	cn.Lock()
	defer cn.Unlock()

	if initHeight == 0 {
		return errors.WithStack(piledb.ErrInvalidInitialHash)
	}

	Height := cn.store.Height()
	if Height > initHeight {
		if h, err := cn.store.Hash(0); err != nil {
			return err
		} else {
			if InitGenesisHash != h {
				return errors.WithStack(piledb.ErrInvalidGenesisHash)
			}
		}
		if h, err := cn.store.Hash(initHeight); err != nil {
			return err
		} else {
			if initHash != h {
				return errors.WithStack(piledb.ErrInvalidInitialHash)
			}
		}
	} else {
		if err := cn.store.StoreInit(InitGenesisHash, initHash, initHeight, initTimestamp); err != nil {
			return err
		}
	}
	if err := cn.store.Prepare(); err != nil {
		return err
	}

	// OnLoadChain
	ctx := types.NewContext(cn.store)
	for _, s := range cn.services {
		if err := s.OnLoadChain(ctx); err != nil {
			return err
		}
	}

	//log.Println("Chain loaded", cn.store.Height(), ctx.PrevHash().String())

	cn.isInit = true
	return nil
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

// Services returns services
func (cn *Chain) Services() []types.Service {
	list := []types.Service{}
	list = append(list, cn.services...)
	return list
}

// TopGenerator returns current top generator
func (cn *Chain) TopGenerator(TimeoutCount uint32) (common.Address, error) {
	return cn.store.TopGenerator(TimeoutCount)
}

// GeneratorInMap returns current top generator
func (cn *Chain) GeneratorsInMap(GeneratorMap map[common.Address]bool, Limit int) ([]common.Address, error) {
	return cn.store.GeneratorsInMap(GeneratorMap, Limit)
}

// TopRankInMap returns current top generator
func (cn *Chain) TopGeneratorInMap(GeneratorMap map[common.Address]bool) (common.Address, uint32, error) {
	return cn.store.TopGeneratorInMap(GeneratorMap)
}

// ServiceByName returns the service by the name
func (cn *Chain) ServiceByName(name string) (types.Service, error) {
	s, has := cn.serviceMap[name]
	if !has {
		return nil, errors.WithStack(ErrNotExistService)
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

// Provider returns the context of the chain
func (cn *Chain) Provider() types.Provider {
	return cn.store
}

// Store returns the store of the chain
func (cn *Chain) Store() *Store {
	return cn.store
}

// WaitConnectedBlock is wait untile target block stored
func (cn *Chain) WaitConnectedBlock(targetBlock uint32) {
	if cn.Provider().Height() >= targetBlock {
		return
	}
	id := uuid.New()
	wc := common.NewSyncChan()
	cn.waitLock.Lock()
	cn.waitChan[id] = wc
	cn.waitLock.Unlock()
	defer func() {
		cn.waitLock.Lock()
		delete(cn.waitChan, id)
		cn.waitLock.Unlock()
		wc.Close()
	}()

	done := make(chan bool)
	go func() {
		defer close(done)
		for cn.Provider().Height() < targetBlock {
			time.Sleep(time.Millisecond * 10)
		}
	}()

	var conntced uint32
	for conntced < targetBlock {
		select {
		case data := <-wc.Chan:
			conntced = data.(uint32)
		case <-done:
			return
		}
	}
}

// NewContext returns the context of the chain
func (cn *Chain) NewContext() *types.Context {
	return types.NewContext(cn.store)
}

// ConnectBlock try to connect block to the chain
func (cn *Chain) ConnectBlock(b *types.Block, SigMap map[hash.Hash256]common.Address) error {
	cn.closeLock.RLock()
	defer cn.closeLock.RUnlock()
	if cn.isClose {
		return errors.WithStack(ErrChainClosed)
	}

	cn.Lock()
	defer cn.Unlock()

	if err := cn.validateHeader(&b.Header); err != nil {
		return err
	}
	if err := cn.ValidateSignature(&b.Header, b.Body.BlockSignatures); err != nil {
		return err
	}
	ctx := types.NewContext(cn.store)
	if receipts, err := cn.executeBlockOnContext(b, ctx, SigMap); err != nil {
		return err
	} else {
		return cn.connectBlockWithContext(b, ctx, receipts)
	}
}

func (cn *Chain) ValidateSignature(bh *types.Header, sigs []common.Signature) error {
	Top, err := cn.store.rankTable.TopRank(bh.TimeoutCount)
	if err != nil {
		return err
	}
	if Top.Address != bh.Generator {
		return errors.WithStack(ErrInvalidTopAddress)
	}

	GeneratorSignature := sigs[0]
	h, _, err := bin.WriterToHash(bh)
	if err != nil {
		return err
	}
	pubkey, err := common.RecoverPubkey(bh.ChainID, h, GeneratorSignature)
	if err != nil {
		return err
	}
	if Top.Address != pubkey.Address() {
		return errors.WithStack(ErrInvalidTopSignature)
	}

	if len(sigs) != len(cn.observerKeyMap)/2+2 {
		return errors.WithStack(ErrInvalidSignatureCount)
	}
	KeyMap := map[common.PublicKey]bool{}
	for pubkey := range cn.observerKeyMap {
		KeyMap[pubkey] = true
	}
	bs := types.BlockSign{
		HeaderHash:         h,
		GeneratorSignature: sigs[0],
	}
	ObserverSignatures := sigs[1:]
	sh, _, err := bin.WriterToHash(&bs)
	if err != nil {
		return err
	}
	if err := common.ValidateSignaturesMajority(bh.ChainID, sh, ObserverSignatures, KeyMap); err != nil {
		return err
	}
	return nil
}

func (cn *Chain) connectBlockWithContext(b *types.Block, ctx *types.Context, receipts types.Receipts) error {
	if b.Header.ContextHash != ctx.Hash() {
		log.Println("CONNECT", ctx.Hash(), b.Header.ContextHash, ctx.Dump())
		panic("")
		// return errors.WithStack(ErrInvalidContextHash)
	}

	if cn.store.Version(b.Header.Height) > 1 {
		if b.Header.ReceiptHash != bin.MustWriterToHash(&receipts) {
			log.Println("CONNECT", bin.MustWriterToHash(&receipts), b.Header.ReceiptHash, receipts)
			panic("")
			return errors.WithStack(ErrInvalidReceiptHash)
		}
	}

	if ctx.StackSize() > 1 {
		return errors.WithStack(types.ErrDirtyContext)
	}

	if err := cn.store.StoreBlock(b, ctx, receipts); err != nil {
		return err
	}
	var ca []*common.SyncChan
	cn.waitLock.Lock()
	for _, c := range cn.waitChan {
		ca = append(ca, c)
	}
	cn.waitLock.Unlock()
	for _, c := range ca {
		c.Send(b.Header.Height)
	}

	for _, s := range cn.services {
		s.OnBlockConnected(b.Clone(), ctx)
	}
	return nil
}

func (cn *Chain) executeBlockOnContext(b *types.Block, ctx *types.Context, sm map[hash.Hash256]common.Address) (types.Receipts, error) {
	TxSigners, TxHashes, err := cn.validateTransactionSignatures(b, sm)
	if err != nil {
		return nil, err
	}

	types.CheckABI(b, cn.NewContext())

	// Execute Transctions
	currentSlot := types.ToTimeSlot(b.Header.Timestamp)
	receipts := types.Receipts{}
	for i, tx := range b.Body.Transactions {
		slot := types.ToTimeSlot(tx.Timestamp)
		if slot < currentSlot-1 {
			return nil, errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		} else if slot > currentSlot {
			return nil, errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}

		sn := ctx.Snapshot()
		if err := ctx.UseTimeSlot(slot, string(TxHashes[i][:])); err != nil {
			ctx.Revert(sn)
			return nil, err
		}
		TXID := types.TransactionID(b.Header.Height, uint16(len(b.Body.Transactions)))
		if tx.VmType != types.Evm {
			if tx.To == common.ZeroAddr {
				if !ctx.IsAdmin(TxSigners[i]) {
					ctx.Revert(sn)
					return nil, errors.WithStack(ErrInvalidAdminAddress)
				}
				if _, err := cn.ExecuteTransaction(ctx, tx, TXID); err != nil {
					ctx.Revert(sn)
					return nil, err
				}
			} else {
				if err := ExecuteContractTx(ctx, tx, TxSigners[i], TXID); err != nil {
					ctx.Revert(sn)
					return nil, err
				}
			}
			receipt := new(etypes.Receipt)
			receipts = append(receipts, receipt)
		} else {
			if _, receipt, err := cn.ApplyEvmTransaction(ctx, tx, uint16(i), TxSigners[i]); err != nil {
				ctx.Revert(sn)
				return nil, err
			} else {
				receipts = append(receipts, receipt)
			}
		}
		ctx.Commit(sn)
	}
	if ctx.StackSize() > 1 {
		return nil, errors.WithStack(types.ErrDirtyContext)
	}
	if b.Header.Height%prefix.RewardIntervalBlocks == 0 {
		if _, err := ctx.ProcessReward(ctx, b); err != nil {
			return nil, err
		}
	}
	if ctx.StackSize() > 1 {
		return nil, errors.WithStack(types.ErrDirtyContext)
	}
	return receipts, nil
}

func (cn *Chain) validateHeader(bh *types.Header) error {
	height, lastHash := cn.store.LastStatus()
	if bh.ChainID.Cmp(cn.store.ChainID()) != 0 {
		return errors.Wrapf(ErrInvalidChainID, "chainid %v, %v", bh.ChainID, cn.store.ChainID())
	}
	if bh.Version > cn.store.Version(bh.Height) {
		return errors.WithStack(ErrInvalidVersion)
	}
	if bh.PrevHash != lastHash {
		return errors.WithStack(ErrInvalidPrevHash)
	}
	if bh.Timestamp <= cn.store.LastTimestamp() {
		return errors.WithStack(ErrInvalidTimestamp)
	}
	var emptyAddr common.Address
	if bh.Generator == emptyAddr {
		return errors.WithStack(ErrInvalidGenerator)
	}
	if bh.Height != height+1 {
		return errors.WithStack(ErrInvalidHeight)
	}

	if bh.Height == cn.store.InitHeight()+1 {
		if bh.Version <= 0 {
			return errors.WithStack(ErrInvalidVersion)
		}
	} else {
		TargetHeader, err := cn.store.Header(height)
		if err != nil {
			return err
		}
		if bh.Version < TargetHeader.Version {
			return errors.WithStack(ErrInvalidVersion)
		}
		if bh.ChainID.Cmp(TargetHeader.ChainID) != 0 {
			return errors.WithStack(ErrInvalidChainID)
		}
	}
	return nil
}

func (cn *Chain) validateTransactionSignatures(b *types.Block, SigMap map[hash.Hash256]common.Address) ([]common.Address, []hash.Hash256, error) {
	TxHashes := make([]hash.Hash256, len(b.Body.Transactions)+1)
	TxHashes[0] = b.Header.PrevHash
	TxSigners := make([]common.Address, len(b.Body.Transactions))
	if len(b.Body.Transactions) > 0 {
		var wg sync.WaitGroup
		cpuCnt := runtime.NumCPU()
		if len(b.Body.Transactions) < 1000 {
			cpuCnt = 1
		}
		txUnit := len(b.Body.Transactions) / cpuCnt
		if len(b.Body.Transactions)%cpuCnt != 0 {
			txUnit++
		}
		errs := make(chan error, cpuCnt)
		defer close(errs)
		for i := 0; i < cpuCnt; i++ {
			lastCnt := (i + 1) * txUnit
			if lastCnt > len(b.Body.Transactions) {
				lastCnt = len(b.Body.Transactions)
			}
			wg.Add(1)
			go func(sidx int, txs []*types.Transaction) {
				defer wg.Done()
				for q, tx := range txs {
					TxHash := tx.Hash(b.Header.Height)
					TxHashes[sidx+q+1] = TxHash
					hasSigner := false
					if SigMap != nil {
						if s, has := SigMap[TxHash]; has {
							TxSigners[sidx+q] = s
							hasSigner = true
						}
					}
					if !hasSigner {
						sig := b.Body.TransactionSignatures[sidx+q]
						pubkey, err := common.RecoverPubkey(tx.ChainID, tx.Message(), sig)
						if err != nil {
							errs <- err
							return
						}
						TxSigners[sidx+q] = pubkey.Address()
					}
				}
			}(i*txUnit, b.Body.Transactions[i*txUnit:lastCnt])
		}
		wg.Wait()
		if len(errs) > 0 {
			err := <-errs
			return nil, nil, err
		}
	}
	if h, err := BuildLevelRoot(TxHashes); err != nil {
		return nil, nil, err
	} else if b.Header.LevelRootHash != h {
		return nil, nil, errors.WithStack(ErrInvalidLevelRootHash)
	}
	return TxSigners, TxHashes[1:], nil
}

func (cn *Chain) ExecuteTransaction(ctx *types.Context, tx *types.Transaction, TXID string) ([]*ctypes.Event, error) {
	types.ExecLock.Lock()
	defer types.ExecLock.Unlock()

	switch tx.Method {
	case admin.AdminAdd:
		return nil, ctx.SetAdmin(common.BytesToAddress(tx.Args), true)
	case admin.AdminRemove:
		return nil, ctx.SetAdmin(common.BytesToAddress(tx.Args), false)
	case admin.GeneratorAdd:
		return nil, ctx.SetGenerator(common.BytesToAddress(tx.Args), true)
	case admin.GeneratorRemove:
		return nil, ctx.SetGenerator(common.BytesToAddress(tx.Args), false)
	case admin.ContractDeploy:
		data := &DeployContractData{}
		if _, err := data.ReadFrom(bytes.NewReader(tx.Args)); err != nil {
			return nil, err
		}
		if cont, err := ctx.DeployContract(data.Owner, data.ClassID, data.Args); err != nil {
			return nil, err
		} else {
			addr := cont.Address()
			_, i, err := types.ParseTransactionID(TXID)
			if err != nil {
				return nil, err
			}
			return []*ctypes.Event{{
				Index:  i,
				Type:   ctypes.EventTagTxMsg,
				Result: bin.TypeWriteAll(addr),
			}}, nil
		}
	case admin.TransactionSetBasicFee:
		if iss, err := bin.TypeReadAll(tx.Args, 1); err != nil {
			return nil, err
		} else if fee, ok := iss[0].(*amount.Amount); ok {
			ctx.SetBasicFee(fee)
		} else {
			return nil, errors.WithStack(ErrInvalidBasicFee)
		}

		_, i, err := types.ParseTransactionID(TXID)
		if err != nil {
			return nil, err
		}
		return []*ctypes.Event{{
			Index:  i,
			Type:   ctypes.EventTagTxMsg,
			Result: tx.Args,
		}}, nil
	default:
		return nil, errors.WithStack(ErrUnknownTransactionMethod)
	}
}

// ethereum type transaction 실행
func (cn *Chain) ApplyEvmTransaction(ctx *types.Context, tx *types.Transaction, ti uint16, signer common.Address) ([]*ctypes.Event, *etypes.Receipt, error) {
	etx := new(etypes.Transaction)
	if err := etx.UnmarshalBinary([]byte(tx.Args)); err != nil {
		return nil, nil, err
	}

	if err := core.ValidateTx(etx, true); err != nil {
		return nil, nil, err
	}
	statedb := types.NewStateDB(ctx)
	statedb.Prepare(etx.Hash(), int(ti))
	receipt, evs, err := core.ApplyTransaction(statedb, etx)
	if err != nil {
		log.Println("core.ApplyTransaction", err)
		return nil, nil, err
	}

	ens := []*ctypes.Event{}
	if fee, err := ChargeFee(ctx, receipt.GasUsed, signer); err != nil {
		return nil, nil, err
	} else {
		en := &ctypes.Event{
			Index:  ti,
			Type:   ctypes.EventTagTxFee,
			Result: bin.TypeWriteAll(fee),
		}
		ens = append(ens, en)
	}

	if receipt.ContractAddress != (common.Address{}) {
		addr := receipt.ContractAddress
		en := &ctypes.Event{
			Index:  ti,
			Type:   ctypes.EventTagTxMsg,
			Result: bin.TypeWriteAll(addr),
		}
		ens = append(ens, en)
	}

	for _, ev := range evs {
		ev.Index = ti
		ens = append(ens, ev)
	}

	return ens, receipt, nil
}
