package node

import (
	"sort"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
)

func (ob *ObserverNode) sendMessage(Priority int, Address common.Address, m p2p.Serializable) {
	ob.sendChan <- &p2p.SendMessageItem{
		Address: Address,
		Packet:  p2p.MessageToPacket(m),
	}
}

func (ob *ObserverNode) sendMessagePacket(Priority int, Address common.Address, bs []byte) {
	ob.sendChan <- &p2p.SendMessageItem{
		Address: Address,
		Packet:  bs,
	}
}

func (ob *ObserverNode) hasBlockVoteHalf(TargetHeight uint32) (bool, *BlockVoteMessage) {
	if ob.round.TargetHeight == TargetHeight {
		if br, has := ob.round.BlockRoundMap[ob.round.TargetHeight]; has {
			mp := map[common.PublicKey]*BlockVoteMessage{}
			for k, v := range br.BlockVoteMessageWaitMap {
				mp[k] = v
			}
			for k, v := range br.BlockVoteMap {
				mp[k] = v
			}
			if len(mp) >= len(ob.observerKeyMap)/2 {
				for _, v := range mp {
					return true, v
				}
			}
		}
	}
	return false, nil
}

func (ob *ObserverNode) createRoundVoteMessage() (*RoundVoteMessage, error) {
	cp := ob.cn.Provider()
	height := cp.Height()
	lastHash := cp.LastHash()

	var nm *RoundVoteMessage
	if has, v := ob.hasBlockVoteHalf(height + 1); has {
		nm = &RoundVoteMessage{
			ChainID:      cp.ChainID(),
			LastHash:     lastHash,
			TargetHeight: height + 1,
			TimeoutCount: uint32(v.Header.TimeoutCount),
			Generator:    v.Header.Generator,
			PublicKey:    ob.myPublicKey,
			Timestamp:    uint64(time.Now().UnixNano()),
			IsReply:      false,
		}
	}

	if nm == nil {
		Top, TimeoutCount, err := ob.cn.TopGeneratorInMap(ob.adjustGeneratorMap())
		if err != nil {
			return nil, err
		}

		nm = &RoundVoteMessage{
			ChainID:      cp.ChainID(),
			LastHash:     lastHash,
			TargetHeight: height + 1,
			TimeoutCount: uint32(TimeoutCount),
			Generator:    Top,
			PublicKey:    ob.myPublicKey,
			Timestamp:    uint64(time.Now().UnixNano()),
			IsReply:      false,
		}
	}

	return nm, nil
}

func (ob *ObserverNode) sendRoundVote() error {
	nm, err := ob.createRoundVoteMessage()
	if err != nil {
		return err
	}

	ob.statusLock.Lock()
	status, has := ob.statusMap[string(nm.Generator[:])]
	ob.statusLock.Unlock()

	cp := ob.cn.Provider()
	height := cp.Height()
	if !has || status.Height < height {
		lastHash := cp.LastHash()
		ob.sendMessage(1, nm.Generator, &p2p.StatusMessage{
			Version:  cp.Version(),
			Height:   height,
			LastHash: lastHash,
		})
	}
	ob.round.VoteFailCount = 0

	ob.messageQueue.Push(&messageItem{
		PublicKey: ob.myPublicKey,
		Message:   nm,
	})

	ob.ms.BroadcastPacket(p2p.MessageToPacket(nm))
	return nil
}

