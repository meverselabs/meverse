package pof

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/debug"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

func (ob *ObserverNode) onObserverRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}

	if msg, is := m.(*BlockGenMessage); is {
		ob.messageQueue.Push(&messageItem{
			Message: msg,
			Packet:  bs,
		})
	} else {
		var pubhash common.PublicHash
		copy(pubhash[:], []byte(p.ID()))
		ob.messageQueue.Push(&messageItem{
			PublicHash: pubhash,
			Message:    m,
		})
	}
	return nil
}

func (ob *ObserverNode) handleObserverMessage(SenderPublicHash common.PublicHash, m interface{}, raw []byte) error {
	cp := ob.cs.cn.Provider()

	switch msg := m.(type) {
	case *RoundVoteMessage:
		if !ob.cs.observerKeyMap.Has(SenderPublicHash) {
			return ErrInvalidObserverKey
		}

		//[check round]
		if msg.RoundVote.TargetHeight != ob.round.TargetHeight {
			if msg.RoundVote.TargetHeight < ob.round.TargetHeight {
				if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
					ob.sendStatusTo(SenderPublicHash)
				}
			}
			return ErrInvalidVote
		}
		if msg.RoundVote.ChainID != cp.ChainID() {
			rlog.Println("if msg.RoundVote.ChainID != cp.ChainID() {")
			return ErrInvalidVote
		}
		if msg.RoundVote.LastHash != cp.LastHash() {
			rlog.Println("if msg.RoundVote.LastHash != cp.LastHash() {", msg.RoundVote.LastHash.String(), cp.LastHash().String())
			return ErrInvalidVote
		}
		Top, err := ob.cs.rt.TopRank(int(msg.RoundVote.TimeoutCount))
		if err != nil {
			rlog.Println("Top, err := ob.cs.rt.TopRank(int(msg.RoundVote.TimeoutCount))", msg.RoundVote.TimeoutCount)
			return err
		}
		if msg.RoundVote.Formulator != Top.Address {
			rlog.Println("if msg.RoundVote.Formulator != Top.Address {", msg.RoundVote.Formulator.String(), Top.Address.String(), msg.RoundVote.TimeoutCount)
			return ErrInvalidVote
		}
		if msg.RoundVote.FormulatorPublicHash != Top.PublicHash {
			rlog.Println("if msg.RoundVote.FormulatorPublicHash != Top.PublicHash {")
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != RoundVoteState {
			if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
				if ob.round.RoundState == BlockVoteState {
					br, has := ob.round.BlockRoundMap[msg.RoundVote.TargetHeight]
					if has {
						ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
					}
				}
				ob.sendRoundVoteTo(SenderPublicHash)
			}
			return ErrInvalidRoundState
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteMessageMap[SenderPublicHash]; has {
			if msg.RoundVote.Timestamp <= old.RoundVote.Timestamp {
				if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
					ob.sendRoundVoteTo(SenderPublicHash)
				}
				return ErrAlreadyVoted
			}
		}
		ob.round.RoundVoteMessageMap[SenderPublicHash] = msg

		if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
			ob.sendRoundVoteTo(SenderPublicHash)
		}
		if len(ob.round.RoundVoteMessageMap) >= ob.cs.observerKeyMap.Len()/2+2 {
			ob.round.RoundState = RoundVoteAckState
			if ob.roundFirstTime == 0 {
				ob.roundFirstTime = uint64(time.Now().UnixNano())
				ob.roundFirstHeight = uint32(cp.Height())
			}

			ob.sendRoundVoteAck()

			for pubhash, msg := range ob.round.RoundVoteAckMessageWaitMap {
				ob.messageQueue.Push(&messageItem{
					PublicHash: pubhash,
					Message:    msg,
				})
			}
		}
	case *RoundVoteAckMessage:
		//rlog.Println(cp.Height(), "RoundVoteAckMessage", ob.round.RoundState, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		if !ob.cs.observerKeyMap.Has(SenderPublicHash) {
			return ErrInvalidObserverKey
		}

		//[check round]
		if msg.RoundVoteAck.TargetHeight != ob.round.TargetHeight {
			if msg.RoundVoteAck.TargetHeight < ob.round.TargetHeight {
				if !msg.RoundVoteAck.IsReply && SenderPublicHash != ob.myPublicHash {
					ob.sendStatusTo(SenderPublicHash)
				}
			}
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.ChainID != cp.ChainID() {
			rlog.Println("if msg.RoundVoteAck.ChainID != cp.ChainID() {")
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.LastHash != cp.LastHash() {
			rlog.Println("if msg.RoundVoteAck.LastHash != cp.LastHash() {")
			return ErrInvalidVote
		}
		Top, err := ob.cs.rt.TopRank(int(msg.RoundVoteAck.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.RoundVoteAck.Formulator != Top.Address {
			rlog.Println("if msg.RoundVoteAck.Formulator != Top.Address {")
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.FormulatorPublicHash != Top.PublicHash {
			rlog.Println("if msg.RoundVoteAck.FormulatorPublicHash != Top.PublicHash {")
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != RoundVoteAckState {
			if ob.round.RoundState < RoundVoteAckState {
				ob.round.RoundVoteAckMessageWaitMap[SenderPublicHash] = msg
			} else {
				if !msg.RoundVoteAck.IsReply && SenderPublicHash != ob.myPublicHash {
					if ob.round.RoundState == BlockVoteState {
						br, has := ob.round.BlockRoundMap[msg.RoundVoteAck.TargetHeight]
						if has {
							ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
						}
					}
					ob.sendRoundVoteAckTo(SenderPublicHash)
				}
			}
			return ErrInvalidRoundState
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteAckMessageMap[SenderPublicHash]; has {
			if msg.RoundVoteAck.Timestamp <= old.RoundVoteAck.Timestamp {
				if !msg.RoundVoteAck.IsReply {
					ob.sendRoundVoteAckTo(SenderPublicHash)
				}
				return ErrAlreadyVoted
			}
		}
		ob.round.RoundVoteAckMessageMap[SenderPublicHash] = msg

		rlog.Println(ob.myPublicHash.String(), cp.Height(), "RoundVoteAckMessage", ob.round.RoundState, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))

		if !msg.RoundVoteAck.IsReply && SenderPublicHash != ob.myPublicHash {
			ob.sendRoundVoteAckTo(SenderPublicHash)
		}

		if len(ob.round.RoundVoteAckMessageMap) >= ob.cs.observerKeyMap.Len()/2+1 {
			var MinRoundVoteAck *RoundVoteAck
			PublicHashCountMap := map[common.PublicHash]int{}
			TimeoutCountMap := map[uint32]int{}
			for _, msg := range ob.round.RoundVoteAckMessageMap {
				vt := msg.RoundVoteAck
				TimeoutCount := TimeoutCountMap[vt.TimeoutCount]
				TimeoutCount++
				TimeoutCountMap[vt.TimeoutCount] = TimeoutCount
				PublicHashCount := PublicHashCountMap[vt.PublicHash]
				PublicHashCount++
				PublicHashCountMap[vt.PublicHash] = PublicHashCount
				if TimeoutCount >= ob.cs.observerKeyMap.Len()/2+1 && PublicHashCount >= ob.cs.observerKeyMap.Len()/2+1 {
					MinRoundVoteAck = vt
					break
				}
			}

			if MinRoundVoteAck != nil {
				ob.round.RoundState = BlockWaitState
				ob.round.MinRoundVoteAck = MinRoundVoteAck
				ob.round.VoteFailCount = 0
				RemainBlocks := ob.cs.maxBlocksPerFormulator
				if MinRoundVoteAck.TimeoutCount == 0 {
					RemainBlocks = ob.cs.maxBlocksPerFormulator - ob.cs.blocksBySameFormulator
				}
				for TargetHeight, br := range ob.round.BlockRoundMap {
					if TargetHeight >= ob.round.TargetHeight+RemainBlocks {
						delete(ob.round.BlockRoundMap, TargetHeight)
					} else if br.BlockGenMessageWait != nil {
						if br.BlockGenMessageWait.Block.Header.Generator != ob.round.MinRoundVoteAck.Formulator {
							br.BlockGenMessageWait = nil
						}
					}
				}

				if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
					if debug.DEBUG {
						rlog.Println(ob.myPublicHash.String(), "Observer", "BlockReqMessage", ob.round.MinRoundVoteAck.Formulator.String(), ob.round.MinRoundVoteAck.TimeoutCount, cp.Height())
					}
					nm := &BlockReqMessage{
						PrevHash:             ob.round.MinRoundVoteAck.LastHash,
						TargetHeight:         ob.round.MinRoundVoteAck.TargetHeight,
						TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
						Formulator:           ob.round.MinRoundVoteAck.Formulator,
						FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
					}
					ob.sendMessage(0, ob.round.MinRoundVoteAck.Formulator, nm)
					//ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, p2p.MessageToPacket(nm))
				}

				br := ob.round.BlockRoundMap[ob.round.TargetHeight]
				if br != nil {
					if br.BlockGenMessageWait != nil && br.BlockGenMessage == nil {
						if br.BlockGenMessageWait.Block.Header.Generator == ob.round.MinRoundVoteAck.Formulator {
							ob.messageQueue.Push(&messageItem{
								Message: br.BlockGenMessageWait,
							})
						}
						br.BlockGenMessageWait = nil
					}
				}
			}
		}
	case *BlockGenMessage:
		rlog.Println(ob.myPublicHash.String(), cp.Height(), "BlockGenMessage", ob.round.RoundState, msg.Block.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.Block.Header.Height]
		if !has {
			rlog.Println(msg.Block.Header.Generator.String(), "br, has := ob.round.BlockRoundMap[msg.Block.Header.Height]", msg.Block.Header.Height, ob.round.TargetHeight)
			return ErrInvalidVote
		}
		if br.BlockGenMessage != nil {
			rlog.Println(msg.Block.Header.Generator.String(), "if br.BlockGenMessage != nil {", msg.Block.Header.Height, ob.round.TargetHeight)
			return ErrInvalidVote
		}

		if ob.round.MinRoundVoteAck != nil {
			if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
				if len(raw) > 0 {
					ob.ms.BroadcastPacket(raw)
					if debug.DEBUG {
						rlog.Println(ob.myPublicHash.String(), cp.Height(), "BlockGenBroadcast", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				}
			} else {
				if len(raw) > 0 {
					var MaxValue uint64
					var MaxID string
					base := binutil.LittleEndian.Uint64(ob.round.MinRoundVoteAck.PublicHash[:])
					IDs := []string{}
					for _, p := range ob.ms.Peers() {
						IDs = append(IDs, p.ID())
					}
					IDs = append(IDs, string(ob.myPublicHash[:]))
					for _, ID := range IDs {
						if ID != string(ob.round.MinRoundVoteAck.PublicHash[:]) {
							value := binutil.LittleEndian.Uint64([]byte(ID))
							var diff uint64
							if base > value {
								diff = base - value
							} else {
								diff = value - base
							}
							if MaxValue < diff {
								MaxValue = diff
								MaxID = ID
							}
						}
					}

					var MaxHash common.PublicHash
					copy(MaxHash[:], []byte(MaxID))
					if ob.myPublicHash == MaxHash {
						adjustMap := ob.adjustFormulatorMap()
						delete(adjustMap, ob.round.MinRoundVoteAck.Formulator)

						var NextTop *Rank
						if len(adjustMap) > 0 {
							r, _, err := ob.cs.rt.TopRankInMap(adjustMap)
							if err != nil {
								return err
							}
							NextTop = r
						}
						if NextTop != nil {
							ob.sendMessagePacket(1, NextTop.Address, raw)
							if debug.DEBUG {
								rlog.Println(ob.myPublicHash.String(), cp.Height(), "BlockGenToNextTop", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
							}
						}
					}
				}
			}
		}

		if msg.Block.Header.Height != ob.round.TargetHeight {
			if msg.Block.Header.Height > ob.round.TargetHeight {
				br.BlockGenMessageWait = msg
			}
			rlog.Println(msg.Block.Header.Generator.String(), "if msg.Block.Header.Height != ob.round.TargetHeight {", msg.Block.Header.Height, ob.round.TargetHeight)
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != BlockWaitState {
			if ob.round.RoundState < BlockWaitState {
				br.BlockGenMessageWait = msg
			}
			rlog.Println(msg.Block.Header.Generator.String(), "if ob.round.RoundState != BlockWaitState {", ob.round.RoundState, BlockWaitState)
			return ErrInvalidVote
		}
		TimeoutCount, err := ob.cs.DecodeConsensusData(msg.Block.Header.ConsensusData)
		if err != nil {
			return err
		}
		Top, err := ob.cs.rt.TopRank(int(TimeoutCount))
		if err != nil {
			return err
		}
		if msg.Block.Header.Generator != Top.Address {
			rlog.Println(msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != Top.Address {", msg.Block.Header.Generator.String(), Top.Address.String(), TimeoutCount)
			return ErrInvalidVote
		}
		if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Formulator {
			rlog.Println(msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Formulator {")
			return ErrInvalidVote
		}
		bh := encoding.Hash(msg.Block.Header)
		if pubkey, err := common.RecoverPubkey(bh, msg.GeneratorSignature); err != nil {
			return err
		} else if Signer := common.NewPublicHash(pubkey); Signer != Top.PublicHash {
			return ErrInvalidTopSignature
		} else if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {
			rlog.Println(msg.Block.Header.Generator.String(), "if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {")
			return ErrInvalidVote
		}
		if err := ob.cs.ct.ValidateHeader(&msg.Block.Header); err != nil {
			rlog.Println(msg.Block.Header.Generator.String(), "if err := ob.cs.ct.ValidateHeader(&msg.Block.Header); err != nil {", err)
			return err
		}

		//[if valid block]
		Now := uint64(time.Now().UnixNano())
		if msg.Block.Header.Timestamp > Now+uint64(10*time.Second) {
			rlog.Println(msg.Block.Header.Generator.String(), "if msg.Block.Header.Timestamp > Now+uint64(10*time.Second) {")
			return ErrInvalidVote
		}

		ctx := ob.cs.ct.NewContext()
		ChainID := ob.cs.cn.Provider().ChainID()
		sm := map[hash.Hash256][]common.PublicHash{}
		for i, tx := range msg.Block.Transactions {
			t := msg.Block.TransactionTypes[i]
			TxHash := chain.HashTransactionByType(ChainID, t, tx)
			if v, err := ob.sigCache.Get(TxHash); err != nil {
			} else if v != nil {
				sm[TxHash] = []common.PublicHash{v.(common.PublicHash)} //TEMP
			}
		}
		if err := ob.cs.ct.ExecuteBlockOnContext(msg.Block, ctx, sm); err != nil {
			rlog.Println(msg.Block.Header.Generator.String(), "if err := ob.cs.ct.ExecuteBlockOnContext(msg.Block, ctx); err != nil {", err)
			return err
		}
		if msg.Block.Header.ContextHash != ctx.Hash() {
			rlog.Println(msg.Block.Header.Generator.String(), "if msg.Block.Header.ContextHash != ctx.Hash() {")
			return chain.ErrInvalidContextHash
		}

		ob.round.RoundState = BlockVoteState
		br.BlockGenMessage = msg
		br.Context = ctx

		ob.sendBlockVote(br.BlockGenMessage)

		for pubhash, msg := range br.BlockVoteMessageWaitMap {
			ob.messageQueue.Push(&messageItem{
				PublicHash: pubhash,
				Message:    msg,
			})
		}
	case BlockGenRequestMessage:
		if !ob.cs.observerKeyMap.Has(SenderPublicHash) {
			return ErrInvalidObserverKey
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.BlockGenRequest.TargetHeight]
		if !has {
			rlog.Println("br, has := ob.round.BlockRoundMap[msg.BlockGenRequest.TargetHeight]")
			return ErrInvalidVote
		}

		if msg.BlockGenRequest.TargetHeight != ob.round.TargetHeight {
			if msg.BlockGenRequest.TargetHeight < ob.round.TargetHeight {
				if SenderPublicHash != ob.myPublicHash {
					ob.sendStatusTo(SenderPublicHash)
				}
			}
			return ErrInvalidVote
		}
		if msg.BlockGenRequest.ChainID != cp.ChainID() {
			rlog.Println("if msg.BlockGenRequest.ChainID != cp.ChainID() {")
			return ErrInvalidVote
		}
		if msg.BlockGenRequest.LastHash != cp.LastHash() {
			rlog.Println("if msg.BlockGenRequest.LastHash != cp.LastHash() {")
			return ErrInvalidVote
		}
		Top, err := ob.cs.rt.TopRank(int(msg.BlockGenRequest.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.BlockGenRequest.Formulator != Top.Address {
			rlog.Println("if msg.BlockGenRequest.Formulator != Top.Address {")
			return ErrInvalidVote
		}
		if msg.BlockGenRequest.FormulatorPublicHash != Top.PublicHash {
			rlog.Println("if msg.BlockGenRequest.FormulatorPublicHash != Top.PublicHash {")
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState == BlockVoteState {
			if _, has := br.BlockVoteMap[SenderPublicHash]; !has {
				ob.sendBlockGenTo(br.BlockGenMessage, SenderPublicHash)
			}
			ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
		}
	case *BlockVoteMessage:
		//rlog.Println(cp.Height(), encoding.Hash(msg.BlockVote.Header), "BlockVoteMessage", ob.round.RoundState, msg.BlockVote.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		if !ob.cs.observerKeyMap.Has(SenderPublicHash) {
			return ErrInvalidObserverKey
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.BlockVote.Header.Height]
		if !has {
			return ErrInvalidVote
		}
		if msg.BlockVote.Header.Height != ob.round.TargetHeight {
			if msg.BlockVote.Header.Height < ob.round.TargetHeight {
				ob.sendStatusTo(SenderPublicHash)
			} else {
				br.BlockVoteMessageWaitMap[SenderPublicHash] = msg
			}
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != BlockVoteState {
			if ob.round.RoundState < BlockVoteState {
				if _, has := ob.ignoreMap[msg.BlockVote.Header.Generator]; has {
					delete(ob.ignoreMap, msg.BlockVote.Header.Generator)
					ob.round.VoteFailCount = 0
				}
				br.BlockVoteMessageWaitMap[SenderPublicHash] = msg

				if ob.round.RoundState == BlockWaitState && br.BlockGenMessageWait == nil && br.BlockGenMessage == nil {
					ob.sendBlockGenRequest(br)
				}
			}
			return ErrInvalidVote
		}
		TimeoutCount, err := ob.cs.DecodeConsensusData(msg.BlockVote.Header.ConsensusData)
		if err != nil {
			return err
		}
		Top, err := ob.cs.rt.TopRank(int(TimeoutCount))
		if err != nil {
			return err
		}
		if msg.BlockVote.Header.Generator != Top.Address {
			rlog.Println("Observer", cp.Height(), ob.myPublicHash.String(), Top.Address.String(), encoding.Hash(msg.BlockVote.Header), "if msg.BlockVote.Header.Generator != Top.Address {")
			return ErrInvalidVote
		}
		bh := encoding.Hash(msg.BlockVote.Header)
		pubkey, err := common.RecoverPubkey(bh, msg.BlockVote.GeneratorSignature)
		if err != nil {
			return err
		}
		Signer := common.NewPublicHash(pubkey)
		if Signer != Top.PublicHash {
			rlog.Println(encoding.Hash(msg.BlockVote.Header), "if Signer != Top.PublicHash {")
			return ErrInvalidTopSignature
		}
		if msg.BlockVote.Header.Generator != ob.round.MinRoundVoteAck.Formulator {
			rlog.Println(encoding.Hash(msg.BlockVote.Header), "if msg.BlockVote.Header.Generator != ob.round.MinRoundVoteAck.Formulator {")
			return ErrInvalidVote
		}
		if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {
			rlog.Println(encoding.Hash(msg.BlockVote.Header), "if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {")
			return ErrInvalidVote
		}
		if msg.BlockVote.Header.PrevHash != cp.LastHash() {
			rlog.Println(encoding.Hash(msg.BlockVote.Header), "if msg.BlockVote.Header.PrevHash != cp.LastHash() {")
			return ErrInvalidVote
		}
		if bh != encoding.Hash(br.BlockGenMessage.Block.Header) {
			rlog.Println(encoding.Hash(msg.BlockVote.Header), "if bh != encoding.Hash(br.BlockGenMessage.Block.Header) {")
			return ErrInvalidVote
		}
		if err := ob.cs.ct.ValidateHeader(msg.BlockVote.Header); err != nil {
			rlog.Println(msg.BlockVote.Header.Generator.String(), "if err := ob.cs.ct.ValidateHeader(msg.BlockVote.Header); err != nil {")
			return err
		}

		s := &types.BlockSign{
			HeaderHash:         bh,
			GeneratorSignature: msg.BlockVote.GeneratorSignature,
		}
		if pubkey, err := common.RecoverPubkey(encoding.Hash(s), msg.BlockVote.ObserverSignature); err != nil {
			return err
		} else if SenderPublicHash != common.NewPublicHash(pubkey) {
			return ErrInvalidVote
		}

		if _, has := br.BlockVoteMap[SenderPublicHash]; has {
			if !msg.BlockVote.IsReply {
				ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
			}
			return ErrAlreadyVoted
		}
		br.BlockVoteMap[SenderPublicHash] = msg.BlockVote

		rlog.Println("Observer", cp.Height(), encoding.Hash(msg.BlockVote.Header), "BlockVoteMessage", ob.round.RoundState, msg.BlockVote.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))

		//[check state]
		if !msg.BlockVote.IsReply && SenderPublicHash != ob.myPublicHash {
			ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
		}

		//[apply vote]
		if len(br.BlockVoteMap) >= ob.cs.observerKeyMap.Len()/2+1 {
			sigs := []common.Signature{}
			for _, vt := range br.BlockVoteMap {
				sigs = append(sigs, vt.ObserverSignature)
			}

			PastTime := uint64(time.Now().UnixNano()) - ob.roundFirstTime
			ExpectedTime := uint64(msg.BlockVote.Header.Height-ob.roundFirstHeight) * uint64(BlockTime)
			if PastTime < ExpectedTime {
				diff := time.Duration(ExpectedTime - PastTime)
				if diff > BlockTime {
					diff = BlockTime
				}
				time.Sleep(diff)
			}

			b := &types.Block{
				Header:                br.BlockGenMessage.Block.Header,
				TransactionTypes:      br.BlockGenMessage.Block.TransactionTypes,
				Transactions:          br.BlockGenMessage.Block.Transactions,
				TransactionSignatures: br.BlockGenMessage.Block.TransactionSignatures,
				Signatures:            append([]common.Signature{br.BlockGenMessage.GeneratorSignature}, sigs...),
			}
			if err := ob.cs.ct.ConnectBlockWithContext(b, br.Context); err != nil {
				return err
			} else {
				ob.broadcastStatus()
			}
			delete(ob.ignoreMap, ob.round.MinRoundVoteAck.Formulator)

			adjustMap := ob.adjustFormulatorMap()
			delete(adjustMap, ob.round.MinRoundVoteAck.Formulator)
			var NextTop *Rank
			if len(adjustMap) > 0 {
				r, _, err := ob.cs.rt.TopRankInMap(adjustMap)
				if err != nil {
					return err
				}
				NextTop = r
			}

			if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
				nm := &BlockObSignMessage{
					TargetHeight: msg.BlockVote.Header.Height,
					BlockSign: &types.BlockSign{
						HeaderHash:         bh,
						GeneratorSignature: msg.BlockVote.GeneratorSignature,
					},
					ObserverSignatures: sigs,
				}
				bs := p2p.MessageToPacket(nm)
				ob.sendMessagePacket(0, ob.round.MinRoundVoteAck.Formulator, bs)
				//ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, bs)
				if NextTop != nil {
					ob.sendMessagePacket(0, NextTop.Address, bs)
					//ob.fs.SendTo(NextTop.Address, bs)
				}
			} else {
				if NextTop != nil {
					delete(adjustMap, NextTop.Address)
				}
				if len(adjustMap) > 0 {
					ranks, err := ob.cs.rt.RanksInMap(adjustMap, 3)
					if err == nil {
						for _, v := range ranks {
							ob.sendMessage(1, v.Address, &p2p.StatusMessage{
								Version:  b.Header.Version,
								Height:   b.Header.Height,
								LastHash: bh,
							})
						}
					}
				}
			}
			if debug.DEBUG {
				rlog.Println("Observer", cp.Height(), "BlockConnected", b.Header.Generator.String(), ob.round.RoundState, msg.BlockVote.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}

			NextHeight := ob.round.TargetHeight + 1
			Top, err := ob.cs.rt.TopRank(0)
			if err != nil {
				return err
			}
			brNext, has := ob.round.BlockRoundMap[NextHeight]
			if has && Top.Address == ob.round.MinRoundVoteAck.Formulator {
				ob.round.RoundState = BlockWaitState
				ob.round.VoteFailCount = 0
				ob.round.TargetHeight++
				if brNext.BlockGenMessageWait != nil && brNext.BlockGenMessage == nil {
					ob.messageQueue.Push(&messageItem{
						Message: brNext.BlockGenMessageWait,
					})
				} else if brNext.BlockGenMessage != nil {
					ob.sendBlockGenRequest(brNext)
				}
			} else {
				ob.resetVoteRound(false)
			}
		}
	case *p2p.RequestMessage:
		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
		Height := ob.cs.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		bs, err := p2p.BlockPacketWithCache(msg, ob.cs.cn.Provider(), ob.batchCache, ob.singleCache)
		if err != nil {
			return err
		}
		if err := ob.ms.SendTo(SenderPublicHash, bs); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		Height := cp.Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				if BaseHeight > msg.Height {
					break
				}
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
					if !ob.requestTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					ob.sendRequestBlockTo(SenderPublicHash, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
						if !ob.requestTimer.Exist(i) {
							ob.sendRequestBlockTo(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(SenderPublicHash.String(), h.String(), msg.LastHash.String(), msg.Height)
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := ob.addBlock(b); err != nil {
				if err != nil {
					panic(chain.ErrFoundForkedBlock)
				}
				return err
			}
		}
	default:
		return p2p.ErrUnknownMessage
	}
	return nil
}
