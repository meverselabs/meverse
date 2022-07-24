package node

import (
	"log"
	"math/rand"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/txpool"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

// OnConnected is called after a new  peer is connected
func (fr *GeneratorNode) OnConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.statusMap[p.ID()] = &p2p.Status{}
	fr.statusLock.Unlock()

	cp := fr.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnDisconnected is called when the  peer is disconnected
func (fr *GeneratorNode) OnDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.statusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestBlocks()
}

// OnRecv called when message received
func (fr *GeneratorNode) OnRecv(p peer.Peer, bs []byte) error {
	fr.recvChan <- &p2p.RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	return nil
}

func (fr *GeneratorNode) handlePeerMessage(ID string, m interface{}) error {
	var SenderPublicKey common.PublicKey
	copy(SenderPublicKey[:], []byte(ID))

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		//log.Println("Recv.RequestMessage", SenderPublicKey.String(), msg.Height)
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
		Height := fr.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		bs, err := p2p.BlockPacketWithCache(msg, fr.cn.Provider(), fr.batchCache, fr.singleCache)
		if err != nil {
			return err
		}
		fr.sendMessagePacket(0, SenderPublicKey, bs)
		//log.Println("Send.BlockMessage", SenderPublicKey.String(), msg.Height)
	case *p2p.StatusMessage:
		//log.Println("Recv.StatusMessage", SenderPublicKey.String(), msg.Height)
		fr.statusLock.Lock()
		if status, has := fr.statusMap[ID]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		fr.statusLock.Unlock()

		Height := fr.cn.Provider().Height()
		if Height < msg.Height {
			fr.tryRequestBlocks()
		} else {
			h, err := fr.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				log.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				fr.nm.RemovePeer(ID)
			}
		}
	case *p2p.BlockMessage:
		//log.Println("Recv.BlockMessage", SenderPublicKey.String(), msg.Blocks[0].Header.Height)
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if errors.Cause(err) == chain.ErrFoundForkedBlock {
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
		if fr.txWaitQ.Size() > 2000000 {
			return errors.WithStack(txpool.ErrTransactionPoolOverflowed)
		}
		if len(msg.Txs) > 1000 {
			return errors.WithStack(p2p.ErrTooManyTrasactionInMessage)
		}
		currentSlot := types.ToTimeSlot(fr.cn.Provider().LastTimestamp())
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
			if !fr.txpool.IsExist(TxHash) {
				fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
					TxHash: TxHash,
					Tx:     tx,
					Sig:    sig,
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
		// return errors.WithStack(p2p.ErrUnknownMessage)
	}
	return nil
}

func (fr *GeneratorNode) tryRequestBlocks() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	BaseHeight := fr.cn.Provider().Height() + 1
	for i := uint32(0); i < 10; i++ {
		TargetHeight := BaseHeight + i
		if !fr.requestTimer.Exist(TargetHeight) {
			enables := []string{}
			fr.statusLock.Lock()
			for PubKey, status := range fr.statusMap {
				if status.Height >= TargetHeight {
					enables = append(enables, PubKey)
				}
			}
			fr.statusLock.Unlock()

			if len(enables) > 0 {
				idx := rand.Intn(len(enables))
				var TargetPublicKey common.PublicKey
				copy(TargetPublicKey[:], []byte(enables[idx]))
				fr.sendRequestBlockToNode(TargetPublicKey, TargetHeight, 1)
			}
		}
	}
}
