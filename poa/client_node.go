package poa

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

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
	"github.com/fletaio/fleta/service/p2p/peer"
)

// ClientNode procudes a block by the consensus
type ClientNode struct {
	sync.Mutex
	cs               *Consensus
	ms               *ClientNodeMesh
	nm               *p2p.NodeMesh
	key              key.Key
	ndkey            key.Key
	myPublicHash     common.PublicHash
	statusLock       sync.Mutex
	genLock          sync.Mutex
	txMsgChans       []*chan *p2p.TxMsgItem
	txMsgIdx         uint64
	statusMap        map[string]*p2p.Status
	anStatusMap      map[string]*p2p.Status
	requestTimer     *p2p.RequestTimer
	requestNodeTimer *p2p.RequestTimer
	requestLock      sync.RWMutex
	blockQ           *queue.SortedQueue
	txpool           *txpool.TransactionPool
	txQ              *queue.ExpireQueue
	isRunning        bool
	closeLock        sync.RWMutex
	isClose          bool
}

// NewClientNode returns a ClientNode
func NewClientNode(key key.Key, ndkey key.Key, NetAddressMap map[common.PublicHash]string, SeedNodeMap map[common.PublicHash]string, cs *Consensus, peerStorePath string) *ClientNode {
	ci := &ClientNode{
		cs:               cs,
		key:              key,
		ndkey:            ndkey,
		myPublicHash:     common.NewPublicHash(ndkey.PublicKey()),
		statusMap:        map[string]*p2p.Status{},
		anStatusMap:      map[string]*p2p.Status{},
		requestTimer:     p2p.NewRequestTimer(nil),
		requestNodeTimer: p2p.NewRequestTimer(nil),
		blockQ:           queue.NewSortedQueue(),
		txpool:           txpool.NewTransactionPool(),
		txQ:              queue.NewExpireQueue(),
	}
	ci.ms = NewClientNodeMesh(key, NetAddressMap, ci)
	ci.nm = p2p.NewNodeMesh(ci.cs.cn.Provider().ChainID(), ndkey, SeedNodeMap, ci, peerStorePath)
	ci.txQ.AddGroup(60 * time.Second)
	ci.txQ.AddGroup(600 * time.Second)
	ci.txQ.AddGroup(3600 * time.Second)
	ci.txQ.AddHandler(ci)
	return ci
}

// Close terminates the Client
func (ci *ClientNode) Close() {
	ci.closeLock.Lock()
	defer ci.closeLock.Unlock()

	ci.Lock()
	defer ci.Unlock()

	ci.isClose = true
	ci.cs.cn.Close()
}

// Init initializes Client
func (ci *ClientNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})
	fc.Register(types.DefineHashedType("p2p.TransactionMessage"), &p2p.TransactionMessage{})
	fc.Register(types.DefineHashedType("p2p.PeerListMessage"), &p2p.PeerListMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestPeerListMessage"), &p2p.RequestPeerListMessage{})
	return nil
}

