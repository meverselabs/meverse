package node

import (
	"fmt"
	"log"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

// OnObserverConnected is called after a new observer peer is connected
func (fr *GeneratorNode) OnObserverConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.obStatusMap[p.ID()] = &p2p.Status{}
	fr.statusLock.Unlock()

	cp := fr.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnObserverDisconnected is called when the observer peer is disconnected
func (fr *GeneratorNode) OnObserverDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.obStatusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestNext()
}

func (fr *GeneratorNode) onObserverRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}

	if err := fr.handleObserverMessage(p, m, 0); err != nil {
		switch errors.Cause(err) {
		case ErrInvalidRoundState, ErrAlreadyVoted:
		default:
			fmt.Printf("%+v\n", err)
		}
		return nil
	}
	return nil
}

func (fr *GeneratorNode) handleObserverMessage(p peer.Peer, m interface{}, RetryCount int) error {
	cp := fr.cn.Provider()

	switch msg := m.(type) {
	case *BlockReqMessage:
		if DEBUG {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "BlockReqMessage", msg.TargetHeight)
		}

		TargetHeight := fr.cn.Provider().Height() + 1
		if msg.TargetHeight < TargetHeight {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Past Target Height")
			return nil
		}
		if msg.TargetHeight <= fr.lastGenHeight {
			if time.Now().UnixNano() < fr.lastGenTime+int64(10*time.Second) {
				var nm *BlockGenMessage
				fr.Lock()
				if gm, has := fr.lastGenItemMap[msg.TargetHeight]; has {
					nm = gm.BlockGen
				}
				fr.Unlock()

				if nm != nil {
					fr.ms.SendTo(p.ID(), nm)
				}
				log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Wait 30 Sec")
				return nil
			}
			fr.lastReqLock.Lock()
			fr.lastReqMessage = nil
			fr.lastReqLock.Unlock()
		}
		fr.lastReqLock.Lock()
		if fr.lastReqMessage != nil {
			if msg.TargetHeight <= fr.lastReqMessage.TargetHeight {
				fr.lastReqLock.Unlock()
				log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Current Target Height")
				return nil
			}
		}
		fr.lastReqLock.Unlock()

		fr.Lock()
		defer fr.Unlock()

		if msg.TargetHeight > TargetHeight {
			if msg.TargetHeight > TargetHeight+10 {
				log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Far Future Target Height")
				return nil
			}
			if RetryCount >= 10 {
				log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Retry Timeover")
				return nil
			}
			if RetryCount == 0 {
				Count := uint8(msg.TargetHeight - TargetHeight)
				if Count > 10 {
					Count = 10
				}

				sm := &p2p.RequestMessage{
					Height: TargetHeight,
					Count:  Count,
				}
				p.SendPacket(p2p.MessageToPacket(sm))
			}
			go func() {
				time.Sleep(50 * time.Millisecond)
				if err := fr.handleObserverMessage(p, m, RetryCount+1); err != nil {
					switch errors.Unwrap(err) {
					case ErrInvalidRoundState, ErrAlreadyVoted:
					default:
						fmt.Printf("%+v\n", err)
					}
				}
			}()
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Future Height")
			return nil
		}

		if msg.Generator != fr.key.PublicKey().Address() {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Not My Address")
			return errors.WithStack(ErrInvalidRequest)
		}
		if msg.PrevHash != cp.LastHash() {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Not Prev Hash")
			return errors.WithStack(ErrInvalidRequest)
		}

		Top, err := fr.cn.TopGenerator(msg.TimeoutCount)
		if err != nil {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Invalid Top")
			return err
		}
		if msg.Generator != Top {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Invalid Top2")
			return errors.WithStack(ErrInvalidRequest)
		}
		fr.lastReqLock.Lock()
		fr.lastReqMessage = msg
		fr.lastReqLock.Unlock()

		go func(ID string, req *BlockReqMessage) error {
			fr.genLock.Lock()
			defer fr.genLock.Unlock()

			fr.Lock()
			defer fr.Unlock()

			TargetHeight := fr.cn.Provider().Height() + 1
			if req.TargetHeight < TargetHeight {
				return nil
			}

			fr.lastReqLock.Lock()
			if fr.lastReqMessage != nil {
				if req.TargetHeight < fr.lastReqMessage.TargetHeight {
					fr.lastReqLock.Unlock()
					return nil
				}
			}
			fr.lastReqLock.Unlock()

			err := fr.genBlock(ID, req)
			if err != nil {
				fmt.Printf("gen block err %+v", err)
			}
			return err
		}(p.ID(), msg)
		return nil
	case *BlockGenMessage:
		log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Recv.BlockGenMessage", msg.Block.Header.Height)

		TargetHeight := fr.cn.Provider().Height() + 1
		if msg.Block.Header.Height < TargetHeight {
			return nil
		}
		if msg.Block.Header.Generator != fr.key.PublicKey().Address() {
			fr.lastReqLock.Lock()
			fr.lastReqMessage = nil
			fr.lastReqLock.Unlock()
		}
		fr.Lock()
		defer fr.Unlock()

		item, has := fr.lastGenItemMap[msg.Block.Header.Height]
		if has {
			if item.ObSign != nil {
				if item.ObSign.BlockSign.HeaderHash != bin.MustWriterToHash(&msg.Block.Header) {
					return errors.WithStack(ErrInvalidRequest)
				}
			}
			item.BlockGen = msg
			item.Recv = true
		} else {
			item = &genItem{
				BlockGen: msg,
				ObSign:   nil,
				Context:  nil,
				Recv:     true,
			}
			fr.lastGenItemMap[msg.Block.Header.Height] = item
		}

		go fr.updateByGenItem()
		return nil
	case *BlockObSignMessage:
		if DEBUG {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Recv.BlockObSignMessage", msg.TargetHeight)
		}

		TargetHeight := fr.cn.Provider().Height() + 1
		if msg.TargetHeight < TargetHeight {
			return nil
		}

		fr.Lock()
		if item, has := fr.lastGenItemMap[msg.TargetHeight]; has {
			if item.BlockGen != nil {
				if msg.BlockSign.HeaderHash != bin.MustWriterToHash(&item.BlockGen.Block.Header) {
					item.BlockGen = nil
					item.Context = nil
				}
			}
			item.ObSign = msg
		} else {
			fr.lastGenItemMap[msg.TargetHeight] = &genItem{
				BlockGen: nil,
				ObSign:   msg,
				Context:  nil,
			}
		}
		fr.Unlock()

		fr.statusLock.Lock()
		if status, has := fr.obStatusMap[p.ID()]; has {
			if status.Height < msg.TargetHeight {
				status.Height = msg.TargetHeight
			}
		}
		fr.statusLock.Unlock()

		go fr.updateByGenItem()
		return nil
	case *p2p.BlockMessage:
		if DEBUG {
			log.Println("Generator Recv.Ob.BlockMessage", msg.Blocks[0].Header.Height)
		}
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if errors.Cause(err) == chain.ErrFoundForkedBlock {
					panic(err)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.statusLock.Lock()
			if status, has := fr.obStatusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.statusLock.Unlock()

			fr.tryRequestNext()
		}
		return nil
	case *p2p.StatusMessage:
		fr.statusLock.Lock()
		if status, has := fr.obStatusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		fr.statusLock.Unlock()

		fr.tryRequestNext()
		return nil
	case *p2p.TransactionMessage:
		for i, tx := range msg.Txs {
			sig := msg.Signatures[i]
			TxHash := tx.HashSig()
			if !fr.txpool.IsExist(TxHash) {
				fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
					TxHash: TxHash,
					Tx:     tx,
					Sig:    sig,
				})
			}
		}
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return errors.WithStack(p2p.ErrUnknownMessage)
	}
}

