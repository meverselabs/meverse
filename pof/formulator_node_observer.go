package pof

import (
	"bytes"
	"log"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// OnObserverConnected is called after a new observer peer is connected
func (fr *FormulatorNode) OnObserverConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.obStatusMap[p.ID()] = &p2p.Status{}
	fr.statusLock.Unlock()

	cp := fr.cs.cn.Provider()
	height, lastHash := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnObserverDisconnected is called when the observer peer is disconnected
func (fr *FormulatorNode) OnObserverDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.obStatusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestNext()
}

func (fr *FormulatorNode) onObserverRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}

	if err := fr.handleObserverMessage(p, m, 0); err != nil {
		//rlog.Println(err)
		return nil
	}
	return nil
}

func (fr *FormulatorNode) handleObserverMessage(p peer.Peer, m interface{}, RetryCount int) error {
	cp := fr.cs.cn.Provider()

	switch msg := m.(type) {
	case *BlockReqMessage:
		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockReqMessage", msg.TargetHeight)

		fr.Lock()
		defer fr.Unlock()

		TargetHeight := fr.cs.cn.Provider().Height() + 1
		if msg.TargetHeight < TargetHeight {
			return nil
		}
		if msg.TargetHeight <= fr.lastGenHeight {
			if time.Now().UnixNano() < fr.lastGenTime+int64(30*time.Second) {
				return nil
			}
			fr.lastReqMessage = nil
		}
		if fr.lastReqMessage != nil {
			if msg.TargetHeight <= fr.lastReqMessage.TargetHeight {
				return nil
			}
		}
		if msg.TargetHeight > TargetHeight {
			if msg.TargetHeight > TargetHeight+10 {
				return nil
			}
			if RetryCount >= 10 {
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
				fr.handleObserverMessage(p, m, RetryCount+1)
			}()
			return nil
		}

		if msg.Formulator != fr.Config.Formulator {
			return ErrInvalidRequest
		}
		if msg.FormulatorPublicHash != fr.frPublicHash {
			return ErrInvalidRequest
		}
		if msg.PrevHash != cp.LastHash() {
			return ErrInvalidRequest
		}
		Top, err := fr.cs.rt.TopRank(int(msg.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.Formulator != Top.Address {
			return ErrInvalidRequest
		}
		fr.lastReqMessage = msg

		go func(req *BlockReqMessage) error {
			fr.genLock.Lock()
			defer fr.genLock.Unlock()

			fr.Lock()
			defer fr.Unlock()

			return fr.genBlock(p, req)
		}(msg)
		return nil
	case *BlockGenMessage:
		rlog.Println("Formulator", fr.Config.Formulator.String(), "Recv.BlockGenMessage", msg.Block.Header.Height)

		TargetHeight := fr.cs.cn.Provider().Height() + 1
		if msg.Block.Header.Height < TargetHeight {
			return nil
		}
		if msg.Block.Header.Generator != fr.Config.Formulator {
			fr.lastReqMessage = nil
		}
		fr.Lock()
		item, has := fr.lastGenItemMap[msg.Block.Header.Height]
		if has {
			if item.ObSign != nil {
				if item.ObSign.BlockSign.HeaderHash != encoding.Hash(msg.Block.Header) {
					return ErrInvalidRequest
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
		fr.Unlock()

		go fr.updateByGenItem()
		return nil
	case *BlockObSignMessage:
		rlog.Println("Formulator", fr.Config.Formulator.String(), "Recv.BlockObSignMessage", msg.TargetHeight)

		TargetHeight := fr.cs.cn.Provider().Height() + 1
		if msg.TargetHeight < TargetHeight {
			return nil
		}

		fr.Lock()
		if item, has := fr.lastGenItemMap[msg.TargetHeight]; has {
			if item.BlockGen != nil {
				if msg.BlockSign.HeaderHash != encoding.Hash(item.BlockGen.Block.Header) {
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
		log.Println("Recv.Ob.BlockMessage", msg.Blocks[0].Header.Height)
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
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
		TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), msg.TxType, msg.Tx)
		fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
			TxHash:  TxHash,
			Message: msg,
		})
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
}

func (fr *FormulatorNode) tryRequestNext() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	TargetHeight := fr.cs.cn.Provider().Height() + 1
	if item, has := fr.lastGenItemMap[TargetHeight]; has && item.Recv && item.BlockGen != nil {
		return
	}

	if !fr.requestTimer.Exist(TargetHeight) {
		if fr.blockQ.Find(uint64(TargetHeight)) == nil {
			fr.statusLock.Lock()
			var TargetPubHash string
			for pubhash, status := range fr.obStatusMap {
				if TargetHeight <= status.Height {
					TargetPubHash = pubhash
					break
				}
			}
			fr.statusLock.Unlock()

			if len(TargetPubHash) > 0 {
				fr.sendRequestBlockTo(TargetPubHash, TargetHeight, 1)
			}
		}
	}
}

func (fr *FormulatorNode) updateByGenItem() {
	fr.Lock()
	defer fr.Unlock()

	TargetHeight := fr.cs.cn.Provider().Height() + 1

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
						ctx = target.Context.NextContext(encoding.Hash(target.BlockGen.Block.Header), target.BlockGen.Block.Header.Timestamp)
					}
					target = next
					continue
				}
				if ctx == nil {
					ctx = fr.cs.cn.NewContext()
				}
				if err := fr.cs.ct.ExecuteBlockOnContext(item.BlockGen.Block, ctx); err != nil {
					log.Println("updateByGenItem.prevItem.ConnectBlockWithContext", err)
					return
				}
				target.Context = ctx

				TargetHeight++
				next, has := fr.lastGenItemMap[TargetHeight]
				if has {
					ctx = target.Context.NextContext(encoding.Hash(target.BlockGen.Block.Header), target.BlockGen.Block.Header.Timestamp)
				}
				target = next
			}
			return
		}
		log.Println("updateByGenItem", TargetHeight, item.BlockGen != nil, item.ObSign != nil, item.Context != nil)

		b := &types.Block{
			Header:                item.BlockGen.Block.Header,
			TransactionTypes:      item.BlockGen.Block.TransactionTypes,
			Transactions:          item.BlockGen.Block.Transactions,
			TransactionSignatures: item.BlockGen.Block.TransactionSignatures,
			TransactionResults:    item.BlockGen.Block.TransactionResults,
			Signatures:            append([]common.Signature{item.BlockGen.GeneratorSignature}, item.ObSign.ObserverSignatures...),
		}
		if item.Context != nil {
			if err := fr.cs.ct.ConnectBlockWithContext(b, item.Context); err != nil {
				log.Println("updateByGenItem.ConnectBlockWithContext", err)
				delete(fr.lastGenItemMap, b.Header.Height)
				return
			}
		} else {
			if err := fr.cs.cn.ConnectBlock(b); err != nil {
				log.Println("updateByGenItem.ConnectBlock", err)
				delete(fr.lastGenItemMap, b.Header.Height)
				return
			}
		}
		fr.broadcastStatus()
		fr.cleanPool(b)
		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))
		delete(fr.lastGenItemMap, b.Header.Height)

		TargetHeight++
		item = fr.lastGenItemMap[TargetHeight]
	}
}

