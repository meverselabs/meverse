package pof

import (
	"time"

	"github.com/fletaio/fleta/service/p2p"
)

func (fr *FormulatorNode) broadcastStatus() error {
	cp := fr.cs.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	fr.ms.BroadcastMessage(nm)
	return nil
}

func (fr *FormulatorNode) sendRequestBlockTo(TargetID string, Height uint32) error {
	nm := &p2p.RequestMessage{
		Height: Height,
	}
	fr.ms.SendTo(TargetID, nm)
	fr.requestTimer.Add(Height, 10*time.Second, TargetID)
	return nil
}
