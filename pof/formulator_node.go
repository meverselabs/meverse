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
	ndkey                key.Key
	myPublicHash         common.PublicHash
	lastGenMessages      []*BlockGenMessage
	lastObSignMessageMap map[uint32]*BlockObSignMessage
	lastContextes        []*types.Context
	lastReqMessage       *BlockReqMessage
	lastGenHeight        uint32
	lastGenTime          int64
	txMsgChans           []*chan *p2p.TxMsgItem
	txMsgIdx             uint64
	statusMap            map[string]*p2p.Status
	obStatusMap          map[string]*p2p.Status
	requestTimer         *p2p.RequestTimer
	requestNodeTimer     *p2p.RequestTimer
	requestLock          sync.RWMutex
	blockQ               *queue.SortedQueue
	txpool               *txpool.TransactionPool
	txQ                  *queue.ExpireQueue
	isRunning            bool
	closeLock            sync.RWMutex
	isClose              bool
}

// NewFormulatorNode returns a FormulatorNode
func NewFormulatorNode(Config *FormulatorConfig, key key.Key, ndkey key.Key, NetAddressMap map[common.PublicHash]string, SeedNodeMap map[common.PublicHash]string, cs *Consensus, peerStorePath string) *FormulatorNode {
	if Config.MaxTransactionsPerBlock == 0 {
		Config.MaxTransactionsPerBlock = 10000
	}
	fr := &FormulatorNode{
		Config:               Config,
		cs:                   cs,
		key:                  key,
		ndkey:                ndkey,
		myPublicHash:         common.NewPublicHash(ndkey.PublicKey()),
		lastGenMessages:      []*BlockGenMessage{},
		lastObSignMessageMap: map[uint32]*BlockObSignMessage{},
		lastContextes:        []*types.Context{},
		statusMap:            map[string]*p2p.Status{},
		obStatusMap:          map[string]*p2p.Status{},
		requestTimer:         p2p.NewRequestTimer(nil),
		requestNodeTimer:     p2p.NewRequestTimer(nil),
		blockQ:               queue.NewSortedQueue(),
		txpool:               txpool.NewTransactionPool(),
		txQ:                  queue.NewExpireQueue(),
	}
	fr.ms = NewFormulatorNodeMesh(key, NetAddressMap, fr)
	fr.nm = p2p.NewNodeMesh(fr.cs.cn.Provider().ChainID(), ndkey, SeedNodeMap, fr, peerStorePath)
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
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})
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
	go fr.requestNodeTimer.Run()

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
					if err := fr.addTx(item.Message.TxType, item.Message.Tx, item.Message.Sigs); err != nil {
						if err != txpool.ErrPastSeq || err != txpool.ErrTooFarSeq {
							(*item.ErrCh) <- err
						} else {
							(*item.ErrCh) <- nil
						}
						break
					}
					(*item.ErrCh) <- nil

					fr.nm.ExceptCastLimit(item.PeerID, item.Message, 7)
				case <-(*pEndCh):
					return
				}
			}
		}(&mch, &ch)
	}

	blockTimer := time.NewTimer(time.Millisecond)
	blockRequestTimer := time.NewTimer(time.Millisecond)
	for !fr.isClose {
		select {
		case <-blockTimer.C:
			fr.Lock()
			hasItem := false
			TargetHeight := uint64(fr.cs.cn.Provider().Height() + 1)
			item := fr.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := fr.cs.cn.ConnectBlock(b); err != nil {
					break
				}
				log.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))
				fr.broadcastStatus()
				fr.cleanPool(b)
				TargetHeight++
				item = fr.blockQ.PopUntil(TargetHeight)
				hasItem = true
			}
			fr.Unlock()

			if hasItem {
				fr.tryRequestBlocks()
			}

			blockTimer.Reset(50 * time.Millisecond)
		case <-blockRequestTimer.C:
			fr.tryRequestBlocks()
			fr.tryRequestNext()
			blockRequestTimer.Reset(500 * time.Millisecond)
		}
	}
	for i := 0; i < WorkerCount; i++ {
		(*workerEnd[i]) <- struct{}{}
	}
}

// AddTx adds tx to txpool that only have valid signatures
func (fr *FormulatorNode) AddTx(tx types.Transaction, sigs []common.Signature) error {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	if err := fr.addTx(t, tx, sigs); err != nil {
		return err
	}
	fr.nm.ExceptCastLimit("", &p2p.TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	}, 7)
	return nil
}

