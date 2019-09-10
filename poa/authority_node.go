package poa

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// AuthorityConfig defines configuration of the authority
type AuthorityConfig struct {
	Authority               common.Address
	MaxTransactionsPerBlock int
}

// AuthorityNode procudes a block by the consensus
type AuthorityNode struct {
	sync.Mutex
	Config       *AuthorityConfig
	cs           *Consensus
	ms           *ClientService
	key          key.Key
	ndkey        key.Key
	myPublicHash common.PublicHash
	statusLock   sync.Mutex
	authority    common.Address
	txMsgChans   []*chan *p2p.TxMsgItem
	txMsgIdx     uint64
	statusMap    map[string]*p2p.Status
	txpool       *txpool.TransactionPool
	txQ          *queue.ExpireQueue
	isRunning    bool
	closeLock    sync.RWMutex
	isClose      bool
	firstTime    uint64
	firstHeight  uint32
}

// NewAuthorityNode returns a AuthorityNode
func NewAuthorityNode(Config *AuthorityConfig, key key.Key, ndkey key.Key, SeedNodeMap map[common.PublicHash]string, cs *Consensus, peerStorePath string) *AuthorityNode {
	an := &AuthorityNode{
		Config:       Config,
		cs:           cs,
		key:          key,
		ndkey:        ndkey,
		myPublicHash: common.NewPublicHash(ndkey.PublicKey()),
		statusMap:    map[string]*p2p.Status{},
		txpool:       txpool.NewTransactionPool(),
	}
	an.ms = NewClientService(an)
	return an
}

// Close terminates the authority
func (an *AuthorityNode) Close() {
	an.closeLock.Lock()
	defer an.closeLock.Unlock()

	an.Lock()
	defer an.Unlock()

	an.isClose = true
	an.cs.cn.Close()
}

// Init initializes authority
func (an *AuthorityNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})
	fc.Register(types.DefineHashedType("p2p.TransactionMessage"), &p2p.TransactionMessage{})
	return nil
}

