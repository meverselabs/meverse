package p2p

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"runtime"
	"strconv"
	"strings"
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
	"github.com/meverselabs/meverse/core/txpool"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p/peer"
)

// Node receives a block by the consensus
type Node struct {
	sync.Mutex
	ChainID      *big.Int
	key          key.Key
	ms           *NodeMesh
	cn           *chain.Chain
	statusLock   sync.Mutex
	myPublicKey  common.PublicKey
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
func NewNode(ChainID *big.Int, key key.Key, SeedNodeMap map[common.PublicKey]string, cn *chain.Chain, peerStorePath string) *Node {
	nd := &Node{
		ChainID:     ChainID,
		key:         key,
		cn:          cn,
		myPublicKey: key.PublicKey(),
		blockQ:      queue.NewSortedQueue(),
		statusMap:   map[string]*Status{},
		txpool:      txpool.NewTransactionPool(),
		txQ:         queue.NewExpireQueue(),
		txWaitQ:     queue.NewLinkedQueue(),
		txSendQ:     queue.NewQueue(),
		recvChan:    make(chan *RecvMessageItem, 1000),
		sendChan:    make(chan *SendMessageItem, 1000),
		singleCache: gcache.New(500).LRU().Build(),
		batchCache:  gcache.New(500).LRU().Build(),
	}
	nd.ms = NewNodeMesh(nd.ChainID, key, SeedNodeMap, nd, peerStorePath)
	nd.requestTimer = NewRequestTimer(nd)
	nd.txQ.AddGroupRepeat(6, 10*time.Second)
	nd.txQ.AddGroup(600 * time.Second)
	nd.txQ.AddHandler(nd)
	return nd
}

// Init initializes node
func (nd *Node) Init() error {
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
func (nd *Node) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) queue.TxExpiredType {
	item := Item.(*TxMsgItem)
	currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(item.Tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			nd.txpool.Remove(item.TxHash, item.Tx)
			return queue.Expired
		} else if slot > currentSlot+10 {
			return queue.Remain
		}
	}

	if item.Tx.IsEtherType {
		if item.Tx.From == common.ZeroAddr {
			nd.txpool.Remove(item.TxHash, item.Tx)
			return queue.Error
		}
		seq := nd.cn.Provider().AddrSeq(item.Tx.From)
		if seq+100 < item.Tx.Seq {
			return queue.Remain
		}
		if item.Tx.Seq < seq {
			nd.txpool.Remove(item.TxHash, item.Tx)
			return queue.Expired
		}
	}

	nd.txSendQ.Push(item)
	if IsLast {
		nd.txpool.Remove(item.TxHash, item.Tx)
	}
	return queue.Resend
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

	WorkerCount := 1
	switch runtime.NumCPU() {
	case 3:
		WorkerCount = 2
	case 4:
		WorkerCount = 2
	case 5:
		WorkerCount = 3
	case 6:
		WorkerCount = 4
	case 7:
		WorkerCount = 5
	case 8:
		WorkerCount = 6
	default:
		WorkerCount = runtime.NumCPU()/2 + 2
		if WorkerCount >= runtime.NumCPU() {
			WorkerCount = runtime.NumCPU() - 1
		}
		if WorkerCount < 1 {
			WorkerCount = 1
		}
	}

	for i := 0; i < WorkerCount; i++ {
		go func() {
			for !nd.isClose {
				Count := 0
				ctx := nd.cn.NewContext()
				currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
				for !nd.isClose {
					v := nd.txWaitQ.Pop()
					if v == nil {
						break
					}

					item := v.(*TxMsgItem)
					slot := types.ToTimeSlot(item.Tx.Timestamp)
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
							// } else if slot > currentSlot+10 {
							// 	continue
						}
					}
					if ctx.IsUsedTimeSlot(slot, string(item.TxHash[:])) {
						continue
					}
					if err := nd.addTx(item.TxHash, item.Tx, item.Sig); err != nil {
						if errors.Cause(err) != ErrInvalidUTXO &&
							errors.Cause(err) != txpool.ErrExistTransaction &&
							errors.Cause(err) != txpool.ErrTransactionPoolOverflowed &&
							errors.Cause(err) != types.ErrUsedTimeSlot &&
							errors.Cause(err) != types.ErrInvalidTransactionTimeSlot {
							log.Printf("TransactionError %v %+v\n", item.TxHash.String(), err)

							if len(item.PeerID) > 0 {
								nd.ms.AddBadPoint(item.PeerID, 1)
							}
						}
						continue
					}
					//log.Println("TransactionAppended", item.TxHash.String())

					nd.txSendQ.Push(item)

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
		for !nd.isClose {
			if nd.ms.HasPeer() {
				msg := &TransactionMessage{
					Txs:        []*types.Transaction{},
					Signatures: []common.Signature{},
				}
				currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
				for {
					v := nd.txSendQ.Pop()
					if v == nil {
						break
					}
					m := v.(*TxMsgItem)
					slot := types.ToTimeSlot(m.Tx.Timestamp)
					if currentSlot > 0 {
						if slot < currentSlot-1 {
							continue
							// } else if slot > currentSlot+10 {
							// 	continue
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
					log.Printf("PacketToMessage %+v\n", err)
					nd.ms.RemovePeer(item.PeerID)
					break
				}
				if err := nd.handlePeerMessage(item.PeerID, m); err != nil {
					log.Printf("handlePeerMessage %+v\n", err)
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
				var Empty common.PublicKey
				if bytes.Equal(item.Target[:], Empty[:]) {
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
			sm := map[hash.Hash256]common.Address{}
			for _, tx := range b.Body.Transactions {
				TxHash := tx.Hash(b.Header.Height)
				item := nd.txpool.Get(TxHash)
				if item != nil {
					sm[TxHash] = item.Signer
				}
			}
			if err := nd.cn.ConnectBlock(b, sm); err != nil {
				log.Printf("%+v\n", err)
				panic(err)
				// break
			}
			nd.cleanPool(b)
			// if nd.cn.Provider().Height()%100 == 0 {
			log.Println("Node", nd.myPublicKey.Address().String(), nd.cn.Provider().Height(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Body.Transactions))
			// }

			txs := nd.txpool.Clean(types.ToTimeSlot(b.Header.Timestamp))
			if len(txs) > 0 {
				svcs := nd.cn.Services()
				for _, s := range svcs {
					s.OnTransactionInPoolExpired(txs)
				}
				log.Println("Transaction EXPIRED", len(txs))
			}

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

		if Count < 10 {
			if hasItem {
				time.Sleep(50 * time.Millisecond)
			} else {
				time.Sleep(200 * time.Millisecond)
			}
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
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
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
	var SenderPublicKey common.PublicKey
	copy(SenderPublicKey[:], []byte(ID))

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
		nd.sendMessagePacket(0, SenderPublicKey, bs)
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
				nd.sendRequestBlockTo(SenderPublicKey, Height+1, 10)
			} else {
				for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
					if !nd.requestTimer.Exist(i) {
						nd.sendRequestBlockTo(SenderPublicKey, i, 1)
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
				log.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				nd.ms.RemovePeer(ID)
			}
		}
		return nil
	case *BlockMessage:
		for _, b := range msg.Blocks {
			if err := nd.addBlock(b); err != nil {
				if errors.Cause(err) == chain.ErrFoundForkedBlock {
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
		if nd.txWaitQ.Size() > 2000000 {
			return errors.WithStack(txpool.ErrTransactionPoolOverflowed)
		}
		if len(msg.Txs) > 1000 {
			return errors.WithStack(ErrTooManyTrasactionInMessage)
		}
		currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
		for i, tx := range msg.Txs {
			slot := types.ToTimeSlot(tx.Timestamp)
			if currentSlot > 0 {
				if slot < currentSlot-1 {
					continue
				} else if slot > currentSlot+10 {
					continue
				}
			}
			sig := msg.Signatures[i]
			TxHash := tx.HashSig()
			if !nd.txpool.IsExist(TxHash) {
				nd.txWaitQ.Push(TxHash, &TxMsgItem{
					TxHash: TxHash,
					Tx:     tx,
					Sig:    sig,
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
		// return errors.WithStack(ErrUnknownMessage)
	}
	// return nil
}

func (nd *Node) addBlock(b *types.Block) error {
	cp := nd.cn.Provider()
	if b.Header.Height <= cp.Height() {

		bh, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if bh != bin.MustWriterToHash(&b.Header) {
			//TODO : critical error signal
			return errors.WithStack(chain.ErrFoundForkedBlock)
		}
	} else {
		if item := nd.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if bin.MustWriterToHash(&old.Header) != bin.MustWriterToHash(&b.Header) {
				//TODO : critical error signal
				return errors.WithStack(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}

// AddTx adds tx to txpool
func (nd *Node) AddTx(tx *types.Transaction, sig common.Signature) error {
	currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
			// } else if slot > currentSlot+10 {
			// 	return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}
	}

	TxHash := tx.HashSig()
	ctx := nd.cn.NewContext()
	if ctx.IsUsedTimeSlot(slot, string(TxHash[:])) {
		return errors.WithStack(types.ErrUsedTimeSlot)
	}
	if err := nd.addTx(TxHash, tx, sig); err != nil {
		return err
	}
	nd.txSendQ.Push(&TxMsgItem{
		TxHash: TxHash,
		Tx:     tx,
		Sig:    sig,
	})
	return nil
}

// PushTx pushes transaction
func (nd *Node) PushTx(tx *types.Transaction, sig common.Signature) error {
	currentSlot := types.ToTimeSlot(nd.cn.Provider().LastTimestamp())
	slot := types.ToTimeSlot(tx.Timestamp)
	if currentSlot > 0 {
		if slot < currentSlot-1 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
			// } else if slot > currentSlot+10 {
			// 	return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}
	}

	TxHash := tx.HashSig()
	pubkey, err := common.RecoverPubkey(tx.ChainID, tx.Message(), sig)
	if err != nil {
		return err
	}
	tx.From = pubkey.Address()

	if !nd.txpool.IsExist(TxHash) {
		nd.txWaitQ.Push(TxHash, &TxMsgItem{
			TxHash: TxHash,
			Tx:     tx,
			Sig:    sig,
		})
	}
	return nil
}

func (nd *Node) TestFullTx() bool {
	return nd.txWaitQ.Size() > 12000
}

func (nd *Node) addTx(TxHash hash.Hash256, tx *types.Transaction, sig common.Signature) error {
	if nd.txpool.IsExist(TxHash) {
		return errors.WithStack(txpool.ErrExistTransaction)
	}
	pubkey, err := common.RecoverPubkey(tx.ChainID, tx.Message(), sig)
	if err != nil {
		return err
	}
	tx.From = pubkey.Address()

	{
		_ctx := nd.cn.NewContext()
		n := _ctx.Snapshot()

		if tx.UseSeq {
			seq := _ctx.AddrSeq(tx.From)
			if seq != tx.Seq {
				if tx.Seq > seq {
					return fmt.Errorf("future nonce. want %v, get %v signer %v ", seq, tx.Seq, tx.From)
				}
				return errors.Errorf("invalid signer sequence siger %v seq %v, got %v", tx.From, seq, tx.Seq)
			}
		}

		txid := types.TransactionID(_ctx.TargetHeight(), 0)
		if tx.To == common.ZeroAddr {
			_, err = nd.cn.ExecuteTransaction(_ctx, tx, txid)
		} else {
			err = chain.TestContractWithOutSeq(_ctx, tx, tx.From)
		}
		// if err != nil && !strings.Contains(err.Error(), "invalid signer sequence siger seq") {
		if err != nil {
			if strings.Contains(err.Error(), "invalid signer sequence siger") {
				str := strings.Replace(err.Error(), "invalid signer sequence siger ", "", -1)
				str = strings.Replace(str, " got ", "", -1)
				strs := strings.Split(str, " seq ")
				if len(strs) != 2 {
					return err
				}
				strs = strings.Split(strs[1], ",")
				if len(strs) != 2 {
					return err
				}
				seq, _ := strconv.Atoi(strs[0])
				get, _ := strconv.Atoi(strs[1])
				if seq >= get {
					log.Printf("%+v\n", err)
					return err
				}
				return fmt.Errorf("future nonce. want %v, get %v signer %v ", seq, tx.Seq, tx.From)
			} else {
				log.Printf("%+v\n", err)
				return err
			}
		}
		_ctx.Revert(n)
	}

	if err := nd.txpool.Push(TxHash, tx, sig, tx.From); err != nil {
		return err
	}
	if tx.IsEtherType || tx.UseSeq {
		cp := nd.cn.Provider()
		seq := cp.AddrSeq(tx.From)
		if tx.Seq < seq {
			return errors.WithStack(txpool.ErrPastSeq)
		} else if tx.Seq > seq+100 {
			return errors.WithStack(txpool.ErrTooFarSeq)
		}
	}
	nd.txQ.Push(string(TxHash[:]), &TxMsgItem{
		TxHash: TxHash,
		Tx:     tx,
		Sig:    sig,
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

		var TargetPublicKey common.PublicKey
		copy(TargetPublicKey[:], []byte(selectedPubHash))
		if enableCount == 10 {
			nd.sendRequestBlockTo(TargetPublicKey, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
				if !nd.requestTimer.Exist(i) {
					nd.sendRequestBlockTo(TargetPublicKey, i, 1)
				}
			}
		}
	}
}

func (nd *Node) cleanPool(b *types.Block) {
	for _, tx := range b.Body.Transactions {
		TxHash := tx.HashSig()
		nd.txpool.Remove(TxHash, tx)
		nd.txQ.Remove(string(TxHash[:]))
	}
}

// TxPoolList returned tx list from txpool
func (nd *Node) TxPoolList() []*txpool.PoolItem {
	return nd.txpool.List()
}

// TxPoolSize returned tx list size  txpool
func (nd *Node) TxPoolSize() int {
	return nd.txpool.Size()
}

// GetTxFromTXPool returned tx from txpool
func (nd *Node) GetTxFromTXPool(TxHash hash.Hash256) *txpool.PoolItem {
	return nd.txpool.Get(TxHash)
}

// GetTxFromTXPool returned tx from txpool
func (nd *Node) HasPeer() int {
	return len(nd.ms.Peers())
}