func (ob *ObserverNode) sendRoundVoteTo(TargetPubKey common.PublicKey) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}

	cp := ob.cn.Provider()
	height := cp.Height()
	lastHash := cp.LastHash()
	if ob.round.RoundState == BlockVoteState {
		MyMsg, has := ob.round.RoundVoteMessageMap[ob.myPublicKey]
		if !has {
			return nil
		}
		nm := &RoundVoteMessage{
			ChainID:      MyMsg.ChainID,
			LastHash:     MyMsg.LastHash,
			TargetHeight: MyMsg.TargetHeight,
			TimeoutCount: MyMsg.TimeoutCount,
			Generator:    MyMsg.Generator,
			PublicKey:    MyMsg.PublicKey,
			Timestamp:    MyMsg.Timestamp,
			IsReply:      true,
		}

		TargetHeight := height + 1
		if MyMsg.TargetHeight != TargetHeight {
			nm.TimeoutCount = 0
			nm.TargetHeight = TargetHeight
			nm.LastHash = lastHash
			nm.Timestamp = uint64(time.Now().UnixNano())
		}

		ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	} else {
		nm, err := ob.createRoundVoteMessage()
		if err != nil {
			return err
		}
		ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	}
	return nil
}

func (ob *ObserverNode) sendRoundVoteAck() error {
	votes := []*voteSortItem{}
	for PubKey, v := range ob.round.RoundVoteMessageMap {
		votes = append(votes, &voteSortItem{
			PublicKey: PubKey,
			Priority:  uint64(v.TimeoutCount),
		})
	}
	sort.Sort(voteSorter(votes))

	MinPublicKey := votes[0].PublicKey
	MinRoundVote := ob.round.RoundVoteMessageMap[MinPublicKey]
	nm := &RoundVoteAckMessage{
		ChainID:      MinRoundVote.ChainID,
		LastHash:     MinRoundVote.LastHash,
		TargetHeight: MinRoundVote.TargetHeight,
		TimeoutCount: MinRoundVote.TimeoutCount,
		Generator:    MinRoundVote.Generator,
		PublicKey:    MinRoundVote.PublicKey,
		Timestamp:    uint64(time.Now().UnixNano()),
		IsReply:      false,
	}

	ob.messageQueue.Push(&messageItem{
		PublicKey: ob.myPublicKey,
		Message:   nm,
	})

	ob.ms.BroadcastPacket(p2p.MessageToPacket(nm))
	return nil
}

func (ob *ObserverNode) sendRoundVoteAckTo(TargetPubKey common.PublicKey) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}

	if ob.round.RoundState == BlockVoteState {
		MyMsg, has := ob.round.RoundVoteAckMessageMap[ob.myPublicKey]
		if !has {
			return nil
		}
		nm := &RoundVoteAckMessage{
			ChainID:      MyMsg.ChainID,
			LastHash:     MyMsg.LastHash,
			TargetHeight: MyMsg.TargetHeight,
			TimeoutCount: MyMsg.TimeoutCount,
			Generator:    MyMsg.Generator,
			PublicKey:    MyMsg.PublicKey,
			Timestamp:    MyMsg.Timestamp,
			IsReply:      true,
		}
		cp := ob.cn.Provider()
		height := cp.Height()
		lastHash := cp.LastHash()
		TargetHeight := height + 1
		if MyMsg.TargetHeight != TargetHeight {
			nm.TimeoutCount = 0
			nm.TargetHeight = TargetHeight
			nm.LastHash = lastHash
			nm.Timestamp = uint64(time.Now().UnixNano())
		}

		ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	} else {
		MyMsg, has := ob.round.RoundVoteAckMessageMap[ob.myPublicKey]
		if !has {
			return nil
		}
		nm := &RoundVoteAckMessage{
			ChainID:      MyMsg.ChainID,
			LastHash:     MyMsg.LastHash,
			TargetHeight: MyMsg.TargetHeight,
			TimeoutCount: MyMsg.TimeoutCount,
			Generator:    MyMsg.Generator,
			PublicKey:    MyMsg.PublicKey,
			Timestamp:    MyMsg.Timestamp,
			IsReply:      true,
		}

		ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	}
	return nil
}

