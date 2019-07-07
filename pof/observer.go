package pof

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/framework/chain"
)

type messageItem struct {
	PublicHash common.PublicHash
	Message    interface{}
	Raw        []byte
}

type Observer struct {
	sync.Mutex
	cs               *Consensus
	round            *VoteRound
	ignoreMap        map[common.Address]int64
	myPublicHash     common.PublicHash
	roundFirstTime   uint64
	roundFirstHeight uint32
	messageQueue     *queue.Queue
}

func NewObserver(cs *Consensus, MyPublicHash common.PublicHash) *Observer {
	return &Observer{
		cs:           cs,
		round:        NewVoteRound(cs.cn.Provider().Height()+1, cs.maxBlocksPerFormulator-cs.blocksBySameFormulator),
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: MyPublicHash,
		messageQueue: queue.NewQueue(),
	}
}

func (ob *Observer) syncVoteRound() {
	cp := ob.cs.cn.Provider()
	TargetHeight := cp.Height() + 1
	if ob.round.TargetHeight < TargetHeight {
		Diff := TargetHeight - ob.round.TargetHeight
		if ob.round.RemainBlocks > Diff {
			ob.round.TargetHeight = TargetHeight
			ob.round.RemainBlocks = ob.round.RemainBlocks - Diff
		} else {
			ob.resetVoteRound(false)
		}
	}
}

func (ob *Observer) resetVoteRound(resetStat bool) {
	ob.round = NewVoteRound(ob.cs.cn.Provider().Height()+1, ob.cs.maxBlocksPerFormulator-ob.cs.blocksBySameFormulator)
	if resetStat {
		ob.roundFirstTime = 0
		ob.roundFirstHeight = 0
	}
}

func (ob *Observer) Run() {
	queueTimer := time.NewTimer(time.Millisecond)
	voteTimer := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-queueTimer.C:
			v := ob.messageQueue.Pop()
			i := 0
			for v != nil {
				i++
				item := v.(*messageItem)
				ob.Lock()
				ob.handleObserverMessage(item.PublicHash, item.Message)
				ob.Unlock()
				v = ob.messageQueue.Pop()
			}
			queueTimer.Reset(10 * time.Millisecond)
		case <-voteTimer.C:
			ob.Lock()
			ob.syncVoteRound()
			if len(ob.adjustFormulatorMap()) > 0 {
				IsFailable := true
				if ob.round.RoundState == RoundVoteState {
					//ob.sendRoundVote()
				} else if ob.round.RoundState == BlockVoteState {
					IsFailable = false
				}
				if IsFailable {
					ob.round.VoteFailCount++
					if ob.round.VoteFailCount > 20 {
						if ob.round.MinRoundVoteAck != nil {
							addr := ob.round.MinRoundVoteAck.Formulator
							_, has := ob.ignoreMap[addr]
							if has {
								//ob.fs.RemovePeer(addr)
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(120*time.Second)
							} else {
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(30*time.Second)
							}
						}
						//ob.sendRoundVote()
					}
				}
			}
			ob.Unlock()

			voteTimer.Reset(100 * time.Millisecond)
			//case <-ob.runEnd:
			//return
		}
	}
}

