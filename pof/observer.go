package pof

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

type messageItem struct {
	PublicHash common.PublicHash
	Message    interface{}
	Raw        []byte
}

type Observer struct {
	sync.Mutex
	key              key.Key
	ms               *ObserverMesh
	fs               *FormulatorService
	cs               *Consensus
	round            *VoteRound
	ignoreMap        map[common.Address]int64
	myPublicHash     common.PublicHash
	roundFirstTime   uint64
	roundFirstHeight uint32
	messageQueue     *queue.Queue
	requestTimer     *p2p.RequestTimer
	blockQ           *queue.SortedQueue
	isRunning        bool
	closeLock        sync.RWMutex
	runEnd           chan struct{}
	isClose          bool

	prevRoundEndTime int64 // FOR DEBUG
}

func NewObserver(key key.Key, NetAddressMap map[common.PublicHash]string, cs *Consensus, MyPublicHash common.PublicHash) *Observer {
	ob := &Observer{
		key:          key,
		cs:           cs,
		round:        NewVoteRound(cs.cn.Provider().Height()+1, cs.maxBlocksPerFormulator-cs.blocksBySameFormulator),
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: MyPublicHash,
		messageQueue: queue.NewQueue(),
		blockQ:       queue.NewSortedQueue(),
	}
	ob.ms = NewObserverMesh(key, NetAddressMap, ob)
	ob.fs = NewFormulatorService(ob)
	ob.requestTimer = p2p.NewRequestTimer(ob)
	return ob
}

// Init initializes observer
func (ob *Observer) Init() error {
	fc := encoding.Factory("pof.message")
	fc.Register(types.DefineHashedType("pof.RoundVoteMessage"), &RoundVoteMessage{})
	fc.Register(types.DefineHashedType("pof.RoundVoteAckMessage"), &RoundVoteAckMessage{})
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenMessage"), &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockVoteMessage"), &BlockVoteMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	return nil
}

// Close terminates the observer
func (ob *Observer) Close() {
	ob.closeLock.Lock()
	defer ob.closeLock.Unlock()

	ob.Lock()
	defer ob.Unlock()

	ob.isClose = true
	ob.cs.cn.Close()
	ob.runEnd <- struct{}{}
}

// Run starts the pof consensus on the observer
func (ob *Observer) Run(BindObserver string, BindFormulator string) {
	ob.Lock()
	if ob.isRunning {
		ob.Unlock()
		return
	}
	ob.isRunning = true
	ob.Unlock()

	go ob.ms.Run(BindObserver)
	go ob.fs.Run(BindFormulator)

	blockTimer := time.NewTimer(time.Millisecond)
	queueTimer := time.NewTimer(time.Millisecond)
	voteTimer := time.NewTimer(time.Millisecond)
	for !ob.isClose {
		select {
		case <-blockTimer.C:
			cp := ob.cs.cn.Provider()
			ob.Lock()
			TargetHeight := uint64(cp.Height() + 1)
			item := ob.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := ob.cs.cn.ConnectBlock(b); err != nil {
					break
				}
				ob.broadcastStatus()
				TargetHeight++
				item = ob.blockQ.PopUntil(TargetHeight)
			}
			ob.Unlock()
			blockTimer.Reset(50 * time.Millisecond)
		case <-queueTimer.C:
			v := ob.messageQueue.Pop()
			i := 0
			for v != nil {
				i++
				item := v.(*messageItem)
				ob.Lock()
				ob.handleObserverMessage(item.PublicHash, item.Message, item.Raw)
				ob.Unlock()
				v = ob.messageQueue.Pop()
			}
			queueTimer.Reset(10 * time.Millisecond)
		case <-voteTimer.C:
			ob.Lock()
			cp := ob.cs.cn.Provider()
			ob.syncVoteRound()
			IsFailable := true
			if len(ob.adjustFormulatorMap()) > 0 {
				log.Println("Observer", cp.Height(), "Current State", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				if ob.round.RoundState == RoundVoteState {
					ob.sendRoundVote()
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
								ob.fs.RemovePeer(addr)
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(120*time.Second)
							} else {
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(30*time.Second)
							}
						}
						ob.resetVoteRound(true)
						ob.sendRoundVote()
					}
				}
			} else {
				log.Println("Observer", cp.Height(), "No Formulator", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}
			ob.Unlock()

			voteTimer.Reset(100 * time.Millisecond)
		case <-ob.runEnd:
			return
		}
	}
}

