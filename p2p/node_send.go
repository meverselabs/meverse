package p2p

import (
	"time"

	"github.com/meverselabs/meverse/common"
)

func (nd *Node) sendMessage(Priority int, Target common.PublicKey, m Serializable) {
	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: MessageToPacket(m),
	}
}

func (nd *Node) sendMessagePacket(Priority int, Target common.PublicKey, bs []byte) {
	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: bs,
	}
}

func (nd *Node) broadcastMessage(Priority int, m Serializable) {
	nd.sendChan <- &SendMessageItem{
		Packet: MessageToPacket(m),
	}
}

func (nd *Node) exceptCastMessage(Priority int, Target common.PublicKey, m Serializable) {
	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: MessageToPacket(m),
		Except: true,
	}
}

func (nd *Node) sendStatusTo(TargetPublicKey common.PublicKey) error {
	if TargetPublicKey == nd.myPublicKey {
		return nil
	}

	cp := nd.cn.Provider()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	nd.sendMessage(0, TargetPublicKey, nm)
	return nil
}

func (nd *Node) broadcastStatus() error {
	cp := nd.cn.Provider()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	nd.ms.BroadcastPacket(MessageToPacket(nm))
	return nil
}

func (nd *Node) sendRequestBlockTo(TargetPublicKey common.PublicKey, Height uint32, Count uint8) error {
	if TargetPublicKey == nd.myPublicKey {
		return nil
	}

	nm := &RequestMessage{
		Height: Height,
		Count:  Count,
	}
	nd.sendMessage(0, TargetPublicKey, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		nd.requestTimer.Add(Height+i, 2*time.Second, string(TargetPublicKey[:]))
	}
	return nil
}
