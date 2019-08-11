package pof

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

func (ob *ObserverNode) sendRoundVote() error {
	Top, TimeoutCount, err := ob.cs.rt.TopRankInMap(ob.adjustFormulatorMap())
	if err != nil {
		return err
	}

	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &RoundVoteMessage{
		RoundVote: &RoundVote{
			ChainID:              cp.ChainID(),
			LastHash:             lastHash,
			TargetHeight:         height + 1,
			TimeoutCount:         uint32(TimeoutCount),
			Formulator:           Top.Address,
			FormulatorPublicHash: Top.PublicHash,
			Timestamp:            uint64(time.Now().UnixNano()),
			IsReply:              false,
		},
	}
	if gh, err := ob.fs.GuessHeight(Top.Address); err != nil {
		ob.fs.SendTo(Top.Address, &p2p.StatusMessage{
			Version:  cp.Version(),
			Height:   height,
			LastHash: lastHash,
		})
	} else if gh < height {
		ob.fs.SendTo(Top.Address, &p2p.StatusMessage{
			Version:  cp.Version(),
			Height:   height,
			LastHash: lastHash,
		})
	}
	ob.round.VoteFailCount = 0

	if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVote)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}

	ob.messageQueue.Push(&messageItem{
		PublicHash: ob.myPublicHash,
		Message:    nm,
	})

	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendRoundVoteTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	if ob.round.RoundState == BlockVoteState {
		MyMsg, has := ob.round.RoundVoteMessageMap[ob.myPublicHash]
		if !has {
			return nil
		}
		nm := &RoundVoteMessage{
			RoundVote: &RoundVote{
				ChainID:              MyMsg.RoundVote.ChainID,
				LastHash:             MyMsg.RoundVote.LastHash,
				TargetHeight:         MyMsg.RoundVote.TargetHeight,
				TimeoutCount:         MyMsg.RoundVote.TimeoutCount,
				Formulator:           MyMsg.RoundVote.Formulator,
				FormulatorPublicHash: MyMsg.RoundVote.FormulatorPublicHash,
				Timestamp:            MyMsg.RoundVote.Timestamp,
				IsReply:              true,
			},
		}

		TargetHeight := height + 1
		if MyMsg.RoundVote.TargetHeight != TargetHeight {
			nm.RoundVote.TimeoutCount = 0
			nm.RoundVote.TargetHeight = TargetHeight
			nm.RoundVote.LastHash = lastHash
			nm.RoundVote.Timestamp = uint64(time.Now().UnixNano())
		}

		if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVote)); err != nil {
			return err
		} else {
			nm.Signature = sig
		}

		ob.ms.SendTo(TargetPubHash, nm)
	} else {
		Top, TimeoutCount, err := ob.cs.rt.TopRankInMap(ob.adjustFormulatorMap())
		if err != nil {
			return err
		}

		nm := &RoundVoteMessage{
			RoundVote: &RoundVote{
				ChainID:              cp.ChainID(),
				LastHash:             lastHash,
				TargetHeight:         height + 1,
				TimeoutCount:         uint32(TimeoutCount),
				Formulator:           Top.Address,
				FormulatorPublicHash: Top.PublicHash,
				Timestamp:            uint64(time.Now().UnixNano()),
				IsReply:              true,
			},
		}

		if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVote)); err != nil {
			return err
		} else {
			nm.Signature = sig
		}

		ob.ms.SendTo(TargetPubHash, nm)
	}
	return nil
}

