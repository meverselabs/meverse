package p2p

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/service/p2p/peer"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Node receives a block by the consensus
type Node struct {
	sync.Mutex
	key          key.Key
	ms           *NodeMesh
	cn           *chain.Chain
	statusLock   sync.Mutex
	myPublicHash common.PublicHash
	requestTimer *RequestTimer
	requestLock  sync.RWMutex
	blockQ       *queue.SortedQueue
	txMsgChans   []*chan *TxMsgItem
	txMsgIdx     uint64
	statusMap    map[string]*Status
	txpool       *txpool.TransactionPool
	txQ          *queue.ExpireQueue
	isRunning    bool
	closeLock    sync.RWMutex
	isClose      bool
}

// NewNode returns a Node
func NewNode(key key.Key, SeedNodeMap map[common.PublicHash]string, cn *chain.Chain, peerStorePath string) *Node {
	nd := &Node{
		key:          key,
		cn:           cn,
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		blockQ:       queue.NewSortedQueue(),
		statusMap:    map[string]*Status{},
		txpool:       txpool.NewTransactionPool(),
		txQ:          queue.NewExpireQueue(),
	}
	nd.ms = NewNodeMesh(cn.Provider().ChainID(), key, SeedNodeMap, nd, peerStorePath)
	nd.requestTimer = NewRequestTimer(nd)
	nd.txQ.AddGroup(60 * time.Second)
	nd.txQ.AddGroup(600 * time.Second)
	nd.txQ.AddGroup(3600 * time.Second)
	nd.txQ.AddHandler(nd)
	rlog.SetRLogAddress("nd:" + nd.myPublicHash.String())
	return nd
}

// Init initializes node
func (nd *Node) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &RequestMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.TransactionMessage"), &TransactionMessage{})
	fc.Register(types.DefineHashedType("p2p.PeerListMessage"), &PeerListMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestPeerListMessage"), &RequestPeerListMessage{})
	return nil
}

// Close terminates the node
func (nd *Node) Close() {
	nd.closeLock.Lock()
	defer nd.closeLock.Unlock()

	nd.Lock()
	defer nd.Unlock()

	nd.isClose = true
	nd.cn.Close()
}

// OnItemExpired is called when the item is expired
func (nd *Node) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	msg := Item.(*TransactionMessage)
	nd.ms.ExceptCastLimit("", msg, 7)
	if IsLast {
		var TxHash hash.Hash256
		copy(TxHash[:], []byte(Key))
		nd.txpool.Remove(TxHash, msg.Tx)
	}
}

// Run starts the node
func (nd *Node) Run(BindAddress string) {
	nd.Lock()
	if nd.isRunning {
		nd.Unlock()
		return
	}
	nd.isRunning = true
	nd.Unlock()

	go nd.ms.Run(BindAddress)
	go nd.requestTimer.Run()

	WorkerCount := runtime.NumCPU() - 1
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	workerEnd := make([]*chan struct{}, WorkerCount)
	nd.txMsgChans = make([]*chan *TxMsgItem, WorkerCount)
	for i := 0; i < WorkerCount; i++ {
		mch := make(chan *TxMsgItem)
		nd.txMsgChans[i] = &mch
		ch := make(chan struct{})
		workerEnd[i] = &ch
		go func(pMsgCh *chan *TxMsgItem, pEndCh *chan struct{}) {
			for {
				select {
				case item := <-(*pMsgCh):
					if err := nd.addTx(item.Message.TxType, item.Message.Tx, item.Message.Sigs); err != nil {
						rlog.Println("TransactionError", chain.HashTransactionByType(nd.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String(), err.Error())
						if err != txpool.ErrPastSeq && err != txpool.ErrTooFarSeq {
							(*item.ErrCh) <- err
						} else {
							(*item.ErrCh) <- nil
						}
						break
					}
					rlog.Println("TransactionAppended", chain.HashTransactionByType(nd.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String())
					(*item.ErrCh) <- nil

					nd.ms.ExceptCastLimit(item.PeerID, item.Message, 7)
				case <-(*pEndCh):
					return
				}
			}
		}(&mch, &ch)
	}

	blockTimer := time.NewTimer(time.Millisecond)
	blockRequestTimer := time.NewTimer(time.Millisecond)
	for !nd.isClose {
		select {
		case <-blockTimer.C:
			nd.Lock()
			hasItem := false
			TargetHeight := uint64(nd.cn.Provider().Height() + 1)
			item := nd.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := nd.cn.ConnectBlock(b); err != nil {
					rlog.Println(err)
					panic(err)
					break
				}
				rlog.Println("Node", nd.myPublicHash.String(), nd.cn.Provider().Height(), "BlockConnected", b.Header.Generator.String(), b.Header.Height)
				TargetHeight++
				item = nd.blockQ.PopUntil(TargetHeight)
				hasItem = true
			}
			nd.Unlock()

			if hasItem {
				nd.broadcastStatus()
				nd.tryRequestBlocks()
			}

			blockTimer.Reset(50 * time.Millisecond)
		case <-blockRequestTimer.C:
			nd.tryRequestBlocks()
			blockRequestTimer.Reset(500 * time.Millisecond)
		}
	}
}

// OnTimerExpired called when rquest expired
func (nd *Node) OnTimerExpired(height uint32, value string) {
	nd.tryRequestBlocks()
}

// OnConnected called when peer connected
func (nd *Node) OnConnected(p peer.Peer) {
	nd.statusLock.Lock()
	nd.statusMap[p.ID()] = &Status{}
	nd.statusLock.Unlock()

	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))
	nd.sendStatusTo(SenderPublicHash)
}

