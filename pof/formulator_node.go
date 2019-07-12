package pof

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// FormulatorConfig defines configuration of the formulator
type FormulatorConfig struct {
	Formulator              common.Address
	MaxTransactionsPerBlock int
}

// FormulatorNode procudes a block by the consensus
type FormulatorNode struct {
	sync.Mutex
	Config               *FormulatorConfig
	cs                   *Consensus
	ms                   *FormulatorNodeMesh
	nm                   *p2p.NodeMesh
	key                  key.Key
	myPublicHash         common.PublicHash
	lastGenMessages      []*BlockGenMessage
	lastObSignMessageMap map[uint32]*BlockObSignMessage
	lastContextes        []*types.Context
	lastReqMessage       *BlockReqMessage
	lastGenHeight        uint32
	txMsgChans           []*chan *p2p.TxMsgItem
	txMsgIdx             uint64
	statusMap            map[string]*p2p.Status
	requestTimer         *p2p.RequestTimer
	requestLock          sync.RWMutex
	blockQ               *queue.SortedQueue
	txpool               *txpool.TransactionPool
	txQ                  *queue.ExpireQueue
	isRunning            bool
	closeLock            sync.RWMutex
	runEnd               chan struct{}
	isClose              bool
}

// NewFormulatorNode returns a FormulatorNode
func NewFormulatorNode(Config *FormulatorConfig, key key.Key, NetAddressMap map[common.PublicHash]string, SeedNodeMap map[common.PublicHash]string, cs *Consensus, peerStorePath string) *FormulatorNode {
	fr := &FormulatorNode{
		Config:               Config,
		cs:                   cs,
		key:                  key,
		myPublicHash:         common.NewPublicHash(key.PublicKey()),
		lastGenMessages:      []*BlockGenMessage{},
		lastObSignMessageMap: map[uint32]*BlockObSignMessage{},
		lastContextes:        []*types.Context{},
		statusMap:            map[string]*p2p.Status{},
		requestTimer:         p2p.NewRequestTimer(nil),
		runEnd:               make(chan struct{}),
		blockQ:               queue.NewSortedQueue(),
		txpool:               txpool.NewTransactionPool(),
		txQ:                  queue.NewExpireQueue(),
	}
	fr.ms = NewFormulatorNodeMesh(key, NetAddressMap, fr)
	fr.nm = p2p.NewNodeMesh(key, SeedNodeMap, fr, peerStorePath)
	fr.txQ.AddGroup(60 * time.Second)
	fr.txQ.AddGroup(600 * time.Second)
	fr.txQ.AddGroup(3600 * time.Second)
	fr.txQ.AddHandler(fr)
	return fr
}

// Close terminates the formulator
func (fr *FormulatorNode) Close() {
	fr.closeLock.Lock()
	defer fr.closeLock.Unlock()

	fr.Lock()
	defer fr.Unlock()

	fr.isClose = true
	fr.cs.cn.Close()
	fr.runEnd <- struct{}{}
}

// Init initializes formulator
func (fr *FormulatorNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenMessage"), &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.TransactionMessage"), &p2p.TransactionMessage{})
	fc.Register(types.DefineHashedType("p2p.PeerListMessage"), &p2p.PeerListMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestPeerListMessage"), &p2p.RequestPeerListMessage{})
	return nil
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

	WorkerCount := runtime.NumCPU() - 1
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	workerEnd := make([]*chan struct{}, WorkerCount)
	fr.txMsgChans = make([]*chan *p2p.TxMsgItem, WorkerCount)
	for i := 0; i < WorkerCount; i++ {
		mch := make(chan *p2p.TxMsgItem)
		fr.txMsgChans[i] = &mch
		ch := make(chan struct{})
		workerEnd[i] = &ch
		go func(pMsgCh *chan *p2p.TxMsgItem, pEndCh *chan struct{}) {
			for {
				select {
				case item := <-(*pMsgCh):
					if err := fr.AddTx(item.Message.Tx, item.Message.Sigs); err != nil {
						if err != txpool.ErrPastSeq || err != txpool.ErrTooFarSeq {
							(*item.ErrCh) <- err
						} else {
							(*item.ErrCh) <- nil
						}
						break
					}
					(*item.ErrCh) <- nil
					if len(item.PeerID) > 0 {
						//fr.pm.ExceptCast(item.PeerID, item.Message)
						//fr.pm.ExceptCastLimit(item.PeerID, item.Message, 7)
					} else {
						//fr.pm.BroadCast(item.Message)
						//fr.pm.BroadCastLimit(item.Message, 7)
					}
				case <-(*pEndCh):
					return
				}
			}
		}(&mch, &ch)
	}

	blockTimer := time.NewTimer(time.Millisecond)
	for !fr.isClose {
		select {
		case <-blockTimer.C:
			cp := fr.cs.cn.Provider()
			fr.Lock()
			TargetHeight := uint64(cp.Height() + 1)
			item := fr.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := fr.cs.cn.ConnectBlock(b); err != nil {
					break
				}
				fr.broadcastStatus()
				fr.cleanPool(b)
				TargetHeight++
				item = fr.blockQ.PopUntil(TargetHeight)
			}
			fr.Unlock()
			blockTimer.Reset(50 * time.Millisecond)
		case <-fr.runEnd:
			for i := 0; i < WorkerCount; i++ {
				(*workerEnd[i]) <- struct{}{}
			}
		}
	}
}