func (fr *GeneratorNode) tryRequestNext() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	TargetHeight := fr.cn.Provider().Height() + 1
	fr.Lock()
	if item, has := fr.lastGenItemMap[TargetHeight]; has && item.Recv && item.BlockGen != nil {
		fr.Unlock()
		return
	}
	fr.Unlock()

	if !fr.requestTimer.Exist(TargetHeight) {
		if fr.blockQ.Find(uint64(TargetHeight)) == nil {
			fr.statusLock.Lock()
			var TargetPubKey string
			for PubKey, status := range fr.obStatusMap {
				if TargetHeight <= status.Height {
					TargetPubKey = PubKey
					break
				}
			}
			fr.statusLock.Unlock()

			if len(TargetPubKey) > 0 {
				fr.sendRequestBlockTo(TargetPubKey, TargetHeight, 1)
			}
		}
	}
}

func (fr *GeneratorNode) updateByGenItem() {
	fr.Lock()
	defer fr.Unlock()

	TargetHeight := fr.cn.Provider().Height() + 1

	item := fr.lastGenItemMap[TargetHeight]
	for {
		if item == nil {
			return
		}
		if item.BlockGen == nil {
			return
		}
		if item.ObSign == nil {
			target := item
			var ctx *types.Context
			for target != nil && target.BlockGen != nil {
				if target.Context != nil {
					TargetHeight++
					next, has := fr.lastGenItemMap[TargetHeight]
					if has {
						ctx = target.Context.NextContext(bin.MustWriterToHash(&target.BlockGen.Block.Header), target.BlockGen.Block.Header.Timestamp)
					}
					target = next
					continue
				}
				if ctx == nil {
					ctx = fr.cn.NewContext()
				}
				sm := map[hash.Hash256]common.Address{}
				for _, tx := range item.BlockGen.Block.Body.Transactions {
					TxHash := tx.Hash(item.BlockGen.Block.Header.Height)
					item := fr.txpool.Get(TxHash)
					if item != nil {
						sm[TxHash] = item.Signer
					}
				}
				if err := fr.ct.ExecuteBlockOnContext(item.BlockGen.Block, ctx, sm); err != nil {
					log.Printf("updateByGenItem.prevItem.ExecuteBlockOnContext %+v\n", err)
					return
				}
				target.Context = ctx

				TargetHeight++
				next, has := fr.lastGenItemMap[TargetHeight]
				if has {
					ctx = target.Context.NextContext(bin.MustWriterToHash(&target.BlockGen.Block.Header), target.BlockGen.Block.Header.Timestamp)
				}
				target = next
			}
			return
		}
		// log.Println("updateByGenItem", TargetHeight, item.BlockGen != nil, item.ObSign != nil, item.Context != nil)

		b := &types.Block{
			Header: item.BlockGen.Block.Header,
			Body: types.Body{
				Transactions:          item.BlockGen.Block.Body.Transactions,
				TransactionSignatures: item.BlockGen.Block.Body.TransactionSignatures,
				Events:                item.BlockGen.Block.Body.Events,
				BlockSignatures:       append([]common.Signature{item.BlockGen.GeneratorSignature}, item.ObSign.ObserverSignatures...),
			},
		}
		if item.Context != nil {
			if err := fr.ct.ConnectBlockWithContext(b, item.Context); err != nil {
				log.Printf("updateByGenItem.ConnectBlockWithContext %+v\n", err)
				delete(fr.lastGenItemMap, b.Header.Height)
				go fr.tryRequestBlocks()
				return
			}
		} else {
			sm := map[hash.Hash256]common.Address{}
			for _, tx := range b.Body.Transactions {
				TxHash := tx.Hash(b.Header.Height)
				item := fr.txpool.Get(TxHash)
				if item != nil {
					sm[TxHash] = item.Signer
				}
			}
			if err := fr.cn.ConnectBlock(b, sm); err != nil {
				log.Printf("updateByGenItem.ConnectBlock %+v\n", err)
				delete(fr.lastGenItemMap, b.Header.Height)
				go fr.tryRequestBlocks()
				return
			}
		}
		fr.broadcastStatus()
		fr.cleanPool(b)
		if DEBUG {
			log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Body.Transactions), fr.txpool.Size())
		}
		delete(fr.lastGenItemMap, b.Header.Height)

		txs := fr.txpool.Clean(types.ToTimeSlot(b.Header.Timestamp))
		if len(txs) > 0 {
			svcs := fr.cn.Services()
			for _, s := range svcs {
				s.OnTransactionInPoolExpired(txs)
			}
			log.Println("Transaction EXPIRED", len(txs))
		}

		TargetHeight++
		item = fr.lastGenItemMap[TargetHeight]
	}
}