func (ob *ObserverNode) sendRoundVoteAck() error {
	MinRoundVote := ob.round.RoundVoteMessageMap[ob.round.MinVotePublicHash].RoundVote
	nm := &RoundVoteAckMessage{
		RoundVoteAck: &RoundVoteAck{
			ChainID:              MinRoundVote.ChainID,
			LastHash:             MinRoundVote.LastHash,
			TargetHeight:         MinRoundVote.TargetHeight,
			TimeoutCount:         MinRoundVote.TimeoutCount,
			Formulator:           MinRoundVote.Formulator,
			FormulatorPublicHash: MinRoundVote.FormulatorPublicHash,
			PublicHash:           ob.round.MinVotePublicHash,
			Timestamp:            uint64(time.Now().UnixNano()),
			IsReply:              false,
		},
	}
	if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVoteAck)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}

	ob.messageQueue.Push(&messageItem{
		PublicHash: ob.myPublicHash,
		Message:    nm,
	})

	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendRoundVoteAckTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	if ob.round.RoundState == BlockVoteState {
		MyMsg, has := ob.round.RoundVoteAckMessageMap[ob.myPublicHash]
		if !has {
			return nil
		}
		nm := &RoundVoteAckMessage{
			RoundVoteAck: &RoundVoteAck{
				ChainID:              MyMsg.RoundVoteAck.ChainID,
				LastHash:             MyMsg.RoundVoteAck.LastHash,
				TargetHeight:         MyMsg.RoundVoteAck.TargetHeight,
				TimeoutCount:         MyMsg.RoundVoteAck.TimeoutCount,
				Formulator:           MyMsg.RoundVoteAck.Formulator,
				FormulatorPublicHash: MyMsg.RoundVoteAck.FormulatorPublicHash,
				PublicHash:           MyMsg.RoundVoteAck.PublicHash,
				Timestamp:            MyMsg.RoundVoteAck.Timestamp,
				IsReply:              true,
			},
		}
		cp := ob.cs.cn.Provider()
		height, lastHash, _ := cp.LastStatus()
		TargetHeight := height + 1
		if MyMsg.RoundVoteAck.TargetHeight != TargetHeight {
			nm.RoundVoteAck.TimeoutCount = 0
			nm.RoundVoteAck.TargetHeight = TargetHeight
			nm.RoundVoteAck.LastHash = lastHash
			nm.RoundVoteAck.Timestamp = uint64(time.Now().UnixNano())
		}

		if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVoteAck)); err != nil {
			return err
		} else {
			nm.Signature = sig
		}

		ob.ms.SendTo(TargetPubHash, nm)
	} else {
		MyMsg, has := ob.round.RoundVoteAckMessageMap[ob.myPublicHash]
		if !has {
			return nil
		}
		nm := &RoundVoteAckMessage{
			RoundVoteAck: &RoundVoteAck{
				ChainID:              MyMsg.RoundVoteAck.ChainID,
				LastHash:             MyMsg.RoundVoteAck.LastHash,
				TargetHeight:         MyMsg.RoundVoteAck.TargetHeight,
				TimeoutCount:         MyMsg.RoundVoteAck.TimeoutCount,
				Formulator:           MyMsg.RoundVoteAck.Formulator,
				FormulatorPublicHash: MyMsg.RoundVoteAck.FormulatorPublicHash,
				PublicHash:           MyMsg.RoundVoteAck.PublicHash,
				Timestamp:            MyMsg.RoundVoteAck.Timestamp,
				IsReply:              true,
			},
		}

		if sig, err := ob.key.Sign(encoding.Hash(nm.RoundVoteAck)); err != nil {
			return err
		} else {
			nm.Signature = sig
		}

		ob.ms.SendTo(TargetPubHash, nm)
	}
	return nil
}

