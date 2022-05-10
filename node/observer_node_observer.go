package node

import (
	"log"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
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
		var PubKey common.PublicKey
		copy(PubKey[:], []byte(p.ID()))
		ob.messageQueue.Push(&messageItem{
			PublicKey: PubKey,
			Message:   m,
		})
	}
	return nil
}

func (ob *ObserverNode) handleObserverMessage(SenderPublicKey common.PublicKey, m interface{}, raw []byte) error {
	cp := ob.cn.Provider()

	switch msg := m.(type) {
	case *RoundVoteMessage:
		if !ob.observerKeyMap[SenderPublicKey] {
			return errors.WithStack(ErrInvalidObserverKey)
		}

		//[check round]
		if msg.TargetHeight != ob.round.TargetHeight {
			if msg.TargetHeight < ob.round.TargetHeight {
				if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
					ob.sendStatusTo(SenderPublicKey)
				}
			}
			return errors.Wrap(ErrInvalidVote, "msg.TargetHeight != ob.round.TargetHeight")
		}
		if msg.ChainID.Cmp(cp.ChainID()) != 0 {
			return errors.Wrap(ErrInvalidVote, "msg.ChainID != cp.ChainID()")
		}
		if msg.LastHash != cp.LastHash() {
			return errors.WithMessagef(ErrInvalidVote, "msg.LastHash != cp.LastHash(), %v, %v", msg.LastHash.String(), cp.LastHash().String())
		}
		Top, err := ob.cn.TopGenerator(msg.TimeoutCount)
		if err != nil {
			return errors.WithMessagef(err, "ob.cn.TopGenerator(msg.TimeoutCount) %v", msg.TimeoutCount)
		}
		if msg.Generator != Top {
			return errors.WithMessagef(ErrInvalidVote, "if msg.Generator != Top.Address {", msg.Generator.String(), Top.String(), msg.TimeoutCount)
		}

		//[check state]
		if ob.round.RoundState != RoundVoteState {
			if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
				if ob.round.RoundState == BlockVoteState {
					br, has := ob.round.BlockRoundMap[msg.TargetHeight]
					if has {
						ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicKey)
					}
				}
				ob.sendRoundVoteTo(SenderPublicKey)
			}
			return errors.WithStack(ErrInvalidRoundState)
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteMessageMap[SenderPublicKey]; has {
			if msg.Timestamp <= old.Timestamp {
				if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
					ob.sendRoundVoteTo(SenderPublicKey)
				}
				return errors.WithStack(ErrAlreadyVoted)
			}
		}
		ob.round.RoundVoteMessageMap[SenderPublicKey] = msg

		if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
			ob.sendRoundVoteTo(SenderPublicKey)
		}
		if len(ob.round.RoundVoteMessageMap) >= len(ob.observerKeyMap)/2+2 {
			ob.round.RoundState = RoundVoteAckState
			if ob.roundFirstTime == 0 {
				ob.roundFirstTime = uint64(time.Now().UnixNano())
				ob.roundFirstHeight = uint32(cp.Height())
			}

			ob.sendRoundVoteAck()

			for PubKey, msg := range ob.round.RoundVoteAckMessageWaitMap {
				ob.messageQueue.Push(&messageItem{
					PublicKey: PubKey,
					Message:   msg,
				})
			}
		}
	case *RoundVoteAckMessage:
		//log.Println(ob.obID, cp.Height(), "RoundVoteAckMessage", ob.round.RoundState, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		if !ob.observerKeyMap[SenderPublicKey] {
			return errors.WithStack(ErrInvalidObserverKey)
		}

		//[check round]
		if msg.TargetHeight != ob.round.TargetHeight {
			if msg.TargetHeight < ob.round.TargetHeight {
				if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
					ob.sendStatusTo(SenderPublicKey)
				}
			}
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.ChainID.Cmp(cp.ChainID()) != 0 {
			log.Println(ob.obID, "if msg.ChainID != cp.ChainID() {")
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.LastHash != cp.LastHash() {
			log.Println(ob.obID, "if msg.LastHash != cp.LastHash() {")
			return errors.WithStack(ErrInvalidVote)
		}
		Top, err := ob.cn.TopGenerator(msg.TimeoutCount)
		if err != nil {
			return err
		}
		if msg.Generator != Top {
			log.Println(ob.obID, "if msg.Generator != Top.Address {")
			return errors.WithStack(ErrInvalidVote)
		}

		//[check state]
		if ob.round.RoundState != RoundVoteAckState {
			if ob.round.RoundState < RoundVoteAckState {
				ob.round.RoundVoteAckMessageWaitMap[SenderPublicKey] = msg
			} else {
				if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
					if ob.round.RoundState == BlockVoteState {
						br, has := ob.round.BlockRoundMap[msg.TargetHeight]
						if has {
							ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicKey)
						}
					}
					ob.sendRoundVoteAckTo(SenderPublicKey)
				}
			}
			return errors.WithStack(ErrInvalidRoundState)
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteAckMessageMap[SenderPublicKey]; has {
			if msg.Timestamp <= old.Timestamp {
				if !msg.IsReply {
					ob.sendRoundVoteAckTo(SenderPublicKey)
				}
				return errors.WithStack(ErrAlreadyVoted)
			}
		}
		ob.round.RoundVoteAckMessageMap[SenderPublicKey] = msg

		if DEBUG {
			log.Println(ob.obID, "ob", ob.myPublicKey.Address().String(), cp.Height(), "RoundVoteAckMessage", ob.round.RoundState, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		}

		if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
			ob.sendRoundVoteAckTo(SenderPublicKey)
		}

		if len(ob.round.RoundVoteAckMessageMap) >= len(ob.observerKeyMap)/2+1 {
			var MinRoundVoteAck *RoundVoteAckMessage
			PublicKeyCountMap := map[common.PublicKey]int{}
			TimeoutCountMap := map[uint32]int{}
			for _, msg := range ob.round.RoundVoteAckMessageMap {
				vt := msg
				TimeoutCount := TimeoutCountMap[vt.TimeoutCount]
				TimeoutCount++
				TimeoutCountMap[vt.TimeoutCount] = TimeoutCount
				PublicKeyCount := PublicKeyCountMap[vt.PublicKey]
				PublicKeyCount++
				PublicKeyCountMap[vt.PublicKey] = PublicKeyCount
				if TimeoutCount >= len(ob.observerKeyMap)/2+1 && PublicKeyCount >= len(ob.observerKeyMap)/2+1 {
					MinRoundVoteAck = vt
					break
				}
			}

			if MinRoundVoteAck != nil {
				ob.round.RoundState = BlockWaitState
				ob.round.MinRoundVoteAck = MinRoundVoteAck
				ob.round.VoteFailCount = 0
				RemainBlocks := prefix.MaxBlocksPerGenerator
				for TargetHeight, br := range ob.round.BlockRoundMap {
					if TargetHeight >= ob.round.TargetHeight+RemainBlocks {
						delete(ob.round.BlockRoundMap, TargetHeight)
					} else if br.BlockGenMessageWait != nil {
						if br.BlockGenMessageWait.Block.Header.Generator != ob.round.MinRoundVoteAck.Generator {
							br.BlockGenMessageWait = nil
						}
					}
				}

				if ob.round.MinRoundVoteAck.PublicKey == ob.myPublicKey {
					if DEBUG {
						log.Println(ob.obID, "Observer", cp.Height(), "BlockReqMessage", ob.round.MinRoundVoteAck.Generator.String(), ob.round.MinRoundVoteAck.TimeoutCount, cp.Height())
					}
					nm := &BlockReqMessage{
						PrevHash:     ob.round.MinRoundVoteAck.LastHash,
						TargetHeight: ob.round.MinRoundVoteAck.TargetHeight,
						TimeoutCount: ob.round.MinRoundVoteAck.TimeoutCount,
						Generator:    ob.round.MinRoundVoteAck.Generator,
					}
					ob.sendMessage(0, ob.round.MinRoundVoteAck.Generator, nm)
					//ob.fs.SendTo(ob.round.MinRoundVoteAck.Generator, p2p.MessageToPacket(nm))
				}

				br := ob.round.BlockRoundMap[ob.round.TargetHeight]
				if br != nil {
					if ob.round.MinRoundVoteAck.PublicKey != ob.myPublicKey {
						if has, _ := ob.hasBlockVoteHalf(ob.round.TargetHeight); has {
							ob.sendBlockGenRequest(br)
						}
					}
					if br.BlockGenMessageWait != nil && br.BlockGenMessage == nil {
						if br.BlockGenMessageWait.Block.Header.Generator == ob.round.MinRoundVoteAck.Generator {
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
		if DEBUG {
			log.Println(ob.obID, "ob BlockGenMessage", ob.myPublicKey.Address().String(), cp.Height(), "BlockGenMessage", ob.round.RoundState, msg.Block.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.Block.Header.Height]
		if !has {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "br, has := ob.round.BlockRoundMap[msg.Block.Header.Height]", msg.Block.Header.Height, ob.round.TargetHeight)
			return errors.WithStack(ErrInvalidVote)
		}
		if br.BlockGenMessage != nil {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if br.BlockGenMessage != nil {", msg.Block.Header.Height, ob.round.TargetHeight)
			return errors.WithStack(ErrInvalidVote)
		}

		if ob.round.MinRoundVoteAck != nil {
			if ob.round.MinRoundVoteAck.PublicKey == ob.myPublicKey {
				if len(raw) > 0 {
					ob.ms.BroadcastPacket(raw)
					if DEBUG {
						log.Println(ob.obID, "ob BlockGenMessage", ob.myPublicKey.Address().String(), cp.Height(), "BlockGenBroadcast", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				}
			} else {
				if len(raw) > 0 {
					var MaxValue uint64
					var MaxID string
					base := bin.Uint64(ob.round.MinRoundVoteAck.PublicKey[:])
					IDs := []string{}
					for _, p := range ob.ms.Peers() {
						IDs = append(IDs, p.ID())
					}
					IDs = append(IDs, string(ob.myPublicKey[:]))
					for _, ID := range IDs {
						if ID != string(ob.round.MinRoundVoteAck.PublicKey[:]) {
							value := bin.Uint64([]byte(ID))
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

					var MaxHash common.PublicKey
					copy(MaxHash[:], []byte(MaxID))
					if ob.myPublicKey == MaxHash {
						adjustMap := ob.adjustGeneratorMap()
						delete(adjustMap, ob.round.MinRoundVoteAck.Generator)

						var NextTop common.Address
						if len(adjustMap) > 0 {
							r, _, err := ob.cn.TopGeneratorInMap(adjustMap)
							if err != nil {
								log.Println(ob.obID, "ob BlockGenMessage", ob.myPublicKey.Address().String(), cp.Height(), "BlockGenToNextTop", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
								return err
							}
							NextTop = r
						}
						var zerAddr common.Address
						if NextTop != zerAddr {
							ob.sendMessagePacket(1, NextTop, raw)
							if DEBUG {
								log.Println(ob.obID, "ob BlockGenMessage", ob.myPublicKey.Address().String(), cp.Height(), "BlockGenToNextTop", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
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
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Height != ob.round.TargetHeight {", msg.Block.Header.Height, ob.round.TargetHeight)
			return errors.WithStack(ErrInvalidVote)
		}

		//[check state]
		if ob.round.RoundState != BlockWaitState {
			if ob.round.RoundState < BlockWaitState {
				br.BlockGenMessageWait = msg
			}
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if ob.round.RoundState != BlockWaitState {", ob.round.RoundState, BlockWaitState)
			return errors.WithStack(ErrInvalidVote)
		}
		Top, err := ob.cn.TopGenerator(msg.Block.Header.TimeoutCount)
		if err != nil {
			return err
		}
		if msg.Block.Header.Generator != Top {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != Top.Address {", msg.Block.Header.Generator.String(), Top.String(), msg.Block.Header.TimeoutCount)
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Generator {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Generator {")
			return errors.WithStack(ErrInvalidVote)
		}
		bh := bin.MustWriterToHash(&msg.Block.Header)
		if pubkey, err := common.RecoverPubkey(ob.ChainID, bh, msg.GeneratorSignature); err != nil {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Generator {")
			return err
		} else if Signer := pubkey.Address(); Signer != Top {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Generator {")
			return errors.WithStack(ErrInvalidTopSignature)
		} else if Signer != ob.round.MinRoundVoteAck.Generator {
			log.Printf("ob BlockGenMessage %v %+v", msg.Block.Header.Generator.String(), err)
			return errors.WithStack(ErrInvalidVote)
		}
		if err := ob.ct.ValidateHeader(&msg.Block.Header); err != nil {
			log.Printf("ob BlockGenMessage %v %+v", msg.Block.Header.Generator.String(), err)
			return err
		}

		//[if valid block]
		Now := uint64(time.Now().UnixNano())
		if msg.Block.Header.Timestamp > Now+uint64(10*time.Second) {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if msg.Block.Header.Timestamp > Now+uint64(10*time.Second) {")
			return errors.WithStack(ErrInvalidVote)
		}

		ctx := ob.ct.NewContext()
		if err := ob.ct.ExecuteBlockOnContext(msg.Block, ctx, nil); err != nil {
			log.Println(ob.obID, "ob BlockGenMessage", msg.Block.Header.Generator.String(), "if err := ob.ct.ExecuteBlockOnContext(msg.Block, ctx); err != nil {", err)
			return err
		}

		if msg.Block.Header.ContextHash != ctx.Hash() {
			log.Println(ob.obID, msg.Block.Header.Generator.String(), "if msg.Block.Header.ContextHash != ctx.Hash() {", msg.Block.Header.ContextHash.String(), ctx.Hash().String())
			log.Println(ob.obID, ctx.Dump())
			return errors.WithStack(chain.ErrInvalidContextHash)
		}

		ob.round.RoundState = BlockVoteState
		br.BlockGenMessage = msg
		br.Context = ctx

		ob.sendBlockVote(br.BlockGenMessage)

		for PubKey, msg := range br.BlockVoteMessageWaitMap {
			ob.messageQueue.Push(&messageItem{
				PublicKey: PubKey,
				Message:   msg,
			})
		}
	case *BlockGenRequestMessage:
		if !ob.observerKeyMap[SenderPublicKey] {
			return errors.WithStack(ErrInvalidObserverKey)
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.TargetHeight]
		if !has {
			log.Println(ob.obID, "br, has := ob.round.BlockRoundMap[msg.TargetHeight]")
			return errors.WithStack(ErrInvalidVote)
		}

		if msg.TargetHeight != ob.round.TargetHeight {
			if msg.TargetHeight < ob.round.TargetHeight {
				if SenderPublicKey != ob.myPublicKey {
					ob.sendStatusTo(SenderPublicKey)
				}
			}
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.ChainID.Cmp(cp.ChainID()) != 0 {
			log.Println(ob.obID, "if msg.ChainID != cp.ChainID() {")
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.LastHash != cp.LastHash() {
			log.Println(ob.obID, "if msg.LastHash != cp.LastHash() {")
			return errors.WithStack(ErrInvalidVote)
		}
		Top, err := ob.cn.TopGenerator(msg.TimeoutCount)
		if err != nil {
			return err
		}
		if msg.Generator != Top {
			log.Println(ob.obID, "if msg.Generator != Top {")
			return errors.WithStack(ErrInvalidVote)
		}

		//[check state]
		if ob.round.RoundState == BlockVoteState {
			if _, has := br.BlockVoteMap[SenderPublicKey]; !has {
				ob.sendBlockGenTo(br.BlockGenMessage, SenderPublicKey)
			}
			ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicKey)
		}
	case *BlockVoteMessage:
		if DEBUG {
			log.Println(ob.obID, cp.Height(), bin.MustWriterToHash(msg.Header), "BlockVoteMessage", ob.round.RoundState, msg.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		}
		if !ob.observerKeyMap[SenderPublicKey] {
			return errors.WithStack(ErrInvalidObserverKey)
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.Header.Height]
		if !has {
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.Header.Height != ob.round.TargetHeight {
			if msg.Header.Height < ob.round.TargetHeight {
				ob.sendStatusTo(SenderPublicKey)
			} else {
				br.BlockVoteMessageWaitMap[SenderPublicKey] = msg
			}
			return errors.WithStack(ErrInvalidVote)
		}

		//[check state]
		if ob.round.RoundState != BlockVoteState {
			if ob.round.RoundState < BlockVoteState {
				if _, has := ob.ignoreMap[msg.Header.Generator]; has {
					delete(ob.ignoreMap, msg.Header.Generator)
					ob.round.VoteFailCount = 0
				}
				br.BlockVoteMessageWaitMap[SenderPublicKey] = msg

				if ob.round.RoundState == BlockWaitState && br.BlockGenMessageWait == nil && br.BlockGenMessage == nil {
					ob.sendBlockGenRequest(br)
				}
			}
			return errors.WithStack(ErrInvalidVote)
		}
		Top, err := ob.cn.TopGenerator(msg.Header.TimeoutCount)
		if err != nil {
			return err
		}
		if msg.Header.Generator != Top {
			log.Println(ob.obID, "Observer", cp.Height(), ob.myPublicKey.Address().String(), Top.String(), bin.MustWriterToHash(msg.Header), "if msg.Header.Generator != Top.Address {")
			return errors.WithStack(ErrInvalidVote)
		}
		bh := bin.MustWriterToHash(msg.Header)
		pubkey, err := common.RecoverPubkey(ob.ChainID, bh, msg.GeneratorSignature)
		if err != nil {
			return err
		}
		Signer := pubkey.Address()
		if Signer != Top {
			log.Println(ob.obID, bin.MustWriterToHash(msg.Header), "if Signer != Top.PublicKey {")
			return errors.WithStack(ErrInvalidTopSignature)
		}
		if msg.Header.Generator != ob.round.MinRoundVoteAck.Generator {
			log.Println(ob.obID, bin.MustWriterToHash(msg.Header), "if msg.Header.Generator != ob.round.MinRoundVoteAck.Generator {")
			return errors.WithStack(ErrInvalidVote)
		}
		if Signer != ob.round.MinRoundVoteAck.Generator {
			log.Println(ob.obID, bin.MustWriterToHash(msg.Header), "if Signer != ob.round.MinRoundVoteAck.GeneratorPublicKey {")
			return errors.WithStack(ErrInvalidVote)
		}
		if msg.Header.PrevHash != cp.LastHash() {
			log.Println(ob.obID, bin.MustWriterToHash(msg.Header), "if msg.Header.PrevHash != cp.LastHash() {")
			return errors.WithStack(ErrInvalidVote)
		}
		if bh != bin.MustWriterToHash(&br.BlockGenMessage.Block.Header) {
			log.Println(ob.obID, bin.MustWriterToHash(msg.Header), "if bh != bin.MustWriterToHash(&br.BlockGenMessage.Block.Header) {")
			return errors.WithStack(ErrInvalidVote)
		}
		if err := ob.ct.ValidateHeader(msg.Header); err != nil {
			log.Println(ob.obID, msg.Header.Generator.String(), "if err := ob.ct.ValidateHeader(msg.Header); err != nil {")
			return err
		}

		s := &types.BlockSign{
			HeaderHash:         bh,
			GeneratorSignature: msg.GeneratorSignature,
		}
		if pubkey, err := common.RecoverPubkey(ob.ChainID, bin.MustWriterToHash(s), msg.ObserverSignature); err != nil {
			log.Println(ob.obID, msg.Header.Generator.String(), "common.RecoverPubkey(ob.ChainID, bin.MustWriterToHash(s), msg.ObserverSignature)")
			return err
		} else if SenderPublicKey != pubkey {
			log.Println(ob.obID, msg.Header.Generator.String(), "SenderPublicKey != pubkey")
			return errors.WithStack(ErrInvalidVote)
		}

		if _, has := br.BlockVoteMap[SenderPublicKey]; has {
			if !msg.IsReply {
				ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicKey)
			}
			return errors.WithStack(ErrAlreadyVoted)
		}
		br.BlockVoteMap[SenderPublicKey] = msg

		if DEBUG {
			log.Println(ob.obID, "Observer", cp.Height(), bin.MustWriterToHash(msg.Header), "BlockVoteMessage", ob.round.RoundState, msg.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
		}

		//[check state]
		if !msg.IsReply && SenderPublicKey != ob.myPublicKey {
			ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicKey)
		}

		//[apply vote]
		if len(br.BlockVoteMap) >= len(ob.observerKeyMap)/2+1 {
			sigs := []common.Signature{}
			for _, vt := range br.BlockVoteMap {
				sigs = append(sigs, vt.ObserverSignature)
			}

			PastTime := uint64(time.Now().UnixNano()) - ob.roundFirstTime
			ExpectedTime := uint64(msg.Header.Height-ob.roundFirstHeight) * uint64(BlockTime)
			if PastTime < ExpectedTime {
				diff := time.Duration(ExpectedTime - PastTime)
				if diff > BlockTime {
					diff = BlockTime
				}
				time.Sleep(diff)
			}

			b := &types.Block{
				Header: br.BlockGenMessage.Block.Header,
				Body: types.Body{
					Transactions:          br.BlockGenMessage.Block.Body.Transactions,
					TransactionSignatures: br.BlockGenMessage.Block.Body.TransactionSignatures,
					Events:                br.BlockGenMessage.Block.Body.Events,
					BlockSignatures:       append([]common.Signature{br.BlockGenMessage.GeneratorSignature}, sigs...),
				},
			}
			if err := ob.ct.ConnectBlockWithContext(b, br.Context); err != nil {
				return err
			} else {
				ob.broadcastStatus()
			}
			delete(ob.ignoreMap, ob.round.MinRoundVoteAck.Generator)

			adjustMap := ob.adjustGeneratorMap()
			delete(adjustMap, ob.round.MinRoundVoteAck.Generator)
			var NextTop common.Address
			if len(adjustMap) > 0 {
				r, _, err := ob.cn.TopGeneratorInMap(adjustMap)
				if err != nil {
					return err
				}
				NextTop = r
			}
			if ob.round.MinRoundVoteAck.PublicKey == ob.myPublicKey {
				nm := &BlockObSignMessage{
					TargetHeight: msg.Header.Height,
					BlockSign: &types.BlockSign{
						HeaderHash:         bh,
						GeneratorSignature: msg.GeneratorSignature,
					},
					ObserverSignatures: sigs,
				}
				bs := p2p.MessageToPacket(nm)
				ob.sendMessagePacket(0, ob.round.MinRoundVoteAck.Generator, bs)
				//ob.fs.SendTo(ob.round.MinRoundVoteAck.Generator, bs)
				if NextTop != common.ZeroAddr {
					ob.sendMessagePacket(0, NextTop, bs)
					//ob.fs.SendTo(NextTop.Address, bs)
				}
				bs = nil
			} else {
				if NextTop != common.ZeroAddr {
					delete(adjustMap, NextTop)
				}
				if len(adjustMap) > 0 {
					ranks, err := ob.cn.GeneratorsInMap(adjustMap, 3)
					if err == nil {
						for _, v := range ranks {
							ob.sendMessage(1, v, &p2p.StatusMessage{
								Version:  b.Header.Version,
								Height:   b.Header.Height,
								LastHash: bh,
							})
						}
					}
				}
			}
			if DEBUG {
				log.Println(ob.obID, "Observer", cp.Height(), "BlockConnected", b.Header.Generator.String(), ob.round.RoundState, msg.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}

			NextHeight := ob.round.TargetHeight + 1
			Top, err := ob.cn.TopGenerator(0)
			if err != nil {
				return err
			}
			brNext, has := ob.round.BlockRoundMap[NextHeight]
			if has && Top == ob.round.MinRoundVoteAck.Generator {
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
		Height := ob.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		bs, err := p2p.BlockPacketWithCache(msg, ob.cn.Provider(), ob.batchCache, ob.singleCache)
		if err != nil {
			return err
		}
		if err := ob.ms.SendTo(SenderPublicKey, bs); err != nil {
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
					ob.sendRequestBlockTo(SenderPublicKey, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
						if !ob.requestTimer.Exist(i) {
							ob.sendRequestBlockTo(SenderPublicKey, i, 1)
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
				log.Println(ob.obID, SenderPublicKey.Address().String(), h.String(), msg.LastHash.String(), msg.Height)
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
		return errors.WithStack(p2p.ErrUnknownMessage)
	}
	return nil
}
