package pof

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/encoding"
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
		round:        NewVoteRound(cs.cn.Provider().Height()+1, cs.maxBlocksPerFormulator),
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: MyPublicHash,
		messageQueue: queue.NewQueue(),
	}
}

func (ob *Observer) Test() {
	obround := NewVoteRound(1, 10)
	queueTimer := time.NewTimer(time.Millisecond)
	voteTimer := time.NewTimer(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-queueTimer.C:
				v := ob.messageQueue.Pop()
				i := 0
				for v != nil {
					i++
					item := v.(*messageItem)
					if msg, is := item.Message.(*BlockGenMessage); is {
						ob.Lock()
						//ob.handleBlockGenMessage(msg, item.Raw)
						ob.Unlock()
					} else {
						ob.Lock()
						//ob.handleObserverMessage(item.PublicHash, item.Message, qm)
						ob.Unlock()
						log.Println(msg)
					}
					v = ob.messageQueue.Pop()
				}
				queueTimer.Reset(10 * time.Millisecond)
			case <-voteTimer.C:
				ob.Lock()
				cp := ob.cs.cn.Provider()
				Height := cp.Height()
				if ob.round.BlockRound.TargetHeight < Height {
					Diff := ob.round.BlockRound.TargetHeight - Height
					if ob.round.RemainBlocks > Diff {
						ob.round.RemainBlocks -= Diff
						ob.round.BlockRound = NewBlockRound(Height)
					} else {
						ob.round = NewVoteRound(Height+1, ob.cs.maxBlocksPerFormulator)
					}
				}
				//if len(ob.adjustFormulatorMap()) > 0 {
				if true {
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
							ob.round = NewVoteRound(Height+1, ob.cs.maxBlocksPerFormulator)
							ob.roundFirstTime = 0
							ob.roundFirstHeight = 0
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
	}()

	var (
		ErrInvalidVote       = errors.New("invalid vote")
		ErrInvalidRoundState = errors.New("invalid round state")
		ErrAlreadyVoted      = errors.New("already voted")
	)

	func() error {
		cp := ob.cs.cn.Provider()
		var SenderPublicHash common.PublicHash

		msgCh := make(chan interface{})
		for m := range msgCh {
			msgh := encoding.Hash(m)

			switch msg := m.(type) {
			case *RoundVoteMessage:
				pubkey, err := common.RecoverPubkey(msgh, msg.Signautre)
				if err != nil {
					return err
				} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
					return common.ErrInvalidPublicHash
				} else if !ob.cs.observerKeyMap.Has(obkey) {
					return ErrInvalidObserverKey
				}

				if msg.RoundVote.VoteTargetHeight != cp.Height() {
					if SenderPublicHash != ob.myPublicHash {
						if msg.RoundVote.VoteTargetHeight < cp.Height() {
							//ob.sendStateTo(SenderPublicHash)
						} else {
							ob.round = NewVoteRound(cp.Height()+1, ob.cs.maxBlocksPerFormulator)
							ob.roundFirstTime = 0
							ob.roundFirstHeight = 0
							//ob.sendRoundVote()
						}
					}
					return ErrInvalidVote
				}

				//[same round]
				if obround.RoundState != RoundVoteState {
					if !msg.RoundVote.IsReply {
						//ob.sendRoundVoteTo(SenderPublicHash)
					}
					return ErrInvalidRoundState
				}

				//[same state]
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

				//[same chain]
				if old, has := ob.round.RoundVoteMessageMap[SenderPublicHash]; has {
					if msg.RoundVote.Timestamp <= old.RoundVote.Timestamp {
						if !msg.RoundVote.IsReply {
							//ob.sendRoundVoteTo(SenderPublicHash)
						}
						return ErrAlreadyVoted
					}
				}
				ob.round.RoundVoteMessageMap[SenderPublicHash] = msg

				if !msg.RoundVote.IsReply {
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
				pubkey, err := common.RecoverPubkey(msgh, msg.Signautre)
				if err != nil {
					return err
				} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
					return common.ErrInvalidPublicHash
				} else if !ob.cs.observerKeyMap.Has(obkey) {
					return ErrInvalidObserverKey
				}

				if msg.RoundVoteAck.VoteTargetHeight != cp.Height() {
					if SenderPublicHash != ob.myPublicHash {
						if msg.RoundVoteAck.VoteTargetHeight < cp.Height() {
							//ob.sendStateTo(SenderPublicHash)
						} else {
							ob.round = NewVoteRound(cp.Height()+1, ob.cs.maxBlocksPerFormulator)
							ob.roundFirstTime = 0
							ob.roundFirstHeight = 0
							//ob.sendRoundVote()
						}
					}
					return ErrInvalidVote
				}

				//[same round]
				if ob.round.RoundState != RoundVoteAckState {
					if ob.round.RoundState < RoundVoteAckState {
						ob.round.RoundVoteAckMessageWaitMap[SenderPublicHash] = msg
					} else {
						if !msg.RoundVoteAck.IsReply {
							//ob.sendRoundVoteAckTo(SenderPublicHash)
						}
					}
					return ErrInvalidRoundState
				}

				//[same state]
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

				//[same chain]
				if old, has := ob.round.RoundVoteAckMessageMap[SenderPublicHash]; has {
					if msg.RoundVoteAck.Timestamp <= old.RoundVoteAck.Timestamp {
						if !msg.RoundVoteAck.IsReply {
							//ob.sendRoundVoteTo(SenderPublicHash)
						}
						return ErrAlreadyVoted
					}
				}
				ob.round.RoundVoteAckMessageMap[SenderPublicHash] = msg

				if !msg.RoundVoteAck.IsReply {
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
								PrevHash:             cp.LastHash(),
								TargetHeight:         ob.round.VoteTargetHeight,
								TimeoutCount:         ob.round.MinRoundVoteAck.TimeoutCount,
								Formulator:           ob.round.MinRoundVoteAck.Formulator,
								FormulatorPublicHash: ob.round.MinRoundVoteAck.FormulatorPublicHash,
							}
							log.Println(nm)
							//ob.fs.SendTo(ob.round.MinRoundVoteAck.Formulator, nm)
						}
						//ob.sendRoundSetup()
						if ob.round.BlockRound.BlockGenMessageWait != nil && ob.round.BlockRound.BlockGenMessage == nil {
							ob.messageQueue.Push(&messageItem{
								Message: ob.round.BlockRound.BlockGenMessageWait,
							})
						}
					}
				}
			case *RoundSetupMessage:
				pubkey, err := common.RecoverPubkey(msgh, msg.Signautre)
				if err != nil {
					return err
				} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
					return common.ErrInvalidPublicHash
				} else if !ob.cs.observerKeyMap.Has(obkey) {
					return ErrInvalidObserverKey
				}

				RoundVoteAck := msg.RoundSetup.RoundVoteAcks[0]
				if RoundVoteAck.VoteTargetHeight != cp.Height() {
					if SenderPublicHash != ob.myPublicHash {
						if RoundVoteAck.VoteTargetHeight < cp.Height() {
							//ob.sendStateTo(SenderPublicHash)
						} else {
							ob.round = NewVoteRound(cp.Height()+1, ob.cs.maxBlocksPerFormulator)
							ob.roundFirstTime = 0
							ob.roundFirstHeight = 0
							//ob.sendRoundVote()
						}
					}
					return ErrInvalidVote
				}

				//[same round]
				if ob.round.RoundState != RoundVoteAckState {
					if ob.round.RoundState < BlockWaitState {
						//[if valid]
						//change to BlockVoteState
						//ob.sendRoundSetup()
					} else if ob.round.RoundState == BlockVoteState {
						if !msg.RoundSetup.IsReply {
							//ob.sendBlockGenTo(SenderPublicHash)
						}
					}
					return ErrInvalidRoundState
				}

				//[same state]
				Top, err := ob.cs.rt.TopRank(int(RoundVoteAck.TimeoutCount))
				if err != nil {
					return err
				}
				if RoundVoteAck.Formulator != Top.Address {
					return ErrInvalidVote
				}
				if RoundVoteAck.FormulatorPublicHash != Top.PublicHash {
					return ErrInvalidVote
				}

				//[same chain]
				if old, has := ob.round.RoundVoteAckMessageMap[SenderPublicHash]; has {
					if msg.RoundVoteAck.Timestamp <= old.RoundVoteAck.Timestamp {
						if !msg.RoundVoteAck.IsReply {
							//ob.sendRoundVoteTo(SenderPublicHash)
						}
						return ErrAlreadyVoted
					}
				}
				ob.round.RoundVoteAckMessageMap[SenderPublicHash] = msg

				if !msg.RoundVoteAck.IsReply {
					//ob.sendRoundVoteAckTo(SenderPublicHash)
				}

				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < BlockWaitState {
					//[if valid]
					//change to BlockVoteState
					//send round setup
				} else if obround.RoundState == BlockVoteState {
					//reply block gen if recv not reply -> x
				}
			case *BlockGenMessage:
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < BlockWaitState {
					//save to wait
				} else if obround.RoundState == BlockVoteState {
					//reply block vote if recv not reply -> x
				}
				//[same state]
				//broadcast block gen if min.PubHash == me
				//[if valid block]
				//send block vote
				//change to BlockVoteState
			case *BlockVoteMessage:
				pubkey, err := common.RecoverPubkey(msgh, msg.Signautre)
				if err != nil {
					return err
				} else if obkey := common.NewPublicHash(pubkey); SenderPublicHash != obkey {
					return common.ErrInvalidPublicHash
				} else if !ob.cs.observerKeyMap.Has(obkey) {
					return ErrInvalidObserverKey
				}

				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < BlockVoteState {
					//save to wait
				}
				//[same state]
				//update block vote map
				//[if get more than 3]
				//connect block
				//send block onsign to formulator
				//decrease RemainBlocks
				//[if RemainBlocks > 0]
				//change to BlockWaitState
				//[if RemainBlocks <= 0]
				//clear round
			}
		}
		return nil
	}()
	return

}
