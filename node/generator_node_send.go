package node

import (
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/p2p"
)

func (fr *GeneratorNode) sendMessage(Priority int, Target common.PublicKey, m p2p.Serializable) {
	fr.sendChan <- &p2p.SendMessageItem{
		Target: Target,
		Packet: p2p.MessageToPacket(m),
	}
}

func (fr *GeneratorNode) sendMessagePacket(Priority int, Target common.PublicKey, bs []byte) {
	fr.sendChan <- &p2p.SendMessageItem{
		Target: Target,
		Packet: bs,
	}
}

func (fr *GeneratorNode) broadcastMessage(Priority int, m p2p.Serializable) {
	fr.sendChan <- &p2p.SendMessageItem{
		Packet: p2p.MessageToPacket(m),
	}
}

func (fr *GeneratorNode) broadcastMessagePacket(Priority int, bs []byte) {
	fr.sendChan <- &p2p.SendMessageItem{
		Packet: bs,
	}
}

func (fr *GeneratorNode) exceptCastMessage(Priority int, Target common.PublicKey, m p2p.Serializable) {
	fr.sendChan <- &p2p.SendMessageItem{
		Target: Target,
		Packet: p2p.MessageToPacket(m),
		Except: true,
	}
}

func (fr *GeneratorNode) sendStatusTo(TargetPubKey common.PublicKey) error {
	if TargetPubKey == fr.myPublicKey {
		return nil
	}

	cp := fr.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	fr.sendMessage(0, TargetPubKey, nm)
	return nil
}

func (fr *GeneratorNode) broadcastStatus() error {
	cp := fr.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	bs := p2p.MessageToPacket(nm)
	fr.ms.BroadcastPacket(bs)
	fr.broadcastMessagePacket(0, bs)
	return nil
}

func (fr *GeneratorNode) sendRequestBlockTo(TargetID string, Height uint32, Count uint8) error {
	//log.Println("sendRequestBlockTo", Height, Count)

	nm := &p2p.RequestMessage{
		Height: Height,
		Count:  Count,
	}
	fr.ms.SendTo(TargetID, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		fr.requestTimer.Add(Height+i, 2*time.Second, TargetID)
	}
	return nil
}

func (fr *GeneratorNode) sendRequestBlockToNode(TargetPubKey common.PublicKey, Height uint32, Count uint8) error {
	if TargetPubKey == fr.myPublicKey {
		return nil
	}
	//log.Println("sendRequestBlockToNode", TargetPubKey.String(), Height, Count)

	nm := &p2p.RequestMessage{
		Height: Height,
		Count:  Count,
	}
	fr.sendMessage(0, TargetPubKey, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		fr.requestTimer.Add(Height+i, 2*time.Second, string(TargetPubKey[:]))
	}
	return nil
}
