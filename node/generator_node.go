package node

import (
	"bytes"
	"log"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/common/queue"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/txpool"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
)

type genItem struct {
	BlockGen *BlockGenMessage
	ObSign   *BlockObSignMessage
	Context  *types.Context
	Recv     bool
}

// GeneratorConfig defines configuration of the generator
type GeneratorConfig struct {
	MaxTransactionsPerBlock int
}

// GeneratorNode procudes a block by the consensus
type GeneratorNode struct {
	sync.Mutex
	ChainID        *big.Int
	Config         *GeneratorConfig
	cn             *chain.Chain
	ct             chain.Committer
	ms             *GeneratorNodeMesh
	nm             *p2p.NodeMesh
	key            key.Key
	ndkey          key.Key
	myPublicKey    common.PublicKey
	frPublicKey    common.PublicKey
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
	isRunning      bool
	closeLock      sync.RWMutex
	isClose        bool
}

// NewGeneratorNode returns a GeneratorNode
func NewGeneratorNode(ChainID *big.Int, Config *GeneratorConfig, cn *chain.Chain, key key.Key, ndkey key.Key, NetAddressMap map[common.PublicKey]string, SeedNodeMap map[common.PublicKey]string, peerStorePath string) *GeneratorNode {
	if Config.MaxTransactionsPerBlock == 0 {
		Config.MaxTransactionsPerBlock = 7000
	}
	fr := &GeneratorNode{
		Config:         Config,
		ChainID:        ChainID,
		cn:             cn,
		ct:             chain.NewChainCommiter(cn),
		key:            key,
		ndkey:          ndkey,
		myPublicKey:    ndkey.PublicKey(),
		frPublicKey:    key.PublicKey(),
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
	}
	fr.ms = NewGeneratorNodeMesh(key, NetAddressMap, fr)
	fr.nm = p2p.NewNodeMesh(fr.cn.Provider().ChainID(), ndkey, SeedNodeMap, fr, peerStorePath)
	fr.txQ.AddGroupRepeat(6, 10*time.Second)
	// fr.txQ.AddGroup(600 * time.Second)
	return fr
}

// Init initializes generator
func (fr *GeneratorNode) Init() error {
	return nil
}

// Close terminates the generator
func (fr *GeneratorNode) Close() {
	fr.closeLock.Lock()
	defer fr.closeLock.Unlock()

	fr.Lock()
	defer fr.Unlock()

	fr.isClose = true
	fr.cn.Close()
}

// OnItemExpired is called when the item is expired
func (fr *GeneratorNode) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) queue.TxExpiredType {
	item := Item.(*p2p.TxMsgItem)
	currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(item.Tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			fr.txpool.Remove(item.TxHash, item.Tx)
			return queue.Expired
		} else if slot > currentSlot+10 {
			return queue.Remain
		}
	}
	if item.Tx.IsEtherType {
		if item.Tx.From == common.ZeroAddr {
			fr.txpool.Remove(item.TxHash, item.Tx)
			return queue.Error
		}
		seq := fr.cn.Provider().AddrSeq(item.Tx.From)
		if seq+100 < item.Tx.Seq {
			return queue.Remain
		}
		if item.Tx.Seq < seq {
			fr.txpool.Remove(item.TxHash, item.Tx)
			return queue.Expired
		}
	}

	fr.txSendQ.Push(item)
	if IsLast {
		fr.txpool.Remove(item.TxHash, item.Tx)
	}
	return queue.Resend
}

