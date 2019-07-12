package pof

import (
	"bytes"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/util"
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

// ObserverNode observes a block by the consensus
type ObserverNode struct {
	sync.Mutex
	key              key.Key
	ms               *ObserverNodeMesh
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

// NewObserverNode returns a ObserverNode
func NewObserverNode(key key.Key, NetAddressMap map[common.PublicHash]string, cs *Consensus) *ObserverNode {
	ob := &ObserverNode{
		key:          key,
		cs:           cs,
		round:        NewVoteRound(cs.cn.Provider().Height()+1, cs.maxBlocksPerFormulator-cs.blocksBySameFormulator),
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		messageQueue: queue.NewQueue(),
		blockQ:       queue.NewSortedQueue(),
	}
	ob.ms = NewObserverNodeMesh(key, NetAddressMap, ob)
	ob.fs = NewFormulatorService(ob)
	ob.requestTimer = p2p.NewRequestTimer(ob)
	return ob
}

// Init initializes observer
func (ob *ObserverNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("pof.RoundVoteMessage"), &RoundVoteMessage{})
	fc.Register(types.DefineHashedType("pof.RoundVoteAckMessage"), &RoundVoteAckMessage{})
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenMessage"), &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockVoteMessage"), &BlockVoteMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})
	return nil
}

// Close terminates the observer
func (ob *ObserverNode) Close() {
	ob.closeLock.Lock()
	defer ob.closeLock.Unlock()

	ob.Lock()
	defer ob.Unlock()

	ob.isClose = true
	ob.cs.cn.Close()
	ob.runEnd <- struct{}{}
}

// Run starts the pof consensus on the observer
func (ob *ObserverNode) Run(BindObserver string, BindFormulator string) {
	ob.Lock()
	if ob.isRunning {
		ob.Unlock()
		return
	}
	ob.isRunning = true
	ob.Unlock()

	go ob.ms.Run(BindObserver)
	go ob.fs.Run(BindFormulator)
	go ob.requestTimer.Run()

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
					log.Println(err)
					panic(err)
					break
				}
				log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "BlockConnected", b.Header.Generator.String(), ob.round.RoundState, b.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
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
				if err := ob.handleObserverMessage(item.PublicHash, item.Message, item.Raw); err != nil {
					if m, is := item.Message.(*BlockGenMessage); is {
						log.Println(err, m)
					}
				}
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
				log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "Current State", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				if ob.round.RoundState == RoundVoteState {
					ob.sendRoundVote()
					ob.broadcastStatus()
				} else if ob.round.RoundState == BlockVoteState {
					br, has := ob.round.BlockRoundMap[ob.round.TargetHeight]
					if has {
						ob.sendBlockVote(br.BlockGenMessage)
						IsFailable = false
					}
				}
				if IsFailable {
					ob.round.VoteFailCount++
					if ob.round.VoteFailCount > 20 {
						if ob.round.MinRoundVoteAck != nil {
							addr := ob.round.MinRoundVoteAck.Formulator
							_, has := ob.ignoreMap[addr]
							if has {
								ob.fs.RemovePeer(string(addr[:]))
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(120*time.Second)
							} else {
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(30*time.Second)
							}
						}
						log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "Failure", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
						ob.resetVoteRound(true)
						ob.sendRoundVote()
					}
				}
			} else {
				if cp.Height() > 0 {
					panic("HERE") //TEMP
				}
				log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "No Formulator", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}
			ob.Unlock()

			voteTimer.Reset(100 * time.Millisecond)
		case <-ob.runEnd:
			return
		}
	}
}

// OnTimerExpired called when rquest expired
func (ob *ObserverNode) OnTimerExpired(height uint32, value interface{}) {
	TargetPublicHash := value.(common.PublicHash)
	list := ob.ms.Peers()
	for _, p := range list {
		var pubhash common.PublicHash
		copy(pubhash[:], []byte(p.ID()))
		if pubhash != ob.myPublicHash && pubhash != TargetPublicHash {
			ob.sendRequestBlockTo(pubhash, height)
			break
		}
	}
}

func (ob *ObserverNode) syncVoteRound() {
	cp := ob.cs.cn.Provider()
	TargetHeight := cp.Height() + 1
	if ob.round.TargetHeight < TargetHeight {
		Diff := TargetHeight - ob.round.TargetHeight
		if ob.round.RemainBlocks > Diff {
			ob.round.TargetHeight = TargetHeight
			ob.round.RemainBlocks = ob.round.RemainBlocks - Diff
			if ob.round.MinRoundVoteAck != nil {
				ob.round.RoundState = BlockWaitState
				br, has := ob.round.BlockRoundMap[TargetHeight]
				if has {
					if br.BlockGenMessageWait != nil && br.BlockGenMessage == nil {
						ob.messageQueue.Push(&messageItem{
							Message: br.BlockGenMessageWait,
						})
						br.BlockGenMessageWait = nil
					}
				}
			}
		} else {
			log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "Turn Over", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			ob.resetVoteRound(false)
		}
	}
}