// AddTx adds tx to txpool that only have valid signatures
func (fr *FormulatorNode) AddTx(tx types.Transaction, sigs []common.Signature) error {
	if fr.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}

	TxHash := chain.HashTransactionByType(t, tx)

	ctx := fr.cs.ct.NewContext()
	if fr.txpool.IsExist(TxHash) {
		return txpool.ErrExistTransaction
	}
	if atx, is := tx.(txpool.AccountTransaction); is {
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
	p, err := fr.cs.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := fr.txpool.Push(t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	fr.txQ.Push(string(TxHash[:]), &p2p.TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

// OnTimerExpired called when rquest expired
func (fr *FormulatorNode) OnTimerExpired(height uint32, value interface{}) {
	go fr.tryRequestNext()
}

// OnItemExpired is called when the item is expired
func (fr *FormulatorNode) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	msg := Item.(*p2p.TransactionMessage)
	fr.ms.BroadcastMessage(msg)
	fr.nm.BroadcastMessage(msg)
	if IsLast {
		var TxHash hash.Hash256
		copy(TxHash[:], []byte(Key))
		fr.txpool.Remove(TxHash, msg.Tx)
	}
}

// OnObserverConnected is called after a new observer peer is connected
func (fr *FormulatorNode) OnObserverConnected(p peer.Peer) {
	fr.Lock()
	fr.statusMap[p.ID()] = &p2p.Status{}
	fr.Unlock()
}

// OnObserverDisconnected is called when the observer peer is disconnected
func (fr *FormulatorNode) OnObserverDisconnected(p peer.Peer) {
	fr.Lock()
	delete(fr.statusMap, p.ID())
	fr.Unlock()
}

// OnRecv called when message received
func (fr *FormulatorNode) OnRecv(p peer.Peer, m interface{}) error {
	cp := fr.cs.cn.Provider()

	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		b, err := cp.Block(msg.Height)
		if err != nil {
			return err
		}
		sm := &p2p.BlockMessage{
			Block: b,
		}
		if err := fr.nm.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		Height := cp.Height()
		if Height < msg.Height {
			for i := Height + 1; i <= Height+100 && i <= msg.Height; i++ {
				if !fr.requestTimer.Exist(i) {
					fr.sendRequestBlockToNode(SenderPublicHash, i)
				}
			}
		} else {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				log.Println(p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *p2p.BlockMessage:
		if err := fr.addBlock(msg.Block); err != nil {
			return err
		}
		fr.requestTimer.Remove(msg.Block.Header.Height)
	case *p2p.PeerListMessage:
		fr.nm.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *p2p.RequestPeerListMessage:
		fr.nm.SendPeerList(p.ID())
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (fr *FormulatorNode) onRecv(p peer.Peer, m interface{}) error {
	if err := fr.handleMessage(p, m, 0); err != nil {
		//log.Println(err)
		return nil
	}
	return nil
}

func (fr *FormulatorNode) handleMessage(p peer.Peer, m interface{}, RetryCount int) error {
	cp := fr.cs.cn.Provider()

	switch msg := m.(type) {
	case *BlockReqMessage:
		log.Println("Formulator", fr.Config.Formulator.String(), "BlockReqMessage", msg.TargetHeight, msg.RemainBlocks)
		fr.Lock()
		defer fr.Unlock()

		Height := cp.Height()
		if msg.TargetHeight <= fr.lastGenHeight {
			return nil
		}
		if msg.TargetHeight <= Height {
			return nil
		}
		if msg.TargetHeight > Height+1 {
			if RetryCount >= 40 {
				return nil
			}
			go func() {
				fr.tryRequestNext()
				time.Sleep(50 * time.Millisecond)
				fr.handleMessage(p, m, RetryCount+1)
			}()
			return nil
		}

		Top, err := fr.cs.rt.TopRank(int(msg.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.Formulator != Top.Address {
			return ErrInvalidRequest
		}
		if msg.Formulator != fr.Config.Formulator {
			return ErrInvalidRequest
		}
		if msg.FormulatorPublicHash != common.NewPublicHash(fr.key.PublicKey()) {
			return ErrInvalidRequest
		}
		if msg.PrevHash != cp.LastHash() {
			return ErrInvalidRequest
		}
		if msg.TargetHeight != Height+1 {
			return ErrInvalidRequest
		}

		fr.lastGenMessages = []*BlockGenMessage{}
		fr.lastObSignMessageMap = map[uint32]*BlockObSignMessage{}
		fr.lastContextes = []*types.Context{}
		fr.lastReqMessage = msg

		var ctx *types.Context
		start := time.Now().UnixNano()
		StartBlockTime := uint64(time.Now().UnixNano())
		bNoDelay := false
		TargetBlocksInTurn := msg.RemainBlocks
		if Height > 0 {
			LastHeader, err := cp.Header(Height)
			if err != nil {
				return err
			}
			if StartBlockTime < LastHeader.Timestamp {
				StartBlockTime = LastHeader.Timestamp + uint64(time.Millisecond)
			} else if LastHeader.Timestamp < uint64(msg.RemainBlocks)*uint64(500*time.Millisecond) {
				bNoDelay = true
			}
		}
		for i := uint32(0); i < TargetBlocksInTurn; i++ {
			var TimeoutCount uint32
			if i == 0 {
				ctx = fr.cs.ct.NewContext()
				TimeoutCount = msg.TimeoutCount
			} else {
				lastHeader := fr.lastGenMessages[len(fr.lastGenMessages)-1].Block.Header
				ctx = ctx.NextContext(encoding.Hash(lastHeader), lastHeader.Timestamp)
			}

			Timestamp := StartBlockTime
			if bNoDelay {
				Timestamp += uint64(i) * uint64(time.Millisecond)
			} else {
				Timestamp += uint64(i) * uint64(500*time.Millisecond)
			}

			var buffer bytes.Buffer
			enc := encoding.NewEncoder(&buffer)
			if err := enc.EncodeUint32(TimeoutCount); err != nil {
				return err
			}
			bc := chain.NewBlockCreator(fr.cs.cn, ctx, msg.Formulator, buffer.Bytes())
			if err := bc.Init(); err != nil {
				return err
			}

			timer := time.NewTimer(200 * time.Millisecond)

			fr.txpool.Lock() // Prevent delaying from TxPool.Push
			Count := 0
		TxLoop:
			for {
				select {
				case <-timer.C:
					break TxLoop
				default:
					item := fr.txpool.UnsafePop(ctx)
					if item == nil {
						break TxLoop
					}
					if err := bc.UnsafeAddTx(item.TxType, item.TxHash, item.Transaction, item.Signatures, item.Signers); err != nil {
						log.Println(err)
						continue
					}
					Count++
					if Count > fr.Config.MaxTransactionsPerBlock {
						break TxLoop
					}
				}
			}
			fr.txpool.Unlock() // Prevent delaying from TxPool.Push

			b, err := bc.Finalize()
			if err != nil {
				return err
			}

			nm := &BlockGenMessage{
				Block: b,
			}

			if sig, err := fr.key.Sign(encoding.Hash(b.Header)); err != nil {
				return err
			} else {
				nm.GeneratorSignature = sig
			}

			if err := p.Send(nm); err != nil {
				return err
			}
			log.Println("Formulator", fr.Config.Formulator.String(), "BlockGenMessage", nm.Block.Header.Height, len(nm.Block.Transactions))

			fr.lastGenMessages = append(fr.lastGenMessages, nm)
			fr.lastContextes = append(fr.lastContextes, ctx)
			fr.lastGenHeight = ctx.TargetHeight()

			ExpectedTime := time.Duration(i+1) * 200 * time.Millisecond
			PastTime := time.Duration(time.Now().UnixNano() - start)
			if ExpectedTime > PastTime {
				time.Sleep(ExpectedTime - PastTime)
			}
		}
		return nil
	case *BlockObSignMessage:
		fr.Lock()
		defer fr.Unlock()

		if len(fr.lastGenMessages) == 0 {
			return nil
		}
		if msg.TargetHeight <= fr.cs.cn.Provider().Height() {
			return nil
		}
		if msg.TargetHeight >= fr.lastReqMessage.TargetHeight+10 {
			return ErrInvalidRequest
		}
		fr.lastObSignMessageMap[msg.TargetHeight] = msg

		for len(fr.lastGenMessages) > 0 {
			GenMessage := fr.lastGenMessages[0]
			sm, has := fr.lastObSignMessageMap[GenMessage.Block.Header.Height]
			if !has {
				break
			}
			if GenMessage.Block.Header.Height == sm.TargetHeight {
				ctx := fr.lastContextes[0]

				if sm.BlockSign.HeaderHash != encoding.Hash(GenMessage.Block.Header) {
					return ErrInvalidRequest
				}

				b := &types.Block{
					Header:               GenMessage.Block.Header,
					TransactionTypes:     GenMessage.Block.TransactionTypes,
					Transactions:         GenMessage.Block.Transactions,
					TranactionSignatures: GenMessage.Block.TranactionSignatures,
					Signatures:           append([]common.Signature{GenMessage.GeneratorSignature}, sm.ObserverSignatures...),
				}
				if err := fr.cs.ct.ConnectBlockWithContext(b, ctx); err != nil {
					return err
				}
				fr.broadcastStatus()
				fr.cleanPool(b)
				log.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))

				if status, has := fr.statusMap[p.ID()]; has {
					if status.Height < GenMessage.Block.Header.Height {
						status.Height = GenMessage.Block.Header.Height
					}
				}

				if len(fr.lastGenMessages) > 1 {
					fr.lastGenMessages = fr.lastGenMessages[1:]
					fr.lastContextes = fr.lastContextes[1:]
				} else {
					fr.lastGenMessages = []*BlockGenMessage{}
					fr.lastContextes = []*types.Context{}
				}
			}
		}
		return nil
	case *p2p.RequestMessage:
		b, err := cp.Block(msg.Height)
		if err != nil {
			return err
		}
		sm := &p2p.BlockMessage{
			Block: b,
		}
		if err := fr.ms.SendTo(p.ID(), sm); err != nil {
			return err
		}
		return nil
	case *p2p.BlockMessage:
		if msg.Block.Header.Height <= fr.cs.cn.Provider().Height() {
			return nil
		}
		if err := fr.addBlock(msg.Block); err != nil {
			return err
		}
		fr.requestTimer.Remove(msg.Block.Header.Height)

		fr.Lock()
		if status, has := fr.statusMap[p.ID()]; has {
			if status.Height < msg.Block.Header.Height {
				status.Height = msg.Block.Header.Height
			}
		}
		fr.Unlock()

		fr.tryRequestNext()
		return nil
	case *p2p.StatusMessage:
		fr.Lock()
		defer fr.Unlock()

		if status, has := fr.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}

		TargetHeight := fr.cs.cn.Provider().Height() + 1
		for TargetHeight <= msg.Height {
			if !fr.requestTimer.Exist(TargetHeight) {
				if fr.blockQ.Find(uint64(TargetHeight)) == nil {
					sm := &p2p.RequestMessage{
						Height: TargetHeight,
					}
					if err := p.Send(sm); err != nil {
						return err
					}
					fr.requestTimer.Add(TargetHeight, 10*time.Second, p.ID())
				}
			}
			TargetHeight++
		}
		return nil
	case *p2p.TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&fr.txMsgIdx, 1) % uint64(len(fr.txMsgChans))
		(*fr.txMsgChans[idx]) <- &p2p.TxMsgItem{
			Message: msg,
			PeerID:  "",
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
			panic(chain.ErrFoundForkedBlock)
		}
	} else {
		if item := fr.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				panic(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}

func (fr *FormulatorNode) tryRequestNext() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	TargetHeight := fr.cs.cn.Provider().Height() + 1
	if !fr.requestTimer.Exist(TargetHeight) {
		if fr.blockQ.Find(uint64(TargetHeight)) == nil {
			fr.Lock()
			defer fr.Unlock()

			for pubhash, status := range fr.statusMap {
				if TargetHeight <= status.Height {
					sm := &p2p.RequestMessage{
						Height: TargetHeight,
					}
					if err := fr.ms.SendTo(pubhash, sm); err != nil {
						return
					}
					fr.requestTimer.Add(TargetHeight, 5*time.Second, pubhash)
					return
				}
			}
		}
	}
}

func (fr *FormulatorNode) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(t, tx)
		fr.txpool.Remove(TxHash, tx)
		fr.txQ.Remove(string(TxHash[:]))
	}
}