func (ob *ObserverNode) sendRoundSetup() error {
	nm := &RoundSetupMessage{
		MinRoundVoteAck: &RoundVoteAck{
			ChainID:              ob.round.MinRoundVoteAck.ChainID,
			LastHash:             ob.round.MinRoundVoteAck.LastHash,
			TargetHeight:         ob.round.MinRoundVoteAck.TargetHeight,
			TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
			Formulator:           ob.round.MinRoundVoteAck.Formulator,
			FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
			PublicHash:           ob.round.MinRoundVoteAck.PublicHash,
			Timestamp:            uint64(time.Now().UnixNano()),
			IsReply:              false,
		},
	}
	if sig, err := ob.key.Sign(encoding.Hash(nm.MinRoundVoteAck)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}

	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendBlockVote(gen *BlockGenMessage) error {
	nm := &BlockVoteMessage{
		BlockVote: &BlockVote{
			TargetHeight:       gen.Block.Header.Height,
			Header:             &gen.Block.Header,
			GeneratorSignature: gen.GeneratorSignature,
			IsReply:            false,
		},
	}

	s := &types.BlockSign{
		HeaderHash:         encoding.Hash(gen.Block.Header),
		GeneratorSignature: gen.GeneratorSignature,
	}
	if sig, err := ob.key.Sign(encoding.Hash(s)); err != nil {
		return err
	} else {
		nm.BlockVote.ObserverSignature = sig
	}
	if sig, err := ob.key.Sign(encoding.Hash(nm.BlockVote)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}

	ob.messageQueue.Push(&messageItem{
		PublicHash: ob.myPublicHash,
		Message:    nm,
	})

	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendBlockGenTo(gen *BlockGenMessage, TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}
	ob.ms.SendTo(TargetPubHash, gen)
	return nil
}

func (ob *ObserverNode) sendBlockVoteTo(gen *BlockGenMessage, TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	nm := &BlockVoteMessage{
		BlockVote: &BlockVote{
			TargetHeight:       gen.Block.Header.Height,
			Header:             &gen.Block.Header,
			GeneratorSignature: gen.GeneratorSignature,
			IsReply:            true,
		},
	}

	s := &types.BlockSign{
		HeaderHash:         encoding.Hash(gen.Block.Header),
		GeneratorSignature: gen.GeneratorSignature,
	}
	if sig, err := ob.key.Sign(encoding.Hash(s)); err != nil {
		return err
	} else {
		nm.BlockVote.ObserverSignature = sig
	}
	if sig, err := ob.key.Sign(encoding.Hash(nm.BlockVote)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}

	ob.ms.SendTo(TargetPubHash, nm)
	return nil
}

func (ob *ObserverNode) sendBlockGenRequest(br *BlockRound) error {
	var PublicHash common.PublicHash
	has := false
	for pubhash := range br.BlockVoteMap {
		PublicHash = pubhash
		has = true
		break
	}
	if !has {
		for pubhash := range br.BlockVoteMessageWaitMap {
			PublicHash = pubhash
			has = true
			break
		}
	}
	nm := &BlockGenRequestMessage{
		BlockGenRequest: &BlockGenRequest{
			ChainID:              ob.round.MinRoundVoteAck.ChainID,
			LastHash:             ob.round.MinRoundVoteAck.LastHash,
			TargetHeight:         ob.round.MinRoundVoteAck.TargetHeight,
			TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
			Formulator:           ob.round.MinRoundVoteAck.Formulator,
			FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
			PublicHash:           ob.round.MinRoundVoteAck.PublicHash,
			Timestamp:            uint64(time.Now().UnixNano()),
		},
	}
	if sig, err := ob.key.Sign(encoding.Hash(nm.BlockGenRequest)); err != nil {
		return err
	} else {
		nm.Signature = sig
	}
	if has {
		ob.ms.SendTo(PublicHash, nm)
	} else {
		ob.ms.SendAnyone(nm)
	}
	return nil
}

func (ob *ObserverNode) sendStatusTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.SendTo(TargetPubHash, nm)
	return nil
}

func (ob *ObserverNode) broadcastStatus() error {
	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32, Count uint8) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	nm := &p2p.RequestMessage{
		Height: Height,
		Count:  Count,
	}
	ob.ms.SendTo(TargetPubHash, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		ob.requestTimer.Add(Height+i, 2*time.Second, string(TargetPubHash[:]))
	}
	return nil
}