func (ob *ObserverNode) resetVoteRound(resetStat bool) {
	ob.round = NewVoteRound(ob.cs.cn.Provider().Height()+1, ob.cs.maxBlocksPerFormulator-ob.cs.blocksBySameFormulator)
	ob.prevRoundEndTime = time.Now().UnixNano()
	if resetStat {
		ob.roundFirstTime = 0
		ob.roundFirstHeight = 0
	}
}

func (ob *ObserverNode) onObserverRecv(p p2p.Peer, m interface{}) error {
	if msg, is := m.(*BlockGenMessage); is {
		ob.messageQueue.Push(&messageItem{
			Message: msg,
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

func (ob *ObserverNode) onFormulatorRecv(p p2p.Peer, m interface{}, raw []byte) error {
	cp := ob.cs.cn.Provider()

	switch msg := m.(type) {
	case *BlockGenMessage:
		ob.messageQueue.Push(&messageItem{
			Message: msg,
			Raw:     raw,
		})
	case *p2p.RequestMessage:
		ob.Lock()
		defer ob.Unlock()

		enable := false

		if p.GuessHeight() < msg.Height {
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
					rankMap[string(r.Address[:])] = true
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
	case *p2p.StatusMessage:
		ob.Lock()
		defer ob.Unlock()

		Height := cp.Height()
		if Height >= msg.Height {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				panic(chain.ErrFoundForkedBlock)
			}
		}
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (ob *ObserverNode) handleObserverMessage(SenderPublicHash common.PublicHash, m interface{}, raw []byte) error {
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
						ob.sendBlockGenTo(br.BlockGenMessage, SenderPublicHash)
						ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
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
		//log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "RoundVoteAckMessage", ob.round.RoundState, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
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
					log.Println("Observer", "BlockReqMessage", ob.round.MinRoundVoteAck.PublicHash.String(), ob.myPublicHash.String(), cp.Height())
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
		//log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "BlockGenMessage", ob.round.RoundState, msg.Block.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
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

		if ob.round.MinRoundVoteAck != nil {
			if ob.round.MinRoundVoteAck.PublicHash == ob.myPublicHash && len(raw) > 0 {
				var buffer bytes.Buffer
				buffer.Write(util.Uint16ToBytes(types.DefineHashedType("pof.BlockGenMessage")))
				buffer.Write(raw)
				log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "BroadcastRaw", msg.Block.Header.Height, ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				ob.ms.BroadcastRaw(buffer.Bytes())
			}
		}

		if msg.Block.Header.Height != ob.round.TargetHeight {
			if msg.Block.Header.Height > ob.round.TargetHeight {
				br.BlockGenMessageWait = msg
			}
			return ErrInvalidVote
		}

		//[check state]
		if ob.round.RoundState != BlockWaitState {
			if ob.round.RoundState < BlockWaitState {
				br.BlockGenMessageWait = msg
			}
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
		if msg.Block.Header.PrevHash != cp.LastHash() {
			return ErrInvalidVote
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
		if msg.Block.Header.ContextHash != ctx.Hash() {
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
	case *BlockVoteMessage:
		//log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "BlockVoteMessage", ob.round.RoundState, msg.BlockVote.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
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
		if ob.round.RoundState != BlockVoteState {
			if ob.round.RoundState < BlockVoteState {
				if _, has := ob.ignoreMap[msg.BlockVote.Header.Generator]; has {
					delete(ob.ignoreMap, msg.BlockVote.Header.Generator)
					ob.round.VoteFailCount = 0
				}
				br.BlockVoteMessageWaitMap[SenderPublicHash] = msg
			}
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
				ob.sendBlockVoteTo(br.BlockGenMessage, SenderPublicHash)
			}
			return ErrAlreadyVoted
		}
		br.BlockVoteMap[SenderPublicHash] = msg.BlockVote

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

			b := &types.Block{
				Header:               br.BlockGenMessage.Block.Header,
				TransactionTypes:     br.BlockGenMessage.Block.TransactionTypes,
				Transactions:         br.BlockGenMessage.Block.Transactions,
				TranactionSignatures: br.BlockGenMessage.Block.TranactionSignatures,
				Signatures:           append([]common.Signature{br.BlockGenMessage.GeneratorSignature}, sigs...),
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

			log.Println("Observer", ob.myPublicHash.String(), cp.Height(), "BlockConnected", b.Header.Generator.String(), ob.round.RoundState, msg.BlockVote.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))

			ob.round.RemainBlocks--
			if ob.round.RemainBlocks > 0 {
				ob.round.RoundState = BlockWaitState
				ob.round.VoteFailCount = 0
				delete(ob.round.BlockRoundMap, ob.round.TargetHeight)
				ob.round.TargetHeight++
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
				log.Println(SenderPublicHash.String(), h.String(), msg.LastHash.String(), msg.Height)
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *p2p.BlockMessage:
		if err := ob.addBlock(msg.Block); err != nil {
			return err
		}
		ob.requestTimer.Remove(msg.Block.Header.Height)
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (ob *ObserverNode) addBlock(b *types.Block) error {
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

func (ob *ObserverNode) adjustFormulatorMap() map[common.Address]bool {
	FormulatorMap := ob.fs.FormulatorMap()
	now := time.Now().UnixNano()
	for addr := range FormulatorMap {
		if now < ob.ignoreMap[addr] {
			delete(FormulatorMap, addr)
		}
	}
	return FormulatorMap
}
