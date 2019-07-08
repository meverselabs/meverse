package pof

import (
	"time"

	"github.com/fletaio/fleta/service/p2p"

	"github.com/fletaio/fleta/common"
)

func (fr *Formulator) broadcastStatus() error {
	cp := fr.cs.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	fr.ms.BroadcastMessage(nm)
	return nil
}

func (fr *Formulator) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32) error {
	nm := &p2p.RequestMessage{
		Height: Height,
	}
	fr.ms.SendTo(TargetPubHash, nm)
	fr.requestTimer.Add(Height, 10*time.Second, TargetPubHash)
	return nil
}