func (ob *Observer) handleObserverMessage(SenderPublicHash common.PublicHash, m interface{}) error {
	cp := ob.cs.cn.Provider()

	msgh := encoding.Hash(m)

	ob.syncVoteRound()

	switch msg := m.(type) {
	case *RoundVoteMessage:
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signautre); err != nil {
			return err
		} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
			return common.ErrInvalidPublicHash
		} else if !ob.cs.observerKeyMap.Has(obkey) {
			return ErrInvalidObserverKey
		}

		//[check round]
		if msg.RoundVote.TargetHeight != ob.round.TargetHeight {
			if msg.RoundVote.TargetHeight < ob.round.TargetHeight {
				if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
					//ob.sendStateTo(SenderPublicHash)
				} else {
					ob.resetVoteRound(true)
					//ob.sendRoundVote()
				}
			}
			return ErrInvalidVote
		}
		if msg.RoundVote.RemainBlocks != ob.round.RemainBlocks {
			return ErrInvalidVote
		}
		if msg.RoundVote.LastHash != cp.LastHash() {
			return ErrInvalidVote
		}
		Top, err := ob.cs.rt.TopRank(int(msg.RoundVote.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.RoundVote.Formulator != Top.Address {
			return ErrInvalidVote
		}
		if msg.RoundVote.FormulatorPublicHash != Top.PublicHash {
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != RoundVoteState {
			if SenderPublicHash != ob.myPublicHash {
				if !msg.RoundVote.IsReply {
					if ob.round.RoundState == BlockVoteState {
						br, has := ob.round.BlockRoundMap[msg.RoundVote.TargetHeight]
						if !has {
							return ErrInvalidVote
						}
						log.Println(br)
						//ob.sendBlockVoteTo(br, SenderPublicHash)
					}
					//ob.sendRoundVoteTo(SenderPublicHash)
				}
			}
			return ErrInvalidRoundState
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteMessageMap[SenderPublicHash]; has {
			if msg.RoundVote.Timestamp <= old.RoundVote.Timestamp {
				if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
					//ob.sendRoundVoteTo(SenderPublicHash)
				}
				return ErrAlreadyVoted
			}
		}
		ob.round.RoundVoteMessageMap[SenderPublicHash] = msg

		if !msg.RoundVote.IsReply && SenderPublicHash != ob.myPublicHash {
			//ob.sendRoundVoteTo(SenderPublicHash)
		}
		if len(ob.round.RoundVoteMessageMap) >= ob.cs.observerKeyMap.Len()/2+2 {
			votes := []*voteSortItem{}
			for pubhash, msg := range ob.round.RoundVoteMessageMap {
				votes = append(votes, &voteSortItem{
					PublicHash: pubhash,
					Priority:   uint64(msg.RoundVote.TimeoutCount),
				})
			}
			sort.Sort(voteSorter(votes))

			ob.round.RoundState = RoundVoteAckState
			ob.round.MinVotePublicHash = votes[0].PublicHash

			if ob.roundFirstTime == 0 {
				ob.roundFirstTime = uint64(time.Now().UnixNano())
				ob.roundFirstHeight = uint32(cp.Height())
			}

			//ob.sendRoundVoteAck()

			for pubhash, msg := range ob.round.RoundVoteAckMessageWaitMap {
				ob.messageQueue.Push(&messageItem{
					PublicHash: pubhash,
					Message:    msg,
				})
			}
		}
	case *RoundVoteAckMessage:
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signautre); err != nil {
			return err
		} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
			return common.ErrInvalidPublicHash
		} else if !ob.cs.observerKeyMap.Has(obkey) {
			return ErrInvalidObserverKey
		}

		//[check round]
		if msg.RoundVoteAck.TargetHeight != ob.round.TargetHeight {
			if msg.RoundVoteAck.TargetHeight < ob.round.TargetHeight {
				if !msg.RoundVoteAck.IsReply && SenderPublicHash != ob.myPublicHash {
					//ob.sendStateTo(SenderPublicHash)
				} else {
					ob.resetVoteRound(true)
					//ob.sendRoundVote()
				}
			}
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.RemainBlocks != ob.round.RemainBlocks {
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.LastHash != cp.LastHash() {
			return ErrInvalidVote
		}
		Top, err := ob.cs.rt.TopRank(int(msg.RoundVoteAck.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.RoundVoteAck.Formulator != Top.Address {
			return ErrInvalidVote
		}
		if msg.RoundVoteAck.FormulatorPublicHash != Top.PublicHash {
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
						if !has {
							return ErrInvalidVote
						}
						log.Println(br)
						//ob.sendBlockVoteTo(br, SenderPublicHash)
					}
					//ob.sendRoundVoteAckTo(SenderPublicHash)
				}
			}
			return ErrInvalidRoundState
		}

		//[apply vote]
		if old, has := ob.round.RoundVoteAckMessageMap[SenderPublicHash]; has {
			if msg.RoundVoteAck.Timestamp <= old.RoundVoteAck.Timestamp {
				if !msg.RoundVoteAck.IsReply {
					//ob.sendRoundVoteTo(SenderPublicHash)
				}
				return ErrAlreadyVoted
			}
		}
		ob.round.RoundVoteAckMessageMap[SenderPublicHash] = msg

		if !msg.RoundVoteAck.IsReply && SenderPublicHash != ob.myPublicHash {
			//ob.sendRoundVoteAckTo(SenderPublicHash)
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

				if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
					nm := &BlockReqMessage{
						PrevHash:             ob.round.MinRoundVoteAck.LastHash,
						TargetHeight:         ob.round.MinRoundVoteAck.TargetHeight,
						RemainBlocks:         ob.round.MinRoundVoteAck.RemainBlocks,
						TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
						Formulator:           ob.round.MinRoundVoteAck.Formulator,
						FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
					}
					log.Println(nm)
					//ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, nm)
				}
				br := ob.round.BlockRoundMap[ob.round.TargetHeight]
				if br != nil {
					if br.BlockGenMessageWait != nil && br.BlockGenMessage == nil {
						ob.messageQueue.Push(&messageItem{
							Message: br.BlockGenMessageWait,
						})
						br.BlockGenMessageWait = nil
					}
				}
			}
		}
	case *BlockGenMessage:
		//[check round]
		br, has := ob.round.BlockRoundMap[msg.Block.Header.Height]
		if !has {
			return ErrInvalidVote
		}
		if br.BlockGenMessage != nil {
			return ErrInvalidVote
		}

		TimeoutCount, err := ob.cs.decodeConsensusData(msg.Block.Header.ConsensusData)
		if err != nil {
			return err
		}
		Top, err := ob.cs.rt.TopRank(int(TimeoutCount))
		if err != nil {
			return err
		}
		if msg.Block.Header.Generator != Top.Address {
			return ErrInvalidVote
		}
		bh := encoding.Hash(msg.Block.Header)
		pubkey, err := common.RecoverPubkey(bh, msg.GeneratorSignature)
		if err != nil {
			return err
		}
		Signer := common.NewPublicHash(pubkey)
		if Signer != Top.PublicHash {
			return ErrInvalidTopSignature
		}

		if br.BlockGenMessageWait != nil {
			if bh != encoding.Hash(br.BlockGenMessageWait.Block.Header) {
				return ErrFoundForkedBlockGen
			}
		}

		if msg.Block.Header.Height != ob.round.TargetHeight {
			if msg.Block.Header.Height > ob.round.TargetHeight {
				br.BlockGenMessageWait = msg
			}
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState < BlockWaitState {
			br.BlockGenMessageWait = msg
			return ErrInvalidVote
		}
		if TimeoutCount != ob.round.MinRoundVoteAck.TimeoutCount {
			return ErrInvalidVote
		}
		if msg.Block.Header.Generator != ob.round.MinRoundVoteAck.Formulator {
			return ErrInvalidVote
		}
		if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {
			return ErrInvalidVote
		}
		if msg.Block.Header.PrevHash != ob.round.MinRoundVoteAck.LastHash {
			return ErrInvalidVote
		}
		if msg.Block.Header.PrevHash != cp.LastHash() {
			return ErrInvalidVote
		}

		//[apply vote]
		if !msg.IsReply && SenderPublicHash != ob.myPublicHash && ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
			//ob.sendBlockGen(msg)
		}

		//[if valid block]
		Now := uint64(time.Now().UnixNano())
		if msg.Block.Header.Timestamp > Now+uint64(10*time.Second) {
			return ErrInvalidVote
		}
		if msg.Block.Header.Timestamp < cp.LastTimestamp() {
			return ErrInvalidVote
		}

		ctx := ob.cs.ct.NewContext()
		if err := ob.cs.ct.ExecuteBlockOnContext(msg.Block, ctx); err != nil {
			return err
		}
		ob.round.RoundState = BlockVoteState
		br.BlockGenMessage = msg
		br.Context = ctx

		//ob.sendBlockVote(br)

		for pubhash, msg := range br.BlockVoteMessageWaitMap {
			ob.messageQueue.Push(&messageItem{
				PublicHash: pubhash,
				Message:    msg,
			})
		}
	case *BlockVoteMessage:
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signautre); err != nil {
			return err
		} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
			return common.ErrInvalidPublicHash
		} else if !ob.cs.observerKeyMap.Has(obkey) {
			return ErrInvalidObserverKey
		}

		//[check round]
		br, has := ob.round.BlockRoundMap[msg.BlockVote.Header.Height]
		if !has {
			return ErrInvalidVote
		}
		TimeoutCount, err := ob.cs.decodeConsensusData(msg.BlockVote.Header.ConsensusData)
		if err != nil {
			return err
		}
		Top, err := ob.cs.rt.TopRank(int(TimeoutCount))
		if err != nil {
			return err
		}
		if msg.BlockVote.Header.Generator != Top.Address {
			return ErrInvalidVote
		}
		bh := encoding.Hash(msg.BlockVote.Header)
		pubkey, err := common.RecoverPubkey(bh, msg.BlockVote.GeneratorSignature)
		if err != nil {
			return err
		}
		Signer := common.NewPublicHash(pubkey)
		if Signer != Top.PublicHash {
			return ErrInvalidTopSignature
		}
		if msg.BlockVote.Header.Height != ob.round.TargetHeight {
			if msg.BlockVote.Header.Height < ob.round.TargetHeight {
				//ob.sendStateTo(SenderPublicHash)
			} else {
				br.BlockVoteMessageWaitMap[SenderPublicHash] = msg
			}
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState < BlockVoteState {
			if _, has := ob.ignoreMap[msg.BlockVote.Header.Generator]; has {
				delete(ob.ignoreMap, msg.BlockVote.Header.Generator)
				ob.round.VoteFailCount = 0
			}
			br.BlockVoteMessageWaitMap[SenderPublicHash] = msg
			return ErrInvalidVote
		}
		if TimeoutCount != ob.round.MinRoundVoteAck.TimeoutCount {
			return ErrInvalidVote
		}
		if msg.BlockVote.Header.Generator != ob.round.MinRoundVoteAck.Formulator {
			return ErrInvalidVote
		}
		if Signer != ob.round.MinRoundVoteAck.FormulatorPublicHash {
			return ErrInvalidVote
		}
		if msg.BlockVote.Header.PrevHash != ob.round.MinRoundVoteAck.LastHash {
			return ErrInvalidVote
		}
		if msg.BlockVote.Header.PrevHash != cp.LastHash() {
			return ErrInvalidVote
		}

		s := &types.BlockSign{
			HeaderHash:         bh,
			GeneratorSignature: msg.BlockVote.GeneratorSignature,
		}
		if pubkey, err := common.RecoverPubkey(encoding.Hash(s), msg.BlockVote.ObserverSignature); err != nil {
			return err
		} else {
			if SenderPublicHash != common.NewPublicHash(pubkey) {
				return ErrInvalidVote
			}
		}

		if _, has := br.BlockVoteMap[SenderPublicHash]; has {
			if !msg.BlockVote.IsReply {
				//ob.sendBlockVoteTo(br, SenderPublicHash)
			}
			return ErrAlreadyVoted
		}
		br.BlockVoteMap[SenderPublicHash] = msg.BlockVote

		//[check state]
		if !msg.BlockVote.IsReply && SenderPublicHash != ob.myPublicHash {
			//ob.sendBlockVoteTo(br, SenderPublicHash)
		}

		//[apply vote]
		if len(br.BlockVoteMap) >= ob.cs.observerKeyMap.Len()/2+1 {
			sigs := []common.Signature{}
			for _, vt := range br.BlockVoteMap {
				sigs = append(sigs, vt.ObserverSignature)
			}

			b := &types.Block{
				Header:               br.BlockGenMessage.Block.Header,
				TransactionTypes:     br.BlockGenMessage.Block.TransactionTypes,
				Transactions:         br.BlockGenMessage.Block.Transactions,
				TranactionSignatures: br.BlockGenMessage.Block.TranactionSignatures,
				Signatures:           append([]common.Signature{br.BlockGenMessage.GeneratorSignature}, sigs...),
			}

			if err := ob.cs.ct.ConnectBlockWithContext(b, br.Context); err != nil {
				if err != chain.ErrInvalidHeight {
					return err
				}
			} else {
				//ob.cm.BroadcastHeader(cd.Header)
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
				log.Println(nm)
				//ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, nm)
				//ob.fs.UpdateGuessHeight(ob.round.MinRoundVoteAck.Formulator, nm.TargetHeight)

				if NextTop != nil && NextTop.Address != ob.round.MinRoundVoteAck.Formulator {
					/*
						ob.fs.SendTo(NextTop.Address, &chain.StatusMessage{
							Version:  cd.Header.Version(),
							Height:   cd.Header.Height(),
							LastHash: cd.Header.Hash(),
						})
					*/
				}
			} else {
				if NextTop != nil {
					delete(adjustMap, NextTop.Address)
					//ob.fs.UpdateGuessHeight(NextTop.Address, cd.Header.Height())
				}
				if len(adjustMap) > 0 {
					ranks, err := ob.cs.rt.RanksInMap(adjustMap, 3)
					if err == nil {
						for i, v := range ranks {
							/*
								ob.fs.SendTo(v.Address, &chain.StatusMessage{
									Version:  cd.Header.Version(),
									Height:   cd.Header.Height(),
									LastHash: cd.Header.Hash(),
								})
							*/
							log.Println(i, v)
						}
					}
				}
			}

			PastTime := uint64(time.Now().UnixNano()) - ob.roundFirstTime
			ExpectedTime := uint64(msg.BlockVote.Header.Height-ob.roundFirstHeight) * uint64(500*time.Millisecond)

			if PastTime < ExpectedTime {
				diff := time.Duration(ExpectedTime - PastTime)
				if diff > 500*time.Millisecond {
					diff = 500 * time.Millisecond
				}
				time.Sleep(diff)
				//ob.kn.DebugLog("Observer", ob.kn.Provider().Height(), "Sleep", int64(ExpectedTime-PastTime)/int64(time.Millisecond))
			}

			if ob.round.RemainBlocks > 1 {
				ob.round.RoundState = BlockWaitState
				delete(ob.round.BlockRoundMap, ob.round.TargetHeight)
				ob.round.RemainBlocks--
				ob.round.TargetHeight++
				ob.round.VoteFailCount = 0
				brNext, has := ob.round.BlockRoundMap[ob.round.TargetHeight]
				if has {
					if brNext.BlockGenMessageWait != nil && brNext.BlockGenMessage == nil {
						ob.messageQueue.Push(&messageItem{
							Message: brNext.BlockGenMessageWait,
						})
					}
				}
			} else {
				ob.resetVoteRound(false)
				//.ob.prevRoundEndTime = time.Now().UnixNano()
				//ob.sendRoundVote()
			}
		}
	}
	return nil
}

func (ob *Observer) adjustFormulatorMap() map[common.Address]bool {
	//FormulatorMap := ob.fs.FormulatorMap()
	FormulatorMap := map[common.Address]bool{} //TEMP
	now := time.Now().UnixNano()
	for addr := range FormulatorMap {
		if now < ob.ignoreMap[addr] {
			delete(FormulatorMap, addr)
		}
	}
	return FormulatorMap
}
