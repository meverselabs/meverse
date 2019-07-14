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
	nm := &RoundVoteMessage{
		RoundVote: &RoundVote{
			LastHash:             cp.LastHash(),
			TargetHeight:         cp.Height() + 1,
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
			Height:   cp.Height(),
			LastHash: cp.LastHash(),
		})
	} else if gh < cp.Height() {
		ob.fs.SendTo(Top.Address, &p2p.StatusMessage{
			Version:  cp.Version(),
			Height:   cp.Height(),
			LastHash: cp.LastHash(),
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
	if ob.round.RoundState == BlockVoteState {
		MyMsg, has := ob.round.RoundVoteMessageMap[ob.myPublicHash]
		if !has {
			return nil
		}
		nm := &RoundVoteMessage{
			RoundVote: &RoundVote{
				LastHash:             MyMsg.RoundVote.LastHash,
				TargetHeight:         MyMsg.RoundVote.TargetHeight,
				TimeoutCount:         MyMsg.RoundVote.TimeoutCount,
				Formulator:           MyMsg.RoundVote.Formulator,
				FormulatorPublicHash: MyMsg.RoundVote.FormulatorPublicHash,
				Timestamp:            MyMsg.RoundVote.Timestamp,
				IsReply:              true,
			},
		}
		TargetHeight := cp.Height() + 1
		if MyMsg.RoundVote.TargetHeight != TargetHeight {
			nm.RoundVote.TimeoutCount = 0
			nm.RoundVote.TargetHeight = TargetHeight
			nm.RoundVote.LastHash = cp.LastHash()
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
				LastHash:             cp.LastHash(),
				TargetHeight:         cp.Height() + 1,
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
		TargetHeight := cp.Height() + 1
		if MyMsg.RoundVoteAck.TargetHeight != TargetHeight {
			nm.RoundVoteAck.TimeoutCount = 0
			nm.RoundVoteAck.TargetHeight = TargetHeight
			nm.RoundVoteAck.LastHash = cp.LastHash()
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

func (ob *ObserverNode) sendStatusTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	cp := ob.cs.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	ob.ms.SendTo(TargetPubHash, nm)
	return nil
}

func (ob *ObserverNode) broadcastStatus() error {
	cp := ob.cs.cn.Provider()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	}
	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	nm := &p2p.RequestMessage{
		Height: Height,
	}
	ob.ms.SendTo(TargetPubHash, nm)
	ob.requestTimer.Add(Height, 2*time.Second, string(TargetPubHash[:]))
	return nil
}