// Run runs the generator
func (fr *GeneratorNode) Run(BindAddress string) {
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
				ctx := fr.cn.NewContext()
				currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
				for !fr.isClose {
					v := fr.txWaitQ.Pop()
					if v == nil {
						break
					}

					item := v.(*p2p.TxMsgItem)
					slot := types.ToTimeSlot(item.Tx.Timestamp)
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
						} else if slot > currentSlot+10 {
							continue
						}
					}
					if ctx.IsUsedTimeSlot(slot, string(item.TxHash[:])) {
						continue
					}
					if err := fr.addTx(item.TxHash, item.Tx, item.Sig); err != nil {
						if errors.Cause(err) != p2p.ErrInvalidUTXO &&
							errors.Cause(err) != txpool.ErrExistTransaction &&
							errors.Cause(err) != txpool.ErrTransactionPoolOverflowed &&
							errors.Cause(err) != types.ErrUsedTimeSlot &&
							errors.Cause(err) != types.ErrInvalidTransactionTimeSlot {
							log.Printf("TransactionError %v %+v\n", item.TxHash.String(), err)

							if len(item.PeerID) > 0 {
								fr.nm.AddBadPoint(item.PeerID, 1)
							}
						}
						continue
					}
					//log.Println("TransactionAppended", item.TxHash.String(), Count)

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
					Txs:        []*types.Transaction{},
					Signatures: []common.Signature{},
				}
				currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
				for {
					v := fr.txSendQ.Pop()
					if v == nil {
						break
					}
					m := v.(*p2p.TxMsgItem)
					slot := types.ToTimeSlot(m.Tx.Timestamp)
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
						} else if slot > currentSlot+10 {
							continue
						}
					}
					msg.Txs = append(msg.Txs, m.Tx)
					msg.Signatures = append(msg.Signatures, m.Sig)
					if len(msg.Txs) >= 1000 {
						break
					}
				}
				if len(msg.Txs) > 0 {
					//log.Println("Send.TransactionMessage", len(msg.Txs))
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
					log.Printf("PacketToMessage %+v\n", err)
					fr.nm.RemovePeer(item.PeerID)
					continue
				}
				if err := fr.handlePeerMessage(item.PeerID, m); err != nil {
					log.Printf("handlePeerMessage %+v\n", err)
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
				var EmptyHash common.PublicKey
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
		TargetHeight := uint64(fr.cn.Provider().Height() + 1)
		Count := 0
		item := fr.blockQ.PopUntil(TargetHeight)
		for item != nil {
			b := item.(*types.Block)
			gi, has := fr.lastGenItemMap[b.Header.Height]
			isConnected := false
			if has {
				if gi.BlockGen != nil && gi.Context != nil {
					if gi.BlockGen.Block.Header.Generator == b.Header.Generator {
						if err := fr.ct.ConnectBlockWithContext(b, gi.Context); err != nil {
							log.Printf("blockQ.ConnectBlockWithContext %+v\n", err)
						} else {
							isConnected = true
						}
					}
				}
			}
			if !isConnected {
				sm := map[hash.Hash256]common.Address{}
				for _, tx := range b.Body.Transactions {
					TxHash := tx.Hash(b.Header.Height)
					item := fr.txpool.Get(TxHash)
					if item != nil {
						sm[TxHash] = item.Signer
					}
				}
				if err := fr.cn.ConnectBlock(b, sm); err != nil {
					break
				}
			}
			fr.cleanPool(b)
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Body.Transactions))

			txs := fr.txpool.Clean(types.ToTimeSlot(b.Header.Timestamp))
			if len(txs) > 0 {
				svcs := fr.cn.Services()
				for _, s := range svcs {
					s.OnTransactionInPoolExpired(txs)
				}
				log.Println("Transaction EXPIRED", len(txs))
			}

			fr.lastReqLock.Lock()
			if fr.lastReqMessage != nil {
				if b.Header.Height <= fr.lastReqMessage.TargetHeight+prefix.MaxBlocksPerGenerator {
					if b.Header.Generator != fr.key.PublicKey().Address() {
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
func (fr *GeneratorNode) AddTx(tx *types.Transaction, sig common.Signature) error {
	currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		} else if slot > currentSlot+10 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}
	}

	TxHash := tx.HashSig()
	ctx := fr.cn.NewContext()
	if ctx.IsUsedTimeSlot(slot, string(TxHash[:])) {
		return errors.WithStack(types.ErrUsedTimeSlot)
	}

	if err := fr.addTx(TxHash, tx, sig); err != nil {
		return err
	}
	fr.txSendQ.Push(&p2p.TxMsgItem{
		TxHash: TxHash,
		Tx:     tx,
		Sig:    sig,
	})
	return nil
}

// PushTx pushes transaction
func (fr *GeneratorNode) PushTx(tx *types.Transaction, sig common.Signature) error {
	currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		} else if slot > currentSlot+10 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}
	}

	pubkey, err := common.RecoverPubkey(tx.ChainID, tx.Message(), sig)
	if err != nil {
		return err
	}
	tx.From = pubkey.Address()

	TxHash := tx.HashSig()
	if !fr.txpool.IsExist(TxHash) {
		fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
			TxHash: TxHash,
			Tx:     tx,
			Sig:    sig,
		})
	}
	return nil
}

func (fr *GeneratorNode) addTx(TxHash hash.Hash256, tx *types.Transaction, sig common.Signature) error {
	if fr.txpool.IsExist(TxHash) {
		return errors.WithStack(txpool.ErrExistTransaction)
	}
	pubkey, err := common.RecoverPubkey(tx.ChainID, tx.Message(), sig)
	if err != nil {
		return err
	}
	tx.From = pubkey.Address()
	if err := fr.txpool.Push(TxHash, tx, sig, tx.From); err != nil {
		return err
	}
	if tx.IsEtherType {
		cp := fr.cn.Provider()
		seq := cp.AddrSeq(tx.From)
		if tx.Seq < seq {
			return errors.WithStack(txpool.ErrPastSeq)
		} else if tx.Seq > seq+100 {
			return errors.WithStack(txpool.ErrTooFarSeq)
		}
	}
	fr.txQ.Push(string(TxHash[:]), &p2p.TxMsgItem{
		TxHash: TxHash,
		Tx:     tx,
		Sig:    sig,
	})
	return nil
}

// OnTimerExpired called when rquest expired
func (fr *GeneratorNode) OnTimerExpired(height uint32, value string) {
	go fr.tryRequestBlocks()
}

func (fr *GeneratorNode) addBlock(b *types.Block) error {
	cp := fr.cn.Provider()
	if b.Header.Height <= cp.Height() {
		h, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if h != bin.MustWriterToHash(&b.Header) {
			//TODO : critical error signal
			return errors.WithStack(chain.ErrFoundForkedBlock)
		}
	} else {
		if item := fr.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if bin.MustWriterToHash(&old.Header) != bin.MustWriterToHash(&b.Header) {
				//TODO : critical error signal
				return errors.WithStack(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}

func (fr *GeneratorNode) cleanPool(b *types.Block) {
	for _, tx := range b.Body.Transactions {
		TxHash := tx.HashSig()
		fr.txpool.Remove(TxHash, tx)
		fr.txQ.Remove(string(TxHash[:]))
	}
}

// TxPoolList returned tx list from txpool
func (fr *GeneratorNode) TxPoolList() []*txpool.PoolItem {
	return fr.txpool.List()
}

// GetTxFromTXPool returned tx from txpool
func (fr *GeneratorNode) GetTxFromTXPool(TxHash hash.Hash256) *txpool.PoolItem {
	return fr.txpool.Get(TxHash)
}