// OnDisconnected called when peer disconnected
func (nd *Node) OnDisconnected(p peer.Peer) {
	nd.statusLock.Lock()
	delete(nd.statusMap, p.ID())
	nd.statusLock.Unlock()

	nd.requestTimer.RemovesByValue(p.ID())
	go nd.tryRequestBlocks()
}

// OnRecv called when message received
func (nd *Node) OnRecv(p peer.Peer, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))

	switch msg := m.(type) {
	case *RequestMessage:
		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
		Height := nd.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		list := make([]*types.Block, 0, 10)
		for i := uint32(0); i < uint32(msg.Count); i++ {
			if msg.Height+i > Height {
				break
			}
			b, err := nd.cn.Provider().Block(msg.Height + i)
			if err != nil {
				return err
			}
			list = append(list, b)
		}
		sm := &BlockMessage{
			Blocks: list,
		}
		if err := nd.ms.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
		return nil
	case *StatusMessage:
		nd.statusLock.Lock()
		if status, has := nd.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		nd.statusLock.Unlock()

		Height := nd.cn.Provider().Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				if BaseHeight > msg.Height {
					break
				}
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
					if !nd.requestTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					nd.sendRequestBlockTo(SenderPublicHash, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
						if !nd.requestTimer.Exist(i) {
							nd.sendRequestBlockTo(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := nd.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(chain.ErrFoundForkedBlock, p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				nd.ms.RemovePeer(p.ID())
			}
		}
		return nil
	case *BlockMessage:
		for _, b := range msg.Blocks {
			if err := nd.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					//TODO : critical error signal
					nd.ms.RemovePeer(p.ID())
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			nd.statusLock.Lock()
			if status, has := nd.statusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			nd.statusLock.Unlock()
		}
		return nil
	case *TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&nd.txMsgIdx, 1) % uint64(len(nd.txMsgChans))
		(*nd.txMsgChans[idx]) <- &TxMsgItem{
			Message: msg,
			PeerID:  p.ID(),
			ErrCh:   &errCh,
		}
		err := <-errCh
		if err != ErrInvalidUTXO && err != txpool.ErrExistTransaction {
			return err
		}
		return nil
	case *PeerListMessage:
		nd.ms.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *RequestPeerListMessage:
		nd.ms.SendPeerList(p.ID())
		return nil
	default:
		panic(ErrUnknownMessage) //TEMP
		return ErrUnknownMessage
	}
	return nil
}

func (nd *Node) addBlock(b *types.Block) error {
	cp := nd.cn.Provider()
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
		if item := nd.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
			}
		}
	}
	return nil
}

// AddTx adds tx to txpool that only have valid signatures
func (nd *Node) AddTx(tx types.Transaction, sigs []common.Signature) error {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	if err := nd.addTx(t, tx, sigs); err != nil {
		return err
	}
	nd.ms.ExceptCastLimit("", &TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	}, 7)
	return nil
}

func (nd *Node) addTx(t uint16, tx types.Transaction, sigs []common.Signature) error {
	if nd.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), t, tx)

	ctx := nd.cn.NewContext()
	if nd.txpool.IsExist(TxHash) {
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
	p, err := nd.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := nd.txpool.Push(nd.cn.Provider().ChainID(), t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	nd.txQ.Push(string(TxHash[:]), &TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

func (nd *Node) tryRequestBlocks() {
	nd.requestLock.Lock()
	defer nd.requestLock.Unlock()

	Height := nd.cn.Provider().Height()
	for q := uint32(0); q < 10; q++ {
		BaseHeight := Height + q*10

		var LimitHeight uint32
		var selectedPubHash string
		nd.statusLock.Lock()
		for pubhash, status := range nd.statusMap {
			if BaseHeight+10 <= status.Height {
				selectedPubHash = pubhash
				LimitHeight = status.Height
				break
			}
		}
		if len(selectedPubHash) == 0 {
			for pubhash, status := range nd.statusMap {
				if BaseHeight <= status.Height {
					selectedPubHash = pubhash
					LimitHeight = status.Height
					break
				}
			}
		}
		nd.statusLock.Unlock()

		if len(selectedPubHash) == 0 {
			break
		}
		enableCount := 0
		for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
			if !nd.requestTimer.Exist(i) {
				enableCount++
			}
		}

		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(selectedPubHash))
		if enableCount == 10 {
			nd.sendRequestBlockTo(TargetPublicHash, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
				if !nd.requestTimer.Exist(i) {
					nd.sendRequestBlockTo(TargetPublicHash, i, 1)
				}
			}
		}
	}
}

func (nd *Node) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), t, tx)
		nd.txpool.Remove(TxHash, tx)
		nd.txQ.Remove(string(TxHash[:]))
	}
}

// TxMsgItem used to store transaction message
type TxMsgItem struct {
	Message *TransactionMessage
	PeerID  string
	ErrCh   *chan error
}