func (fr *GeneratorNode) genBlock(ID string, msg *BlockReqMessage) error {
	cp := fr.cn.Provider()

	RemainBlocks := prefix.MaxBlocksPerGenerator

	start := time.Now().UnixNano()
	Now := uint64(time.Now().UnixNano())
	StartBlockTime := Now
	EndBlockTime := StartBlockTime + uint64(BlockTime)*uint64(RemainBlocks)

	LastTimestamp := cp.LastTimestamp()
	if StartBlockTime < LastTimestamp {
		StartBlockTime = LastTimestamp + uint64(time.Millisecond)
	}

	if DEBUG {
		log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "BlockGenBegin", msg.TargetHeight, fr.txpool.Size())
	}

	MaxTxPerBlock := fr.Config.MaxTransactionsPerBlock
	var lastHeader *types.Header
	ctx := fr.cn.NewContext()
	failTxs := []*types.Transaction{}
	failerrs := []error{}
	for i := uint32(0); i < RemainBlocks; i++ {
		var TimeoutCount uint32
		if i == 0 {
			TimeoutCount = msg.TimeoutCount
		} else {
			ctx = ctx.NextContext(bin.MustWriterToHash(lastHeader), lastHeader.Timestamp)
		}

		Timestamp := StartBlockTime + uint64(i)*uint64(BlockTime)
		if Timestamp > EndBlockTime {
			Timestamp = EndBlockTime
		}
		if Timestamp <= ctx.LastTimestamp() {
			Timestamp = ctx.LastTimestamp() + 1
		}

		gaslv := fr.txpool.GasLevel()
		bc := chain.NewBlockCreator(fr.cn, ctx, msg.Generator, TimeoutCount, Timestamp, gaslv)
		if ctx.TargetHeight()%prefix.RewardIntervalBlocks == 0 && fr.cn.Provider().Height() < ctx.TargetHeight()-1 {
			//reward calc wait connected right block in store
			fr.Unlock()
			fr.cn.WaitConnectedBlock(ctx.TargetHeight() - 1)
			fr.Lock()
		}

		timer := time.NewTimer(400 * time.Millisecond)

		fr.txpool.Lock() // Prevent delaying from TxPool.Push
		Count := 0
		currentSlot := types.ToTimeSlot(Timestamp)
	TxLoop:
		for {
			select {
			case <-timer.C:
				break TxLoop
			default:
				item := fr.txpool.UnsafePop(currentSlot)
				if item == nil {
					break TxLoop
				}
				if err := bc.UnsafeAddTx(item.TxHash, item.Transaction, item.Signature, item.Signer); err != nil {
					if errors.Cause(err) != types.ErrUsedTimeSlot {
						fmt.Printf("UnsafeAddTx %+v\n", err)
						failTxs = append(failTxs, item.Transaction)
						failerrs = append(failerrs, err)
					}
					continue
				}
				Count++
				if Count > MaxTxPerBlock {
					break TxLoop
				}
			}
		}
		fr.txpool.Unlock() // Prevent delaying from TxPool.Push

		b, err := bc.Finalize(gaslv)
		if err != nil {
			return err
		}
		sm := &BlockGenMessage{
			Block: b,
		}
		lastHeader = &b.Header

		if sig, err := fr.key.Sign(bin.MustWriterToHash(&b.Header)); err != nil {
			return err
		} else {
			sm.GeneratorSignature = sig
		}
		fr.ms.SendTo(ID, sm)

		// log.Println("Generatorlog", fr.key.PublicKey().Address().String(), "Send.BlockGenMessage", sm.Block.Header.Height, len(sm.Block.Body.Transactions))

		fr.lastGenItemMap[sm.Block.Header.Height] = &genItem{
			BlockGen: sm,
			Context:  ctx,
		}
		fr.lastGenHeight = ctx.TargetHeight()
		fr.lastGenTime = time.Now().UnixNano()

		ExpectedTime := 200*time.Millisecond + time.Duration(i)*BlockTime
		if i == 0 {
			ExpectedTime = 200 * time.Millisecond
		} else if i >= 9 {
			ExpectedTime = BlockTime*time.Duration(i-1) + 400*time.Millisecond
		}
		PastTime := time.Duration(time.Now().UnixNano() - start)
		IsEnd := false
		fr.Unlock()

		fr.lastReqLock.Lock()
		if fr.lastReqMessage == nil {
			IsEnd = true
		}
		if ExpectedTime > PastTime {
			if !IsEnd {
				time.Sleep(ExpectedTime - PastTime)
				if fr.lastReqMessage == nil {
					IsEnd = true
				}
			}
		}
		fr.lastReqLock.Unlock()

		fr.Lock()
		if IsEnd {
			return nil
		}
	}
	services := fr.cn.Services()
	for _, s := range services {
		s.OnTransactionFail(lastHeader.Height, failTxs, failerrs)
	}

	return nil
}
