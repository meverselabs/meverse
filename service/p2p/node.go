package p2p

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
	"github.com/fletaio/fleta/service/p2p/peer"
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
	statusMap    map[string]*Status
	txpool       *txpool.TransactionPool
	txQ          *queue.ExpireQueue
	txWaitQ      *queue.LinkedQueue
	txSendQ      *queue.Queue
	recvChan     chan *RecvMessageItem
	sendChan     chan *SendMessageItem
	singleCache  gcache.Cache
	batchCache   gcache.Cache
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
		txWaitQ:      queue.NewLinkedQueue(),
		txSendQ:      queue.NewQueue(),
		recvChan:     make(chan *RecvMessageItem, 1000),
		sendChan:     make(chan *SendMessageItem, 1000),
		singleCache:  gcache.New(500).LRU().Build(),
		batchCache:   gcache.New(500).LRU().Build(),
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
	fc.Register(StatusMessageType, &StatusMessage{})
	fc.Register(RequestMessageType, &RequestMessage{})
	fc.Register(BlockMessageType, &BlockMessage{})
	fc.Register(TransactionMessageType, &TransactionMessage{})
	fc.Register(PeerListMessageType, &PeerListMessage{})
	fc.Register(RequestPeerListMessageType, &RequestPeerListMessage{})
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
	item := Item.(*TxMsgItem)
	cp := nd.cn.Provider()
	if atx, is := item.Tx.(chain.AccountTransaction); is {
		seq := cp.Seq(atx.From())
		if atx.Seq() <= seq {
			return
		}
	}
	nd.txSendQ.Push(item)
	if IsLast {
		nd.txpool.Remove(item.TxHash, item.Tx)
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

	WorkerCount := runtime.NumCPU() / 2
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	for i := 0; i < WorkerCount; i++ {
		go func() {
			for !nd.isClose {
				Count := 0
				ctw := nd.cn.Provider().NewLoaderWrapper(1)
				for !nd.isClose {
					v := nd.txWaitQ.Pop()
					if v == nil {
						break
					}

					Count++
					if Count > 500 {
						break
					}
					item := v.(*TxMsgItem)
					if err := nd.addTx(ctw, item.TxHash, item.Type, item.Tx, item.Sigs); err != nil {
						if err != ErrInvalidUTXO && err != txpool.ErrExistTransaction && err != txpool.ErrTooFarSeq && err != txpool.ErrPastSeq {
							rlog.Println("TransactionError", item.TxHash.String(), err.Error())
							if len(item.PeerID) > 0 {
								nd.ms.RemovePeer(item.PeerID)
							}
						}
						continue
					}
					rlog.Println("TransactionAppended", item.TxHash.String())

					nd.txSendQ.Push(item)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	go func() {
		for !nd.isClose {
			if nd.ms.HasPeer() {
				msg := &TransactionMessage{
					Types:      []uint16{},
					Txs:        []types.Transaction{},
					Signatures: [][]common.Signature{},
				}
				for {
					v := nd.txSendQ.Pop()
					if v == nil {
						break
					}
					m := v.(*TxMsgItem)
					msg.Types = append(msg.Types, m.Type)
					msg.Txs = append(msg.Txs, m.Tx)
					msg.Signatures = append(msg.Signatures, m.Sigs)
					if len(msg.Types) >= 5000 {
						break
					}
				}
				if len(msg.Types) > 0 {
					//log.Println("Send.TransactionMessage", len(msg.Types))
					nd.broadcastMessage(1, msg)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for i := 0; i < 2; i++ {
		go func() {
			for item := range nd.recvChan {
				if nd.isClose {
					break
				}
				m, err := PacketToMessage(item.Packet)
				if err != nil {
					log.Println("PacketToMessage", err)
					nd.ms.RemovePeer(item.PeerID)
					break
				}
				if err := nd.handlePeerMessage(item.PeerID, m); err != nil {
					log.Println("handlePeerMessage", err)
					nd.ms.RemovePeer(item.PeerID)
					break
				}
			}
		}()
	}

	for i := 0; i < 2; i++ {
		go func() {
			for item := range nd.sendChan {
				if nd.isClose {
					break
				}
				var EmptyHash common.PublicHash
				if bytes.Equal(item.Target[:], EmptyHash[:]) {
					nd.ms.BroadcastPacket(item.Packet)
				} else {
					if item.Except {
						nd.ms.ExceptCast(string(item.Target[:]), item.Packet)
					} else {
						nd.ms.SendTo(item.Target, item.Packet)
					}
				}
			}
		}()
	}

	go func() {
		for !nd.isClose {
			nd.tryRequestBlocks()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for !nd.isClose {
		nd.Lock()
		hasItem := false
		TargetHeight := uint64(nd.cn.Provider().Height() + 1)
		Count := 0
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
			Count++
			if Count > 10 {
				break
			}
			item = nd.blockQ.PopUntil(TargetHeight)
			hasItem = true
		}
		nd.Unlock()

		if hasItem {
			nd.broadcastStatus()
			nd.tryRequestBlocks()
		}

		if hasItem {
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(200 * time.Millisecond)
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

	cp := nd.cn.Provider()
	height, lastHash := cp.LastStatus()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	p.SendPacket(MessageToPacket(nm))
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
func (nd *Node) OnRecv(p peer.Peer, bs []byte) error {
	nd.recvChan <- &RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	return nil
}

func (nd *Node) handlePeerMessage(ID string, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(ID))

	switch msg := m.(type) {
	case *RequestMessage:
		nd.statusLock.Lock()
		status, has := nd.statusMap[ID]
		nd.statusLock.Unlock()
		if has {
			if msg.Height < status.Height {
				if msg.Height+uint32(msg.Count) <= status.Height {
					return nil
				}
				msg.Height = status.Height
			}
		}

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
		bs, err := BlockPacketWithCache(msg, nd.cn.Provider(), nd.batchCache, nd.singleCache)
		if err != nil {
			return err
		}
		nd.sendMessagePacket(0, SenderPublicHash, bs)
		return nil
	case *StatusMessage:
		nd.statusLock.Lock()
		if status, has := nd.statusMap[ID]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		nd.statusLock.Unlock()

		Height := nd.cn.Provider().Height()
		if Height < msg.Height {
			enableCount := 0
			for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
				if !nd.requestTimer.Exist(i) {
					enableCount++
				}
			}
			if Height%10 == 0 && enableCount == 10 {
				nd.sendRequestBlockTo(SenderPublicHash, Height+1, 10)
			} else {
				for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
					if !nd.requestTimer.Exist(i) {
						nd.sendRequestBlockTo(SenderPublicHash, i, 1)
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
				rlog.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				nd.ms.RemovePeer(ID)
			}
		}
		return nil
	case *BlockMessage:
		for _, b := range msg.Blocks {
			if err := nd.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					//TODO : critical error signal
					nd.ms.RemovePeer(ID)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			nd.statusLock.Lock()
			if status, has := nd.statusMap[ID]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			nd.statusLock.Unlock()
		}
		return nil
	case *TransactionMessage:
		//log.Println("Recv.TransactionMessage", nd.txWaitQ.Size(), nd.txpool.Size())
		if nd.txWaitQ.Size() > 200000 {
			return txpool.ErrTransactionPoolOverflowed
		}
		if len(msg.Types) > 5000 {
			return ErrTooManyTrasactionInMessage
		}
		ChainID := nd.cn.Provider().ChainID()
		for i, t := range msg.Types {
			tx := msg.Txs[i]
			sigs := msg.Signatures[i]
			TxHash := chain.HashTransactionByType(ChainID, t, tx)
			if !nd.txpool.IsExist(TxHash) {
				nd.txWaitQ.Push(TxHash, &TxMsgItem{
					TxHash: TxHash,
					Type:   t,
					Tx:     tx,
					Sigs:   sigs,
					PeerID: ID,
				})
			}
		}
		return nil
	case *PeerListMessage:
		nd.ms.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *RequestPeerListMessage:
		nd.ms.SendPeerList(ID)
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
	TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), t, tx)
	if !nd.txpool.IsExist(TxHash) {
		nd.txWaitQ.Push(TxHash, &TxMsgItem{
			TxHash: TxHash,
			Type:   t,
			Tx:     tx,
			Sigs:   sigs,
		})
	}
	return nil
}

func (nd *Node) addTx(ctw types.LoaderWrapper, TxHash hash.Hash256, t uint16, tx types.Transaction, sigs []common.Signature) error {
	if nd.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}
	cp := nd.cn.Provider()
	if nd.txpool.IsExist(TxHash) {
		return txpool.ErrExistTransaction
	}
	if atx, is := tx.(chain.AccountTransaction); is {
		seq := cp.Seq(atx.From())
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
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := nd.txpool.Push(t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	nd.txQ.Push(string(TxHash[:]), &TxMsgItem{
		Type: t,
		Tx:   tx,
		Sigs: sigs,
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
		var MaxHeight uint32
		var maxPubHash string
		nd.statusLock.Lock()
		for pubhash, status := range nd.statusMap {
			if MaxHeight < status.Height {
				maxPubHash = pubhash
				MaxHeight = status.Height
			}
			if BaseHeight+10 <= status.Height {
				selectedPubHash = pubhash
				LimitHeight = status.Height
				break
			}
		}
		nd.statusLock.Unlock()

		if LimitHeight == 0 {
			selectedPubHash = maxPubHash
			LimitHeight = MaxHeight
		}

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
