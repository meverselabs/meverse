package pof

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/bluele/gcache"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

type genItem struct {
	BlockGen *BlockGenMessage
	ObSign   *BlockObSignMessage
	Context  *types.Context
	Recv     bool
}

// FormulatorConfig defines configuration of the formulator
type FormulatorConfig struct {
	Formulator              common.Address
	MaxTransactionsPerBlock int
}

// FormulatorNode procudes a block by the consensus
type FormulatorNode struct {
	sync.Mutex
	Config         *FormulatorConfig
	cs             *Consensus
	ms             *FormulatorNodeMesh
	nm             *p2p.NodeMesh
	key            key.Key
	ndkey          key.Key
	myPublicHash   common.PublicHash
	frPublicHash   common.PublicHash
	statusLock     sync.Mutex
	genLock        sync.Mutex
	lastReqLock    sync.Mutex
	lastGenItemMap map[uint32]*genItem
	lastReqMessage *BlockReqMessage
	lastGenHeight  uint32
	lastGenTime    int64
	statusMap      map[string]*p2p.Status
	obStatusMap    map[string]*p2p.Status
	requestTimer   *p2p.RequestTimer
	requestLock    sync.RWMutex
	blockQ         *queue.SortedQueue
	txpool         *txpool.TransactionPool
	txQ            *queue.ExpireQueue
	txWaitQ        *queue.LinkedQueue
	txSendQ        *queue.Queue
	recvChan       chan *p2p.RecvMessageItem
	sendChan       chan *p2p.SendMessageItem
	singleCache    gcache.Cache
	batchCache     gcache.Cache
	sigCache       gcache.Cache
	isRunning      bool
	closeLock      sync.RWMutex
	isClose        bool
}

// NewFormulatorNode returns a FormulatorNode
func NewFormulatorNode(Config *FormulatorConfig, key key.Key, ndkey key.Key, NetAddressMap map[common.PublicHash]string, SeedNodeMap map[common.PublicHash]string, cs *Consensus, peerStorePath string) *FormulatorNode {
	if Config.MaxTransactionsPerBlock == 0 {
		Config.MaxTransactionsPerBlock = 7000
	}
	fr := &FormulatorNode{
		Config:         Config,
		cs:             cs,
		key:            key,
		ndkey:          ndkey,
		myPublicHash:   common.NewPublicHash(ndkey.PublicKey()),
		frPublicHash:   common.NewPublicHash(key.PublicKey()),
		lastGenItemMap: map[uint32]*genItem{},
		statusMap:      map[string]*p2p.Status{},
		obStatusMap:    map[string]*p2p.Status{},
		requestTimer:   p2p.NewRequestTimer(nil),
		blockQ:         queue.NewSortedQueue(),
		txpool:         txpool.NewTransactionPool(),
		txQ:            queue.NewExpireQueue(),
		txWaitQ:        queue.NewLinkedQueue(),
		txSendQ:        queue.NewQueue(),
		recvChan:       make(chan *p2p.RecvMessageItem, 1000),
		sendChan:       make(chan *p2p.SendMessageItem, 1000),
		singleCache:    gcache.New(500).LRU().Build(),
		batchCache:     gcache.New(500).LRU().Build(),
		sigCache:       gcache.New(100000).LRU().Build(),
	}
	fr.ms = NewFormulatorNodeMesh(key, NetAddressMap, fr)
	fr.nm = p2p.NewNodeMesh(fr.cs.cn.Provider().ChainID(), ndkey, SeedNodeMap, fr, peerStorePath)
	fr.txQ.AddGroup(60 * time.Second)
	fr.txQ.AddGroup(600 * time.Second)
	fr.txQ.AddGroup(3600 * time.Second)
	fr.txQ.AddHandler(fr)
	rlog.SetRLogAddress("fr:" + Config.Formulator.String())
	return fr
}

