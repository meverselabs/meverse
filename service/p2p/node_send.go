package p2p

import (
	"time"

	"github.com/fletaio/fleta/common"
)

func (nd *Node) sendMessage(Priority int, Target common.PublicHash, m interface{}) {
	if _, is := m.([]byte); is {
		panic("")
	}

	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: MessageToPacket(m),
	}
}

func (nd *Node) sendMessagePacket(Priority int, Target common.PublicHash, bs []byte) {
	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: bs,
	}
}

func (nd *Node) broadcastMessage(Priority int, m interface{}) {
	if _, is := m.([]byte); is {
		panic("")
	}

	nd.sendChan <- &SendMessageItem{
		Packet: MessageToPacket(m),
	}
}

func (nd *Node) exceptCastMessage(Priority int, Target common.PublicHash, m interface{}) {
	if _, is := m.([]byte); is {
		panic("")
	}

	nd.sendChan <- &SendMessageItem{
		Target: Target,
		Packet: MessageToPacket(m),
		Except: true,
	}
}

func (nd *Node) sendStatusTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == nd.myPublicHash {
		return nil
	}

	cp := nd.cn.Provider()
	height, lastHash := cp.LastStatus()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	nd.sendMessage(0, TargetPubHash, nm)
	return nil
}

func (nd *Node) broadcastStatus() error {
	cp := nd.cn.Provider()
	height, lastHash := cp.LastStatus()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	nd.ms.BroadcastPacket(MessageToPacket(nm))
	return nil
}

func (nd *Node) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32, Count uint8) error {
	if TargetPubHash == nd.myPublicHash {
		return nil
	}

	nm := &RequestMessage{
		Height: Height,
		Count:  Count,
	}
	nd.sendMessage(0, TargetPubHash, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		nd.requestTimer.Add(Height+i, 2*time.Second, string(TargetPubHash[:]))
	}
	return nil
}