// OnTimerExpired called when rquest expired
func (ob *Observer) OnTimerExpired(height uint32, P interface{}) {
	TargetPubHash := P.(common.PublicHash)
	list := ob.ms.Peers()
	for _, v := range list {
		if v.pubhash != ob.myPublicHash && v.pubhash != TargetPubHash {
			ob.sendRequestBlockTo(v.pubhash, height)
			break
		}
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
	ob.prevRoundEndTime = time.Now().UnixNano()
	if resetStat {
		ob.roundFirstTime = 0
		ob.roundFirstHeight = 0
	}
}

func (ob *Observer) onObserverRecv(p *Peer, m interface{}, raw []byte) error {
	ob.messageQueue.Push(&messageItem{
		PublicHash: p.pubhash,
		Message:    m,
	})
	return nil
}

func (ob *Observer) onFormulatorRecv(p *FormulatorPeer, m interface{}, raw []byte) error {
	if msg, is := m.(*p2p.RequestMessage); is {
		ob.Lock()
		defer ob.Unlock()

		enable := false

		if p.GuessHeight() < msg.Height {
			cp := ob.cs.cn.Provider()
			CountMap := ob.fs.GuessHeightCountMap()
			if CountMap[cp.Height()] < 3 {
				enable = true
			} else {
				ranks, err := ob.cs.rt.RanksInMap(ob.adjustFormulatorMap(), 5)
				if err != nil {
					return err
				}
				rankMap := map[string]bool{}
				for _, r := range ranks {
					rankMap[r.Address.String()] = true
				}
				enable = rankMap[p.ID()]
			}
			if enable {
				p.UpdateGuessHeight(msg.Height)

				b, err := cp.Block(msg.Height)
				if err != nil {
					return err
				}
				sm := &p2p.BlockMessage{
					Block: b,
				}
				if err := p.Send(sm); err != nil {
					return err
				}
			}
		}
	} else if msg, is := m.(*BlockGenMessage); is {
		ob.messageQueue.Push(&messageItem{
			Message: msg,
			Raw:     raw,
		})
	} else {
		panic(ErrUnknownMessage) //TEMP
		return ErrUnknownMessage
	}
	return nil
}

func (ob *Observer) handleObserverMessage(SenderPublicHash common.PublicHash, m interface{}, raw []byte) error {
	cp := ob.cs.cn.Provider()

	ob.syncVoteRound()

	switch msg := m.(type) {
	case *RoundVoteMessage:
		msgh := encoding.Hash(msg.RoundVote)
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signature); err != nil {
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
					ob.sendStatusTo(SenderPublicHash)
				} else {
					ob.resetVoteRound(true)
					ob.sendRoundVote()
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
						ob.sendBlockVoteTo(br, SenderPublicHash)
					}
					ob.sendRoundVoteTo(SenderPublicHash)
				}
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

			ob.sendRoundVoteAck()

			for pubhash, msg := range ob.round.RoundVoteAckMessageWaitMap {
				ob.messageQueue.Push(&messageItem{
					PublicHash: pubhash,
					Message:    msg,
				})
			}
		}
	case *RoundVoteAckMessage:
		msgh := encoding.Hash(msg.RoundVoteAck)
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signature); err != nil {
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
					ob.sendStatusTo(SenderPublicHash)
				} else {
					ob.resetVoteRound(true)
					ob.sendRoundVote()
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
						ob.sendBlockVoteTo(br, SenderPublicHash)
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
					ob.sendRoundVoteTo(SenderPublicHash)
				}
				return ErrAlreadyVoted
			}
		}
		ob.round.RoundVoteAckMessageMap[SenderPublicHash] = msg

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

				if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
					nm := &BlockReqMessage{
						PrevHash:             ob.round.MinRoundVoteAck.LastHash,
						TargetHeight:         ob.round.MinRoundVoteAck.TargetHeight,
						RemainBlocks:         ob.round.MinRoundVoteAck.RemainBlocks,
						TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
						Formulator:           ob.round.MinRoundVoteAck.Formulator,
						FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
					}
					ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, nm)
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
		if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash && len(raw) > 0 {
			ob.ms.BroadcastRaw(raw)
		}

		//[apply vote]
		if !msg.IsReply && SenderPublicHash != ob.myPublicHash && ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash {
			ob.sendBlockVoteTo(br, SenderPublicHash)
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

		ob.sendBlockVote(br)

		for pubhash, msg := range br.BlockVoteMessageWaitMap {
			ob.messageQueue.Push(&messageItem{
				PublicHash: pubhash,
				Message:    msg,
			})
		}
	case *BlockVoteMessage:
		msgh := encoding.Hash(msg.BlockVote)
		if pubkey, err := common.RecoverPubkey(msgh, msg.Signature); err != nil {
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
				ob.sendStatusTo(SenderPublicHash)
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
				ob.sendBlockVoteTo(br, SenderPublicHash)
			}
			return ErrAlreadyVoted
		}
		br.BlockVoteMap[SenderPublicHash] = msg.BlockVote

		//[check state]
		if !msg.BlockVote.IsReply && SenderPublicHash != ob.myPublicHash {
			ob.sendBlockVoteTo(br, SenderPublicHash)
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
				log.Println(nm)
				ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, nm)
				ob.fs.UpdateGuessHeight(ob.round.MinRoundVoteAck.Formulator, nm.TargetHeight)

				if NextTop != nil && NextTop.Address != ob.round.MinRoundVoteAck.Formulator {
					ob.fs.SendTo(NextTop.Address, &p2p.StatusMessage{
						Version:  b.Header.Version,
						Height:   b.Header.Height,
						LastHash: bh,
					})
				}
			} else {
				if NextTop != nil {
					delete(adjustMap, NextTop.Address)
					ob.fs.UpdateGuessHeight(NextTop.Address, b.Header.Height)
				}
				if len(adjustMap) > 0 {
					ranks, err := ob.cs.rt.RanksInMap(adjustMap, 3)
					if err == nil {
						for _, v := range ranks {
							ob.fs.SendTo(v.Address, &p2p.StatusMessage{
								Version:  b.Header.Version,
								Height:   b.Header.Height,
								LastHash: bh,
							})
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
				ob.sendRoundVote()
			}
		}
	case *p2p.RequestMessage:
		b, err := cp.Block(msg.Height)
		if err != nil {
			return err
		}
		sm := &p2p.BlockMessage{
			Block: b,
		}
		if err := ob.ms.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		Height := cp.Height()
		if Height < msg.Height {
			for i := Height + 1; i <= Height+100 && i <= msg.Height; i++ {
				if !ob.requestTimer.Exist(i) {
					ob.sendRequestBlockTo(SenderPublicHash, i)
				}
			}
		} else {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *p2p.BlockMessage:
		if err := ob.addblock(msg.Block); err != nil {
			return err
		}
		ob.requestTimer.Remove(msg.Block.Header.Height)
	default:
		panic(ErrUnknownMessage) //TEMP
		return ErrUnknownMessage
	}
	return nil
}

func (ob *Observer) addblock(b *types.Block) error {
	cp := ob.cs.cn.Provider()
	if b.Header.Height <= cp.Height() {
		h, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if h != encoding.Hash(b.Header) {
			//TODO : critical error signal
			panic(chain.ErrFoundForkedBlock)
		}
	} else {
		if item := ob.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				panic(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}

func (ob *Observer) adjustFormulatorMap() map[common.Address]bool {
	FormulatorMap := ob.fs.FormulatorMap()
	FormulatorMap[common.NewAddress(0, 2, 0)] = true // TEMP
	now := time.Now().UnixNano()
	for addr := range FormulatorMap {
		if now < ob.ignoreMap[addr] {
			delete(FormulatorMap, addr)
		}
	}
	return FormulatorMap
}