func (ob *ObserverNode) sendBlockVote(gen *BlockGenMessage) error {
	nm := &BlockVoteMessage{
		TargetHeight:       gen.Block.Header.Height,
		Header:             &gen.Block.Header,
		GeneratorSignature: gen.GeneratorSignature,
		IsReply:            false,
	}

	s := &types.BlockSign{
		HeaderHash:         bin.MustWriterToHash(&gen.Block.Header),
		GeneratorSignature: gen.GeneratorSignature,
	}
	if sig, err := ob.key.Sign(bin.MustWriterToHash(s)); err != nil {
		return err
	} else {
		nm.ObserverSignature = sig
	}

	ob.messageQueue.Push(&messageItem{
		PublicKey: ob.myPublicKey,
		Message:   nm,
	})

	ob.ms.BroadcastPacket(p2p.MessageToPacket(nm))
	return nil
}

func (ob *ObserverNode) sendBlockGenTo(gen *BlockGenMessage, TargetPubKey common.PublicKey) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}
	ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(gen))
	return nil
}

func (ob *ObserverNode) sendBlockVoteTo(gen *BlockGenMessage, TargetPubKey common.PublicKey) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}

	nm := &BlockVoteMessage{
		TargetHeight:       gen.Block.Header.Height,
		Header:             &gen.Block.Header,
		GeneratorSignature: gen.GeneratorSignature,
		IsReply:            true,
	}

	s := &types.BlockSign{
		HeaderHash:         bin.MustWriterToHash(&gen.Block.Header),
		GeneratorSignature: gen.GeneratorSignature,
	}
	if sig, err := ob.key.Sign(bin.MustWriterToHash(s)); err != nil {
		return err
	} else {
		nm.ObserverSignature = sig
	}

	ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	return nil
}

func (ob *ObserverNode) sendBlockGenRequest(br *BlockRound) error {
	now := uint64(time.Now().UnixNano())
	if br.LastBlockGenRequestTime+uint64(1*time.Second) > now {
		return nil
	}
	br.LastBlockGenRequestTime = now

	var PublicKey common.PublicKey
	has := false
	for PubKey := range br.BlockVoteMap {
		PublicKey = PubKey
		has = true
		break
	}
	if !has {
		for PubKey := range br.BlockVoteMessageWaitMap {
			PublicKey = PubKey
			has = true
			break
		}
	}
	nm := &BlockGenRequestMessage{
		ChainID:      ob.round.MinRoundVoteAck.ChainID,
		LastHash:     ob.round.MinRoundVoteAck.LastHash,
		TargetHeight: ob.round.MinRoundVoteAck.TargetHeight,
		TimeoutCount: ob.round.MinRoundVoteAck.TimeoutCount,
		Generator:    ob.round.MinRoundVoteAck.Generator,
		Timestamp:    uint64(time.Now().UnixNano()),
	}
	if has {
		ob.ms.SendTo(PublicKey, p2p.MessageToPacket(nm))
	} else {
		ob.ms.SendAnyone(p2p.MessageToPacket(nm))
	}
	return nil
}

func (ob *ObserverNode) sendStatusTo(TargetPubKey common.PublicKey) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}

	cp := ob.cn.Provider()
	height := cp.Height()
	lastHash := cp.LastHash()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	nm = nil
	return nil
}

func (ob *ObserverNode) broadcastStatus() error {
	cp := ob.cn.Provider()
	height := cp.Height()
	lastHash := cp.LastHash()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.BroadcastPacket(p2p.MessageToPacket(nm))
	return nil
}

func (ob *ObserverNode) sendRequestBlockTo(TargetPubKey common.PublicKey, Height uint32, Count uint8) error {
	if TargetPubKey == ob.myPublicKey {
		return nil
	}

	nm := &p2p.RequestMessage{
		Height: Height,
		Count:  Count,
	}
	ob.ms.SendTo(TargetPubKey, p2p.MessageToPacket(nm))
	for i := uint32(0); i < uint32(Count); i++ {
		ob.requestTimer.Add(Height+i, 2*time.Second, string(TargetPubKey[:]))
	}
	return nil
}
