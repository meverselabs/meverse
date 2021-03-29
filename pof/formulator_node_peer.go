package pof

import (
	"math/rand"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// OnConnected is called after a new  peer is connected
func (fr *FormulatorNode) OnConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.statusMap[p.ID()] = &p2p.Status{}
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

// OnDisconnected is called when the  peer is disconnected
func (fr *FormulatorNode) OnDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.statusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestBlocks()
}

// OnRecv called when message received
func (fr *FormulatorNode) OnRecv(p peer.Peer, bs []byte) error {
	fr.recvChan <- &p2p.RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	return nil
}

func (fr *FormulatorNode) handlePeerMessage(ID string, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(ID))

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		//log.Println("Recv.RequestMessage", SenderPublicHash.String(), msg.Height)
		fr.statusLock.Lock()
		status, has := fr.statusMap[ID]
		fr.statusLock.Unlock()

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
		Height := fr.cs.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		bs, err := p2p.BlockPacketWithCache(msg, fr.cs.cn.Provider(), fr.batchCache, fr.singleCache)
		if err != nil {
			return err
		}
		fr.sendMessagePacket(0, SenderPublicHash, bs)
		//log.Println("Send.BlockMessage", SenderPublicHash.String(), msg.Height)
	case *p2p.StatusMessage:
		//log.Println("Recv.StatusMessage", SenderPublicHash.String(), msg.Height)
		fr.statusLock.Lock()
		if status, has := fr.statusMap[ID]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		fr.statusLock.Unlock()

		Height := fr.cs.cn.Provider().Height()
		if Height < msg.Height {
			fr.tryRequestBlocks()
		} else {
			h, err := fr.cs.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				fr.nm.RemovePeer(ID)
			}
		}
	case *p2p.BlockMessage:
		//log.Println("Recv.BlockMessage", SenderPublicHash.String(), msg.Blocks[0].Header.Height)
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					fr.nm.RemovePeer(ID)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.statusLock.Lock()
			if status, has := fr.statusMap[ID]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.statusLock.Unlock()
		}
	case *p2p.TransactionMessage:
		//log.Println("Recv.TransactionMessage", fr.txWaitQ.Size(), fr.txpool.Size())
		/*
			if fr.txWaitQ.Size() > 200000 {
				return txpool.ErrTransactionPoolOverflowed
			}
		*/
		if len(msg.Types) > 800 {
			return p2p.ErrTooManyTrasactionInMessage
		}
		ChainID := fr.cs.cn.Provider().ChainID()
		currentSlot := types.ToTimeSlot(fr.cs.cn.Provider().LastTimestamp())
		for i, t := range msg.Types {
			tx := msg.Txs[i]
			slot := types.ToTimeSlot(tx.Timestamp())
			if currentSlot > 0 {
				if slot < currentSlot-1 {
					continue
				} else if slot > currentSlot+10 {
					continue
				}
			}
			sigs := msg.Signatures[i]
			TxHash := chain.HashTransactionByType(ChainID, t, tx)
			if !fr.txpool.IsExist(TxHash) {
				fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
					TxHash: TxHash,
					Type:   t,
					Tx:     tx,
					Sigs:   sigs,
					PeerID: ID,
				})
			}
		}
		return nil
	case *p2p.PeerListMessage:
		fr.nm.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *p2p.RequestPeerListMessage:
		fr.nm.SendPeerList(ID)
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

	BaseHeight := fr.cs.cn.Provider().Height() + 1
	for i := uint32(0); i < 10; i++ {
		TargetHeight := BaseHeight + i
		if !fr.requestTimer.Exist(TargetHeight) {
			enables := []string{}
			fr.statusLock.Lock()
			for pubhash, status := range fr.statusMap {
				if status.Height >= TargetHeight {
					enables = append(enables, pubhash)
				}
			}
			fr.statusLock.Unlock()

			if len(enables) > 0 {
				idx := rand.Intn(len(enables))
				var TargetPublicHash common.PublicHash
				copy(TargetPublicHash[:], []byte(enables[idx]))
				fr.sendRequestBlockToNode(TargetPublicHash, TargetHeight, 1)
			}
		}
	}
}