func (fr *FormulatorNode) addTx(t uint16, tx types.Transaction, sigs []common.Signature) error {
	if fr.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), t, tx)

	ctx := fr.cs.ct.NewContext()
	if fr.txpool.IsExist(TxHash) {
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
	p, err := fr.cs.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := fr.txpool.Push(fr.cs.cn.Provider().ChainID(), t, TxHash, tx, sigs, signers); err != nil {
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
func (fr *FormulatorNode) OnTimerExpired(height uint32, value string) {
	go fr.tryRequestBlocks()
}

// OnItemExpired is called when the item is expired
func (fr *FormulatorNode) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	msg := Item.(*p2p.TransactionMessage)
	fr.nm.ExceptCastLimit("", msg, 7)
	if IsLast {
		var TxHash hash.Hash256
		copy(TxHash[:], []byte(Key))
		fr.txpool.Remove(TxHash, msg.Tx)
	}
}

// OnObserverConnected is called after a new observer peer is connected
func (fr *FormulatorNode) OnObserverConnected(p peer.Peer) {
	fr.Lock()
	fr.obStatusMap[p.ID()] = &p2p.Status{}
	fr.Unlock()
}

// OnObserverDisconnected is called when the observer peer is disconnected
func (fr *FormulatorNode) OnObserverDisconnected(p peer.Peer) {
	fr.Lock()
	delete(fr.obStatusMap, p.ID())
	fr.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestNext()
}

// OnConnected is called after a new  peer is connected
func (fr *FormulatorNode) OnConnected(p peer.Peer) {
	fr.Lock()
	fr.statusMap[p.ID()] = &p2p.Status{}
	fr.Unlock()
}

// OnDisconnected is called when the  peer is disconnected
func (fr *FormulatorNode) OnDisconnected(p peer.Peer) {
	fr.Lock()
	delete(fr.statusMap, p.ID())
	fr.Unlock()
	fr.requestNodeTimer.RemovesByValue(p.ID())
	go fr.tryRequestBlocks()
}

// OnRecv called when message received
func (fr *FormulatorNode) OnRecv(p peer.Peer, m interface{}) error {
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
		cp := fr.cs.cn.Provider()
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
		if err := fr.nm.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		fr.Lock()
		if status, has := fr.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		fr.Unlock()

		Height := fr.cs.cn.Provider().Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
					if !fr.requestTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					fr.sendRequestBlockToNode(SenderPublicHash, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
						if !fr.requestTimer.Exist(i) {
							fr.sendRequestBlockToNode(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := fr.cs.cn.Provider().Hash(msg.Height)
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
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.Lock()
			if status, has := fr.statusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.Unlock()
		}
	case *p2p.TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&fr.txMsgIdx, 1) % uint64(len(fr.txMsgChans))
		(*fr.txMsgChans[idx]) <- &p2p.TxMsgItem{
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

func (fr *FormulatorNode) tryRequestBlocks() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	Height := fr.cs.cn.Provider().Height()
	for q := uint32(0); q < 10; q++ {
		BaseHeight := Height + q*10

		var selectedPubHash string
		fr.Lock()
		for pubhash, status := range fr.statusMap {
			if BaseHeight <= status.Height {
				selectedPubHash = pubhash
				break
			}
		}
		fr.Unlock()

		if len(selectedPubHash) == 0 {
			break
		}
		enableCount := 0
		for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
			if !fr.requestTimer.Exist(i) {
				enableCount++
			}
		}

		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(selectedPubHash))
		if enableCount == 10 {
			fr.sendRequestBlockToNode(TargetPublicHash, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
				if !fr.requestTimer.Exist(i) {
					fr.sendRequestBlockToNode(TargetPublicHash, i, 1)
				}
			}
		}
	}
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
		log.Println("Formulator", fr.Config.Formulator.String(), "BlockReqMessage", msg.TargetHeight)

		fr.Lock()
		defer fr.Unlock()

		Height := cp.Height()
		if msg.TargetHeight <= fr.lastGenHeight && fr.lastGenTime+int64(30*time.Second) > time.Now().UnixNano() {
			return nil
		}
		if msg.TargetHeight <= Height {
			return nil
		}
		if msg.TargetHeight > Height+1 {
			if RetryCount >= 5 {
				return nil
			}

			Count := uint8(msg.TargetHeight - Height - 1)
			if Count > 10 {
				Count = 10
			}

			sm := &p2p.RequestMessage{
				Height: Height + 1,
				Count:  Count,
			}
			if err := p.Send(sm); err != nil {
				return err
			}
			go func() {
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

		start := time.Now().UnixNano()
		Now := uint64(time.Now().UnixNano())
		StartBlockTime := Now
		bNoDelay := false

		RemainBlocks := fr.cs.maxBlocksPerFormulator
		if msg.TimeoutCount == 0 {
			RemainBlocks = fr.cs.maxBlocksPerFormulator - fr.cs.blocksBySameFormulator
		}

		LastTimestamp := cp.LastTimestamp()
		if StartBlockTime < LastTimestamp {
			StartBlockTime = LastTimestamp + uint64(time.Millisecond)
		} else if StartBlockTime > LastTimestamp+uint64(RemainBlocks)*uint64(500*time.Millisecond) {
			bNoDelay = true
		}
		ctx := fr.cs.ct.NewContext()
		for i := uint32(0); i < RemainBlocks; i++ {
			var TimeoutCount uint32
			if i == 0 {
				TimeoutCount = msg.TimeoutCount
			} else {
				lastHeader := fr.lastGenMessages[len(fr.lastGenMessages)-1].Block.Header
				ctx = ctx.NextContext(encoding.Hash(lastHeader), lastHeader.Timestamp)
			}

			Timestamp := StartBlockTime
			if bNoDelay || Timestamp > Now+uint64(3*time.Second) {
				Timestamp += uint64(i) * uint64(time.Millisecond)
			} else {
				Timestamp += uint64(i) * uint64(500*time.Millisecond)
			}
			if Timestamp <= ctx.LastTimestamp() {
				Timestamp = ctx.LastTimestamp() + 1
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
					sn := ctx.Snapshot()
					item := fr.txpool.UnsafePop(ctx)
					ctx.Revert(sn)
					if item == nil {
						break TxLoop
					}
					if err := bc.UnsafeAddTx(fr.Config.Formulator, item.TxType, item.TxHash, item.Transaction, item.Signatures, item.Signers); err != nil {
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

			b, err := bc.Finalize(Timestamp)
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
			fr.lastGenTime = time.Now().UnixNano()

			ExpectedTime := time.Duration(i+1) * 200 * time.Millisecond
			PastTime := time.Duration(time.Now().UnixNano() - start)
			if !bNoDelay && ExpectedTime > PastTime {
				time.Sleep(ExpectedTime - PastTime)
			}
		}
		return nil
	case *BlockObSignMessage:
		log.Println("Formulator", fr.Config.Formulator.String(), "BlockObSignMessage", msg.TargetHeight)

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
					Header:                GenMessage.Block.Header,
					TransactionTypes:      GenMessage.Block.TransactionTypes,
					Transactions:          GenMessage.Block.Transactions,
					TransactionSignatures: GenMessage.Block.TransactionSignatures,
					TransactionResults:    GenMessage.Block.TransactionResults,
					Signatures:            append([]common.Signature{GenMessage.GeneratorSignature}, sm.ObserverSignatures...),
				}
				if err := fr.cs.ct.ConnectBlockWithContext(b, ctx); err != nil {
					return err
				}
				fr.broadcastStatus()
				fr.cleanPool(b)
				log.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))

				if status, has := fr.obStatusMap[p.ID()]; has {
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
		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
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
		if err := fr.ms.SendTo(p.ID(), sm); err != nil {
			return err
		}
		return nil
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.Lock()
			if status, has := fr.obStatusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.Unlock()
		}
		return nil
	case *p2p.StatusMessage:
		fr.Lock()
		if status, has := fr.obStatusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Version = msg.Version
				status.Height = msg.Height
				status.LastHash = msg.LastHash
			}
		}
		fr.Unlock()

		Height := cp.Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
					if !fr.requestTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					fr.sendRequestBlockTo(p.ID(), BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10; i++ {
						if !fr.requestTimer.Exist(i) {
							fr.sendRequestBlockTo(p.ID(), i, 1)
						}
					}
				}
			}
		}
		return nil
	case *p2p.TransactionMessage:
		errCh := make(chan error)
		idx := atomic.AddUint64(&fr.txMsgIdx, 1) % uint64(len(fr.txMsgChans))
		(*fr.txMsgChans[idx]) <- &p2p.TxMsgItem{
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
			var TargetPubHash string
			for pubhash, status := range fr.obStatusMap {
				if TargetHeight <= status.Height {
					TargetPubHash = pubhash
					break
				}
			}
			fr.Unlock()

			if len(TargetPubHash) > 0 {
				fr.sendRequestBlockTo(TargetPubHash, TargetHeight, 1)
			}
		}
	}
}

func (fr *FormulatorNode) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), t, tx)
		fr.txpool.Remove(TxHash, tx)
		fr.txQ.Remove(string(TxHash[:]))
	}
}