// Run runs the authority
func (an *AuthorityNode) Run(BindAddress string) {
	an.Lock()
	if an.isRunning {
		an.Unlock()
		return
	}
	an.isRunning = true
	an.Unlock()

	go an.ms.Run(BindAddress)

	an.firstTime = uint64(time.Now().UnixNano())
	an.firstHeight = an.cs.cn.Provider().Height()

	WorkerCount := runtime.NumCPU() - 1
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	workerEnd := make([]*chan struct{}, WorkerCount)
	an.txMsgChans = make([]*chan *p2p.TxMsgItem, WorkerCount)
	for i := 0; i < WorkerCount; i++ {
		mch := make(chan *p2p.TxMsgItem)
		an.txMsgChans[i] = &mch
		ch := make(chan struct{})
		workerEnd[i] = &ch
		go func(pMsgCh *chan *p2p.TxMsgItem, pEndCh *chan struct{}) {
			for {
				select {
				case item := <-(*pMsgCh):
					if err := an.addTx(item.Message.TxType, item.Message.Tx, item.Message.Sigs); err != nil {
						rlog.Println("TransactionError", chain.HashTransactionByType(an.cs.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String(), err.Error())
						if err != txpool.ErrPastSeq && err != txpool.ErrTooFarSeq {
							(*item.ErrCh) <- err
						} else {
							(*item.ErrCh) <- nil
						}
						break
					}
					rlog.Println("TransactionAppended", chain.HashTransactionByType(an.cs.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String())
					(*item.ErrCh) <- nil
				case <-(*pEndCh):
					return
				}
			}
		}(&mch, &ch)
	}

	blockGenTimer := time.NewTimer(500 * time.Millisecond)
	for !an.isClose {
		select {
		case <-blockGenTimer.C:
			an.genBlock()
			blockGenTimer.Reset(50 * time.Millisecond)
		}
	}
	for i := 0; i < WorkerCount; i++ {
		(*workerEnd[i]) <- struct{}{}
	}
}

// AddTx adds tx to txpool that only have valid signatures
func (an *AuthorityNode) AddTx(tx types.Transaction, sigs []common.Signature) error {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	if err := an.addTx(t, tx, sigs); err != nil {
		return err
	}
	return nil
}

func (an *AuthorityNode) addTx(t uint16, tx types.Transaction, sigs []common.Signature) error {
	if an.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	TxHash := chain.HashTransactionByType(an.cs.cn.Provider().ChainID(), t, tx)

	ctx := an.cs.ct.NewContext()
	if an.txpool.IsExist(TxHash) {
		return txpool.ErrExistTransaction
	}
	if atx, is := tx.(chain.AccountTransaction); is {
		seq := ctx.Seq(atx.From())
		if atx.Seq() <= seq {
			return txpool.ErrPastSeq
		} else if atx.Seq() > seq+100 {
			return txpool.ErrTooFarSeq
		}
	}
	signers := make([]common.PublicHash, 0, len(sigs))
	for _, sig := range sigs {
		pubkey, err := common.RecoverPubkey(TxHash, sig)
		if err != nil {
			return err
		}
		signers = append(signers, common.NewPublicHash(pubkey))
	}
	pid := uint8(t >> 8)
	p, err := an.cs.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := an.txpool.Push(an.cs.cn.Provider().ChainID(), t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	an.txQ.Push(string(TxHash[:]), &p2p.TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

// OnConnected is called after a new  peer is connected
func (an *AuthorityNode) OnConnected(p peer.Peer) {
	an.statusLock.Lock()
	an.statusMap[p.ID()] = &p2p.Status{}
	an.statusLock.Unlock()
}

// OnDisconnected is called when the  peer is disconnected
func (an *AuthorityNode) OnDisconnected(p peer.Peer) {
	an.statusLock.Lock()
	delete(an.statusMap, p.ID())
	an.statusLock.Unlock()
}

// OnRecv called when message received
func (an *AuthorityNode) OnRecv(p peer.Peer, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
		cp := an.cs.cn.Provider()
		Height := cp.Height()
		if msg.Height > Height {
			return nil
		}
		list := make([]*types.Block, 0, 10)
		for i := uint32(0); i < uint32(msg.Count); i++ {
			if msg.Height+i > Height {
				break
			}
			b, err := cp.Block(msg.Height + i)
			if err != nil {
				return err
			}
			list = append(list, b)
		}
		sm := &p2p.BlockMessage{
			Blocks: list,
		}
		if err := an.ms.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		an.statusLock.Lock()
		if status, has := an.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		an.statusLock.Unlock()

		Height := an.cs.cn.Provider().Height()
		if Height < msg.Height {
			panic(chain.ErrFoundForkedBlock)
		} else {
			h, err := an.cs.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				an.ms.RemovePeer(p.ID())
			}
		}
	case *p2p.BlockMessage:
		if len(msg.Blocks) > 0 {
			an.statusLock.Lock()
			if status, has := an.statusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			an.statusLock.Unlock()
		}
	case *p2p.TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&an.txMsgIdx, 1) % uint64(len(an.txMsgChans))
		(*an.txMsgChans[idx]) <- &p2p.TxMsgItem{
			Message: msg,
			PeerID:  p.ID(),
			ErrCh:   &errCh,
		}
		err := <-errCh
		if err != p2p.ErrInvalidUTXO && err != txpool.ErrExistTransaction {
			return err
		}
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (an *AuthorityNode) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(an.cs.cn.Provider().ChainID(), t, tx)
		an.txpool.Remove(TxHash, tx)
		an.txQ.Remove(string(TxHash[:]))
	}
}

func (an *AuthorityNode) genBlock() error {
	cp := an.cs.cn.Provider()

	Timestamp := cp.LastTimestamp() + uint64(500*time.Millisecond)
	ctx := an.cs.ct.NewContext()
	bc := chain.NewBlockCreator(an.cs.cn, ctx, an.Config.Authority, nil)
	if err := bc.Init(); err != nil {
		return err
	}

	timer := time.NewTimer(200 * time.Millisecond)

	an.txpool.Lock() // Prevent delaying from TxPool.Push
	Count := 0
TxLoop:
	for {
		select {
		case <-timer.C:
			break TxLoop
		default:
			sn := ctx.Snapshot()
			item := an.txpool.UnsafePop(ctx)
			ctx.Revert(sn)
			if item == nil {
				break TxLoop
			}
			if err := bc.UnsafeAddTx(an.Config.Authority, item.TxType, item.TxHash, item.Transaction, item.Signatures, item.Signers); err != nil {
				rlog.Println(err)
				continue
			}
			Count++
			if Count > an.Config.MaxTransactionsPerBlock {
				break TxLoop
			}
		}
	}
	an.txpool.Unlock() // Prevent delaying from TxPool.Push

	b, err := bc.Finalize(Timestamp)
	if err != nil {
		return err
	}
	if sig, err := an.key.Sign(encoding.Hash(b.Header)); err != nil {
		return err
	} else {
		b.Signatures = append(b.Signatures, sig)
	}
	if err := an.cs.ct.ConnectBlockWithContext(b, ctx); err != nil {
		return err
	}
	rlog.Println("Authority", "BlockConnected", b.Header.Height, len(b.Transactions))

	an.broadcastStatus()

	PastTime := uint64(time.Now().UnixNano()) - an.firstTime
	ExpectedTime := uint64(b.Header.Height-an.firstHeight) * uint64(500*time.Millisecond)

	if PastTime < ExpectedTime {
		diff := time.Duration(ExpectedTime - PastTime)
		if diff > 500*time.Millisecond {
			diff = 500 * time.Millisecond
		}
		time.Sleep(diff)
	}
	return nil
}