// Run runs the Client
func (ci *ClientNode) Run(BindAddress string) {
	ci.Lock()
	if ci.isRunning {
		ci.Unlock()
		return
	}
	ci.isRunning = true
	ci.Unlock()

	go ci.ms.Run()
	go ci.nm.Run(BindAddress)
	go ci.requestTimer.Run()
	go ci.requestNodeTimer.Run()

	WorkerCount := runtime.NumCPU() - 1
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	workerEnd := make([]*chan struct{}, WorkerCount)
	ci.txMsgChans = make([]*chan *p2p.TxMsgItem, WorkerCount)
	for i := 0; i < WorkerCount; i++ {
		mch := make(chan *p2p.TxMsgItem)
		ci.txMsgChans[i] = &mch
		ch := make(chan struct{})
		workerEnd[i] = &ch
		go func(pMsgCh *chan *p2p.TxMsgItem, pEndCh *chan struct{}) {
			for {
				select {
				case item := <-(*pMsgCh):
					if err := ci.addTx(item.Message.TxType, item.Message.Tx, item.Message.Sigs); err != nil {
						rlog.Println("TransactionError", chain.HashTransactionByType(ci.cs.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String(), err.Error())
						if err != txpool.ErrPastSeq && err != txpool.ErrTooFarSeq {
							(*item.ErrCh) <- err
						} else {
							(*item.ErrCh) <- nil
						}
						break
					}
					rlog.Println("TransactionAppended", chain.HashTransactionByType(ci.cs.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String())
					(*item.ErrCh) <- nil

					ci.ms.BroadcastMessage(item.Message)
				case <-(*pEndCh):
					return
				}
			}
		}(&mch, &ch)
	}

	blockTimer := time.NewTimer(time.Millisecond)
	blockRequestTimer := time.NewTimer(time.Millisecond)
	for !ci.isClose {
		select {
		case <-blockTimer.C:
			ci.Lock()
			hasItem := false
			TargetHeight := uint64(ci.cs.cn.Provider().Height() + 1)
			Count := 0
			item := ci.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := ci.cs.cn.ConnectBlock(b); err != nil {
					break
				}
				ci.cleanPool(b)
				rlog.Println("Client", "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))
				TargetHeight++
				Count++
				if Count > 100 {
					break
				}
				item = ci.blockQ.PopUntil(TargetHeight)
				hasItem = true
			}
			ci.Unlock()

			if hasItem {
				ci.broadcastStatus()
				ci.tryRequestBlocks()
			}

			blockTimer.Reset(50 * time.Millisecond)
		case <-blockRequestTimer.C:
			ci.tryRequestBlocks()
			ci.tryRequestNext()
			blockRequestTimer.Reset(500 * time.Millisecond)
		}
	}
	for i := 0; i < WorkerCount; i++ {
		(*workerEnd[i]) <- struct{}{}
	}
}

// AddTx adds tx to txpool that only have valid signatures
func (ci *ClientNode) AddTx(tx types.Transaction, sigs []common.Signature) error {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	if err := ci.addTx(t, tx, sigs); err != nil {
		return err
	}
	ci.ms.BroadcastMessage(&p2p.TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

func (ci *ClientNode) addTx(t uint16, tx types.Transaction, sigs []common.Signature) error {
	if ci.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	TxHash := chain.HashTransactionByType(ci.cs.cn.Provider().ChainID(), t, tx)

	ctx := ci.cs.ct.NewContext()
	if ci.txpool.IsExist(TxHash) {
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
	p, err := ci.cs.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := ci.txpool.Push(ci.cs.cn.Provider().ChainID(), t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	ci.txQ.Push(string(TxHash[:]), &p2p.TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

// OnTimerExpired called when rquest expired
func (ci *ClientNode) OnTimerExpired(height uint32, value string) {
	go ci.tryRequestBlocks()
}

// OnItemExpired is called when the item is expired
func (ci *ClientNode) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	msg := Item.(*p2p.TransactionMessage)
	ci.ms.BroadcastMessage(msg)
	if IsLast {
		var TxHash hash.Hash256
		copy(TxHash[:], []byte(Key))
		ci.txpool.Remove(TxHash, msg.Tx)
	}
}

// OnAuthorityConnected is called after a new Authority peer is connected
func (ci *ClientNode) OnAuthorityConnected(p peer.Peer) {
	ci.statusLock.Lock()
	ci.anStatusMap[p.ID()] = &p2p.Status{}
	ci.statusLock.Unlock()

	cp := ci.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	p.Send(nm)
}

// OnAuthorityDisconnected is called when the Authority peer is disconnected
func (ci *ClientNode) OnAuthorityDisconnected(p peer.Peer) {
	ci.statusLock.Lock()
	delete(ci.anStatusMap, p.ID())
	ci.statusLock.Unlock()
	ci.requestTimer.RemovesByValue(p.ID())
	go ci.tryRequestNext()
}

// OnConnected is called after a new  peer is connected
func (ci *ClientNode) OnConnected(p peer.Peer) {
	ci.statusLock.Lock()
	ci.statusMap[p.ID()] = &p2p.Status{}
	ci.statusLock.Unlock()
}

// OnDisconnected is called when the  peer is disconnected
func (ci *ClientNode) OnDisconnected(p peer.Peer) {
	ci.statusLock.Lock()
	delete(ci.statusMap, p.ID())
	ci.statusLock.Unlock()
	ci.requestNodeTimer.RemovesByValue(p.ID())
	go ci.tryRequestBlocks()
}

// OnRecv called when message received
func (ci *ClientNode) OnRecv(p peer.Peer, m interface{}) error {
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
		cp := ci.cs.cn.Provider()
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
		if err := ci.nm.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		ci.statusLock.Lock()
		if status, has := ci.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		ci.statusLock.Unlock()

		Height := ci.cs.cn.Provider().Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				if BaseHeight > msg.Height {
					break
				}
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
					if !ci.requestNodeTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					ci.sendRequestBlockToNode(SenderPublicHash, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
						if !ci.requestNodeTimer.Exist(i) {
							ci.sendRequestBlockToNode(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := ci.cs.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				ci.nm.RemovePeer(p.ID())
			}
		}
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := ci.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					ci.nm.RemovePeer(p.ID())
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			ci.statusLock.Lock()
			if status, has := ci.statusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			ci.statusLock.Unlock()
		}
	case *p2p.TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&ci.txMsgIdx, 1) % uint64(len(ci.txMsgChans))
		(*ci.txMsgChans[idx]) <- &p2p.TxMsgItem{
			Message: msg,
			PeerID:  p.ID(),
			ErrCh:   &errCh,
		}
		err := <-errCh
		if err != p2p.ErrInvalidUTXO && err != txpool.ErrExistTransaction {
			return err
		}
		return nil
	case *p2p.PeerListMessage:
		ci.nm.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *p2p.RequestPeerListMessage:
		ci.nm.SendPeerList(p.ID())
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (ci *ClientNode) tryRequestBlocks() {
	ci.requestLock.Lock()
	defer ci.requestLock.Unlock()

	Height := ci.cs.cn.Provider().Height()
	for q := uint32(0); q < 10; q++ {
		BaseHeight := Height + q*10

		var LimitHeight uint32
		var selectedPubHash string
		ci.statusLock.Lock()
		for pubhash, status := range ci.statusMap {
			if BaseHeight+10 <= status.Height {
				selectedPubHash = pubhash
				LimitHeight = status.Height
				break
			}
		}
		if len(selectedPubHash) == 0 {
			for pubhash, status := range ci.statusMap {
				if BaseHeight <= status.Height {
					selectedPubHash = pubhash
					LimitHeight = status.Height
					break
				}
			}
		}
		ci.statusLock.Unlock()

		if len(selectedPubHash) == 0 {
			break
		}
		enableCount := 0
		for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
			if !ci.requestNodeTimer.Exist(i) {
				enableCount++
			}
		}

		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(selectedPubHash))
		if enableCount == 10 {
			ci.sendRequestBlockToNode(TargetPublicHash, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
				if !ci.requestNodeTimer.Exist(i) {
					ci.sendRequestBlockToNode(TargetPublicHash, i, 1)
				}
			}
		}
	}
}

func (ci *ClientNode) onRecv(p peer.Peer, m interface{}) error {
	if err := ci.handleMessage(p, m, 0); err != nil {
		//rlog.Println(err)
		return nil
	}
	return nil
}

func (ci *ClientNode) handleMessage(p peer.Peer, m interface{}, RetryCount int) error {
	cp := ci.cs.cn.Provider()

	switch msg := m.(type) {
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := ci.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					panic(err)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			ci.statusLock.Lock()
			if status, has := ci.anStatusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			ci.statusLock.Unlock()

			ci.tryRequestNext()
		}
		return nil
	case *p2p.StatusMessage:
		ci.statusLock.Lock()
		if status, has := ci.anStatusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		ci.statusLock.Unlock()

		TargetHeight := cp.Height() + 1
		for TargetHeight <= msg.Height {
			if !ci.requestTimer.Exist(TargetHeight) {
				if ci.blockQ.Find(uint64(TargetHeight)) == nil {
					sm := &p2p.RequestMessage{
						Height: TargetHeight,
					}
					if err := p.Send(sm); err != nil {
						return err
					}
					ci.requestTimer.Add(TargetHeight, 2*time.Second, p.ID())
				}
			}
			TargetHeight++
		}
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
}

func (ci *ClientNode) addBlock(b *types.Block) error {
	cp := ci.cs.cn.Provider()
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
		if item := ci.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
			}
		}
	}
	return nil
}

func (ci *ClientNode) tryRequestNext() {
	ci.requestLock.Lock()
	defer ci.requestLock.Unlock()

	TargetHeight := ci.cs.cn.Provider().Height() + 1
	if !ci.requestTimer.Exist(TargetHeight) {
		if ci.blockQ.Find(uint64(TargetHeight)) == nil {
			ci.statusLock.Lock()
			var TargetPubHash string
			for pubhash, status := range ci.anStatusMap {
				if TargetHeight <= status.Height {
					TargetPubHash = pubhash
					break
				}
			}
			ci.statusLock.Unlock()

			if len(TargetPubHash) > 0 {
				ci.sendRequestBlockTo(TargetPubHash, TargetHeight, 1)
			}
		}
	}
}

func (ci *ClientNode) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(ci.cs.cn.Provider().ChainID(), t, tx)
		ci.txpool.Remove(TxHash, tx)
		ci.txQ.Remove(string(TxHash[:]))
	}
}