// Init initializes formulator
func (fr *FormulatorNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenMessage"), &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})
	fc.Register(types.DefineHashedType("p2p.TransactionMessage"), &p2p.TransactionMessage{})
	fc.Register(types.DefineHashedType("p2p.PeerListMessage"), &p2p.PeerListMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestPeerListMessage"), &p2p.RequestPeerListMessage{})
	return nil
}

// Close terminates the formulator
func (fr *FormulatorNode) Close() {
	fr.closeLock.Lock()
	defer fr.closeLock.Unlock()

	fr.Lock()
	defer fr.Unlock()

	fr.isClose = true
	fr.cs.cn.Close()
}

// Run runs the formulator
func (fr *FormulatorNode) Run(BindAddress string) {
	fr.Lock()
	if fr.isRunning {
		fr.Unlock()
		return
	}
	fr.isRunning = true
	fr.Unlock()

	go fr.ms.Run()
	go fr.nm.Run(BindAddress)
	go fr.requestTimer.Run()

	WorkerCount := 1
	switch runtime.NumCPU() {
	case 4:
		WorkerCount = 2
	case 5:
		WorkerCount = 2
	case 6:
		WorkerCount = 3
	case 7:
		WorkerCount = 4
	case 8:
		WorkerCount = 5
	default:
		WorkerCount = runtime.NumCPU()/2 + 1
		if WorkerCount >= runtime.NumCPU() {
			WorkerCount = runtime.NumCPU() - 1
		}
		if WorkerCount < 1 {
			WorkerCount = 1
		}
	}

	for i := 0; i < WorkerCount; i++ {
		go func() {
			for !fr.isClose {
				Count := 0
				ctw := fr.cs.cn.Provider().NewLoaderWrapper(1)
				currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
				for !fr.isClose {
					v := fr.txWaitQ.Pop()
					if v == nil {
						break
					}

					item := v.(*p2p.TxMsgItem)
					slot := types.ToTimeSlot(item.Tx.Timestamp())
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
						} else if slot > currentSlot+10 {
							continue
						}
					}
					if ctw.IsUsedTimeSlot(slot, string(item.TxHash[:])) {
						continue
					}
					if err := fr.addTx(ctw, item.TxHash, item.Type, item.Tx, item.Sigs); err != nil {
						if err != p2p.ErrInvalidUTXO && err != txpool.ErrExistTransaction && err != txpool.ErrTransactionPoolOverflowed && err != types.ErrUsedTimeSlot && err != types.ErrInvalidTransactionTimeSlot {
							rlog.Println("TransactionError", item.TxHash.String(), err.Error())
							panic(err) //TEMP

							if len(item.PeerID) > 0 {
								fr.nm.AddBadPoint(item.PeerID, 1)
							}
						}
						continue
					}
					//rlog.Println("TransactionAppended", item.TxHash.String(), Count)

					fr.txSendQ.Push(item)

					Count++
					if Count > 500 {
						break
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	go func() {
		for !fr.isClose {
			if fr.nm.HasPeer() {
				msg := &p2p.TransactionMessage{
					Types:      []uint16{},
					Txs:        []types.Transaction{},
					Signatures: [][]common.Signature{},
				}
				currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
				for {
					v := fr.txSendQ.Pop()
					if v == nil {
						break
					}
					m := v.(*p2p.TxMsgItem)
					slot := types.ToTimeSlot(m.Tx.Timestamp())
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
						} else if slot > currentSlot+10 {
							continue
						}
					}
					msg.Types = append(msg.Types, m.Type)
					msg.Txs = append(msg.Txs, m.Tx)
					msg.Signatures = append(msg.Signatures, m.Sigs)
					if len(msg.Types) >= 800 {
						break
					}
				}
				if len(msg.Types) > 0 {
					//log.Println("Send.TransactionMessage", len(msg.Types))
					fr.broadcastMessage(1, msg)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for i := 0; i < 2; i++ {
		go func() {
			for item := range fr.recvChan {
				if fr.isClose {
					break
				}
				m, err := p2p.PacketToMessage(item.Packet)
				if err != nil {
					log.Println("PacketToMessage", err)
					fr.nm.RemovePeer(item.PeerID)
					continue
				}
				if err := fr.handlePeerMessage(item.PeerID, m); err != nil {
					log.Println("handlePeerMessage", err)
					fr.nm.RemovePeer(item.PeerID)
					continue
				}
			}
		}()
	}

	for i := 0; i < 2; i++ {
		go func() {
			for item := range fr.sendChan {
				if fr.isClose {
					break
				}
				var EmptyHash common.PublicHash
				if bytes.Equal(item.Target[:], EmptyHash[:]) {
					fr.nm.BroadcastPacket(item.Packet)
				} else {
					if item.Except {
						fr.nm.ExceptCast(string(item.Target[:]), item.Packet)
					} else {
						fr.nm.SendTo(item.Target, item.Packet)
					}
				}
			}
		}()
	}

	go func() {
		for !fr.isClose {
			fr.tryRequestBlocks()
			fr.tryRequestNext()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for !fr.isClose {
		fr.Lock()
		hasItem := false
		TargetHeight := uint64(fr.cs.cn.Provider().Height() + 1)
		Count := 0
		item := fr.blockQ.PopUntil(TargetHeight)
		for item != nil {
			b := item.(*types.Block)
			gi, has := fr.lastGenItemMap[b.Header.Height]
			isConnected := false
			if has {
				if gi.BlockGen != nil && gi.Context != nil {
					if gi.BlockGen.Block.Header.Generator == b.Header.Generator {
						if err := fr.cs.ct.ConnectBlockWithContext(b, gi.Context); err != nil {
							log.Println("blockQ.ConnectBlockWithContext", err)
						} else {
							isConnected = true
						}
					}
				}
			}
			if !isConnected {
				ChainID := fr.cs.cn.Provider().ChainID()
				sm := map[hash.Hash256][]common.PublicHash{}
				for i, tx := range b.Transactions {
					t := b.TransactionTypes[i]
					TxHash := chain.HashTransactionByType(ChainID, t, tx)
					item := fr.txpool.Get(TxHash)
					if item != nil {
						sm[TxHash] = item.Signers
					} else {
						if v, err := fr.sigCache.Get(TxHash); err != nil {
						} else if v != nil {
							sm[TxHash] = []common.PublicHash{v.(common.PublicHash)} //TEMP
						}
					}
				}
				if err := fr.cs.cn.ConnectBlock(b, sm); err != nil {
					break
				}
			}
			fr.cleanPool(b)
			rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))

			txs := fr.txpool.Clean(types.ToTimeSlot(b.Header.Timestamp))
			if len(txs) > 0 {
				svcs := fr.cs.cn.Services()
				for _, s := range svcs {
					s.OnTransactionInPoolExpired(txs)
				}
				rlog.Println("Transaction EXPIRED", len(txs))
			}

			fr.lastReqLock.Lock()
			if fr.lastReqMessage != nil {
				if b.Header.Height <= fr.lastReqMessage.TargetHeight+fr.cs.maxBlocksPerFormulator {
					if b.Header.Generator != fr.Config.Formulator {
						fr.lastReqMessage = nil
					}
				}
			}
			fr.lastReqLock.Unlock()

			delete(fr.lastGenItemMap, b.Header.Height)
			TargetHeight++
			Count++
			if Count > 10 {
				break
			}
			item = fr.blockQ.PopUntil(TargetHeight)
			hasItem = true
		}
		fr.Unlock()

		if hasItem {
			fr.broadcastStatus()
			fr.tryRequestBlocks()
		}

		if Count < 10 {
			if hasItem {
				time.Sleep(50 * time.Millisecond)
			} else {
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
}

// AddTx adds tx to txpool
func (fr *FormulatorNode) AddTx(tx types.Transaction, sigs []common.Signature) error {
	currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp())
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return types.ErrInvalidTransactionTimeSlot
		} else if slot > currentSlot+10 {
			return types.ErrInvalidTransactionTimeSlot
		}
	}

	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), t, tx)
	ctw := fr.cs.cn.Provider().NewLoaderWrapper(1)
	if ctw.IsUsedTimeSlot(slot, string(TxHash[:])) {
		return types.ErrUsedTimeSlot
	}

	if err := fr.addTx(ctw, TxHash, t, tx, sigs); err != nil {
		return err
	}
	fr.txSendQ.Push(&p2p.TxMsgItem{
		TxHash: TxHash,
		Type:   t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

// PushTx pushes transaction
func (fr *FormulatorNode) PushTx(tx types.Transaction, sigs []common.Signature) error {
	currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp())
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return types.ErrInvalidTransactionTimeSlot
		} else if slot > currentSlot+10 {
			return types.ErrInvalidTransactionTimeSlot
		}
	}

	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), t, tx)
	if !fr.txpool.IsExist(TxHash) {
		fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
			TxHash: TxHash,
			Type:   t,
			Tx:     tx,
			Sigs:   sigs,
		})
	}
	return nil
}

