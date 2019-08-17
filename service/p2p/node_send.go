package p2p

import (
	"log"
	"time"

	"github.com/fletaio/fleta/common"
)

func (nd *Node) sendStatusTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == nd.myPublicHash {
		return nil
	}

	cp := nd.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	nd.ms.SendTo(TargetPubHash, nm)
	return nil
}

func (nd *Node) broadcastStatus() error {
	cp := nd.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	nd.ms.BroadcastMessage(nm)

	log.Println("broadcastStatus", height)
	return nil
}

func (nd *Node) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32, Count uint8) error {
	if TargetPubHash == nd.myPublicHash {
		return nil
	}
	log.Println("sendRequestBlockTo", TargetPubHash.String(), Height, Count)

	nm := &RequestMessage{
		Height: Height,
		Count:  Count,
	}
	nd.ms.SendTo(TargetPubHash, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		nd.requestTimer.Add(Height+i, 10*time.Second, string(TargetPubHash[:]))
	}
	return nil
}
