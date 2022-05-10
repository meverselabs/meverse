package node

import (
	"log"

	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

// OnGeneratorConnected is called after a new generator peer is connected
func (ob *ObserverNode) OnGeneratorConnected(p peer.Peer) {
	ob.statusLock.Lock()
	ob.statusMap[p.ID()] = &p2p.Status{}
	ob.statusLock.Unlock()

	cp := ob.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnGeneratorDisconnected is called when the generator peer is disconnected
func (ob *ObserverNode) OnGeneratorDisconnected(p peer.Peer) {
	ob.statusLock.Lock()
	delete(ob.statusMap, p.ID())
	ob.statusLock.Unlock()
}

func (ob *ObserverNode) onGeneratorRecv(p peer.Peer, bs []byte) error {
	item := &p2p.RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	t := p2p.PacketMessageType(bs)
	switch t {
	case BlockGenMessageType:
		m, err := p2p.PacketToMessage(bs)
		if err != nil {
			log.Printf("PacketToMessage %+v\n", err)
			ob.fs.RemovePeer(item.PeerID)
			break
		}
		ob.messageQueue.Push(&messageItem{
			Message: m,
			Packet:  bs,
		})
	case p2p.RequestMessageType:
		ob.recvChan <- item
	case p2p.StatusMessageType:
		ob.recvChan <- item
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return errors.WithStack(p2p.ErrUnknownMessage)
	}
	return nil
}

func (ob *ObserverNode) handleGeneratorMessage(p peer.Peer, m interface{}, bs []byte) error {
	cp := ob.cn.Provider()
	switch msg := m.(type) {
	case *p2p.RequestMessage:
		ob.statusLock.Lock()
		status, has := ob.statusMap[p.ID()]
		ob.statusLock.Unlock()
		if has {
			if msg.Height < status.Height {
				if msg.Height+uint32(msg.Count) <= status.Height {
					return nil
				}
				msg.Height = status.Height
			}
		}

		enable := false
		hasCount := 0
		ob.statusLock.Lock()
		for _, status := range ob.statusMap {
			if status.Height >= msg.Height {
				hasCount++
				if hasCount >= 3 {
					break
				}
			}
		}
		ob.statusLock.Unlock()

		// TODO : it is top leader, only allow top
		// TODO : it is next leader, only allow next
		// TODO : it is not leader, accept 3rd-5th
		if hasCount < 3 {
			enable = true
		} else {
			ob.Lock()
			ranks, err := ob.cn.GeneratorsInMap(ob.adjustGeneratorMap(), 5)
			ob.Unlock()
			if err != nil {
				return err
			}
			rankMap := map[string]bool{}
			for _, r := range ranks {
				rankMap[string(r[:])] = true
			}
			enable = rankMap[p.ID()]
		}
		if enable {
			if msg.Count == 0 {
				msg.Count = 1
			}
			if msg.Count > 10 {
				msg.Count = 10
			}
			Height := ob.cn.Provider().Height()
			if msg.Height > Height {
				return nil
			}
			bs, err := p2p.BlockPacketWithCache(msg, ob.cn.Provider(), ob.batchCache, ob.singleCache)
			if err != nil {
				return err
			}
			p.SendPacket(bs)
		}
	case *p2p.StatusMessage:
		ob.statusLock.Lock()
		if status, has := ob.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		ob.statusLock.Unlock()

		Height := cp.Height()
		if Height >= msg.Height {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				log.Println(ob.obID, chain.ErrFoundForkedBlock, p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				ob.fs.RemovePeer(p.ID())
			}
		}
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return errors.WithStack(p2p.ErrUnknownMessage)
	}
	return nil
}