func (fr *FormulatorNode) addTx(ctw types.LoaderWrapper, TxHash hash.Hash256, t uint16, tx types.Transaction, sigs []common.Signature) error {
	/*
		if fr.txpool.Size() > types.MaxTransactionPerBlock {
			return txpool.ErrTransactionPoolOverflowed
		}
	*/
	if fr.txpool.IsExist(TxHash) {
		return txpool.ErrExistTransaction
	}
	signers := make([]common.PublicHash, 0, len(sigs))
	if v, err := fr.sigCache.Get(TxHash); err != nil {
	} else if v != nil {
		signers = append(signers, v.(common.PublicHash))
	}
	if len(signers) == 0 {
		for _, sig := range sigs {
			pubkey, err := common.RecoverPubkey(TxHash, sig)
			if err != nil {
				return err
			}
			signers = append(signers, common.NewPublicHash(pubkey))
		}
	}
	pid := uint8(t >> 8)
	p, err := fr.cs.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw = types.NewLoaderWrapper(pid, ctw)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := fr.txpool.Push(t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	fr.txQ.Push(string(TxHash[:]), &p2p.TxMsgItem{
		Type: t,
		Tx:   tx,
		Sigs: sigs,
	})
	return nil
}

// OnTimerExpired called when rquest expired
func (fr *FormulatorNode) OnTimerExpired(height uint32, value string) {
	go fr.tryRequestBlocks()
}

// OnItemExpired is called when the item is expired
func (fr *FormulatorNode) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	item := Item.(*p2p.TxMsgItem)
	currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(item.Tx.Timestamp())
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return
		} else if slot > currentSlot+10 {
			return
		}
	}
	fr.txSendQ.Push(item)
	if IsLast {
		fr.txpool.Remove(item.TxHash, item.Tx)
	}
}

func (fr *FormulatorNode) addBlock(b *types.Block) error {
	cp := fr.cs.cn.Provider()
	if b.Header.Height <= cp.Height() {
		h, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if h != encoding.Hash(b.Header) {
			//TODO : critical error signal
			return chain.ErrFoundForkedBlock
		}
	} else {
		if item := fr.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
			}
		}
	}
	return nil
}

func (fr *FormulatorNode) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), t, tx)
		fr.txpool.Remove(TxHash, tx)
		fr.txQ.Remove(string(TxHash[:]))
	}
}

// TxPoolList returned tx list from txpool
func (fr *FormulatorNode) TxPoolList() []*txpool.PoolItem {
	return fr.txpool.List()
}

// GetTxFromTXPool returned tx from txpool
func (fr *FormulatorNode) GetTxFromTXPool(TxHash hash.Hash256) *txpool.PoolItem {
	return fr.txpool.Get(TxHash)
}
