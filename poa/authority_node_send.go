package poa

import (
	"github.com/fletaio/fleta/service/p2p"
)

func (an *AuthorityNode) broadcastStatus() error {
	cp := an.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	an.ms.BroadcastMessage(nm)
	return nil
}