func (fr *FormulatorNode) genBlock(p peer.Peer, msg *BlockReqMessage) error {
	cp := fr.cs.cn.Provider()

	RemainBlocks := fr.cs.maxBlocksPerFormulator
	if msg.TimeoutCount == 0 {
		RemainBlocks = fr.cs.maxBlocksPerFormulator - fr.cs.blocksBySameFormulator
	}

	start := time.Now().UnixNano()
	Now := uint64(time.Now().UnixNano())
	StartBlockTime := Now
	EndBlockTime := StartBlockTime + uint64(500*time.Millisecond)*uint64(RemainBlocks)

	LastTimestamp := cp.LastTimestamp()
	if StartBlockTime < LastTimestamp {
		StartBlockTime = LastTimestamp + uint64(time.Millisecond)
	}

	var lastHeader *types.Header
	ctx := fr.cs.cn.NewContext()
	for i := uint32(0); i < RemainBlocks; i++ {
		var TimeoutCount uint32
		if i == 0 {
			TimeoutCount = msg.TimeoutCount
		} else {
			ctx = ctx.NextContext(encoding.Hash(lastHeader), lastHeader.Timestamp)
		}

		Timestamp := StartBlockTime + uint64(i)*uint64(500*time.Millisecond)
		if Timestamp > EndBlockTime {
			Timestamp = EndBlockTime
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

		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockGenBegin", msg.TargetHeight)

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
					rlog.Println("UnsafeAddTx", err)
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

		sm := &BlockGenMessage{
			Block: b,
		}
		lastHeader = &b.Header

		if sig, err := fr.key.Sign(encoding.Hash(b.Header)); err != nil {
			return err
		} else {
			sm.GeneratorSignature = sig
		}
		p.SendPacket(p2p.MessageToPacket(sm))

		rlog.Println("Formulator", fr.Config.Formulator.String(), "Send.BlockGenMessage", sm.Block.Header.Height, len(sm.Block.Transactions))

		fr.lastGenItemMap[sm.Block.Header.Height] = &genItem{
			BlockGen: sm,
			Context:  ctx,
		}
		fr.lastGenHeight = ctx.TargetHeight()
		fr.lastGenTime = time.Now().UnixNano()

		ExpectedTime := 200*time.Millisecond + time.Duration(i)*500*time.Millisecond
		if i == 0 {
			ExpectedTime = 200 * time.Millisecond
		} else if i >= 9 {
			ExpectedTime = 4200*time.Millisecond + time.Duration(i-9+1)*200*time.Millisecond
		}
		PastTime := time.Duration(time.Now().UnixNano() - start)
		if ExpectedTime > PastTime {
			IsEnd := false
			fr.Unlock()
			if fr.lastReqMessage == nil {
				IsEnd = true
			}
			if !IsEnd {
				time.Sleep(ExpectedTime - PastTime)
				if fr.lastReqMessage == nil {
					IsEnd = true
				}
			}
			fr.Lock()
			if IsEnd {
				return nil
			}
		} else {
			IsEnd := false
			fr.Unlock()
			if fr.lastReqMessage == nil {
				IsEnd = true
			}
			fr.Lock()
			if IsEnd {
				return nil
			}
		}
	}
	return nil
}
