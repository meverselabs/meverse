package pof

import (
	"log"
	"sync"
	"time"

	"github.com/bluele/gcache"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/debug"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/apiserver"
	"github.com/fletaio/fleta/service/p2p"
)

type messageItem struct {
	PublicHash common.PublicHash
	Message    interface{}
	Packet     []byte
}

// ObserverNode observes a block by the consensus
type ObserverNode struct {
	sync.Mutex
	key              key.Key
	ms               *ObserverNodeMesh
	fs               *FormulatorService
	cs               *Consensus
	round            *VoteRound
	roundFirstTime   uint64
	roundFirstHeight uint32
	ignoreMap        map[common.Address]int64
	myPublicHash     common.PublicHash
	statusLock       sync.Mutex
	statusMap        map[string]*p2p.Status
	requestTimer     *p2p.RequestTimer
	blockQ           *queue.SortedQueue
	messageQueue     *queue.Queue
	recvChan         chan *p2p.RecvMessageItem
	sendChan         chan *p2p.SendMessageItem
	singleCache      gcache.Cache
	batchCache       gcache.Cache
	isRunning        bool
	closeLock        sync.RWMutex
	isClose          bool

	prevRoundEndTime int64 // FOR DEBUG
}

// NewObserverNode returns a ObserverNode
func NewObserverNode(key key.Key, NetAddressMap map[common.PublicHash]string, cs *Consensus) *ObserverNode {
	ob := &ObserverNode{
		key:          key,
		cs:           cs,
		round:        NewVoteRound(cs.cn.Provider().Height()+1, cs.maxBlocksPerFormulator),
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		statusMap:    map[string]*p2p.Status{},
		blockQ:       queue.NewSortedQueue(),
		messageQueue: queue.NewQueue(),
		recvChan:     make(chan *p2p.RecvMessageItem, 1000),
		sendChan:     make(chan *p2p.SendMessageItem, 1000),
		singleCache:  gcache.New(500).LRU().Build(),
		batchCache:   gcache.New(500).LRU().Build(),
	}
	ob.ms = NewObserverNodeMesh(key, NetAddressMap, ob)
	ob.fs = NewFormulatorService(ob)
	ob.requestTimer = p2p.NewRequestTimer(ob)

	rlog.SetRLogAddress("ob:" + ob.myPublicHash.String())
	return ob
}

// Init initializes observer
func (ob *ObserverNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("pof.RoundVoteMessage"), &RoundVoteMessage{})
	fc.Register(types.DefineHashedType("pof.RoundVoteAckMessage"), &RoundVoteAckMessage{})
	fc.Register(types.DefineHashedType("pof.NextRoundVoteMessage"), &NextRoundVoteMessage{})
	fc.Register(types.DefineHashedType("pof.NextRoundVoteAckMessage"), &NextRoundVoteAckMessage{})
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(BlockGenMessageType, &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockVoteMessage"), &BlockVoteMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenRequestMessage"), &BlockGenRequestMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})

	if s, err := ob.cs.cn.ServiceByName("fleta.apiserver"); err != nil {
	} else if as, is := s.(*apiserver.APIServer); !is {
	} else {
		js, err := as.JRPC("observer")
		if err != nil {
			return err
		}
		js.Set("formulatorMap", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			m := ob.fs.FormulatorMap()
			nm := map[string]bool{}
			for k, v := range m {
				nm[k.String()] = v
			}
			return nm, nil
		})
		js.Set("adjustFormulatorMap", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			m := ob.adjustFormulatorMap()
			nm := map[string]bool{}
			for k, v := range m {
				nm[k.String()] = v
			}
			return nm, nil
		})
	}
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

	for i := 0; i < 2; i++ {
		go func() {
			for item := range ob.recvChan {
				if ob.isClose {
					break
				}
				m, err := p2p.PacketToMessage(item.Packet)
				if err != nil {
					log.Println("PacketToMessage", err)
					ob.fs.RemovePeer(item.PeerID)
					continue
				}
				if p, has := ob.fs.Peer(item.PeerID); has {
					if err := ob.handleFormulatorMessage(p, m, item.Packet); err != nil {
						log.Println("Formulator Error", p.Name(), err)
						ob.fs.RemovePeer(item.PeerID)
						continue
					}
				}
			}
		}()
	}

	for i := 0; i < 2; i++ {
		go func() {
			for item := range ob.sendChan {
				if ob.isClose {
					break
				}
				if len(item.Packet) > 0 {
					if err := ob.fs.SendTo(item.Address, item.Packet); err != nil {
						ob.fs.RemovePeer(string(item.Address[:]))
					}
				} else {
					if err := ob.fs.SendTo(item.Address, item.Packet); err != nil {
						ob.fs.RemovePeer(string(item.Address[:]))
					}
				}
			}
		}()
	}

	blockTimer := time.NewTimer(time.Millisecond)
	queueTimer := time.NewTimer(time.Millisecond)
	voteTimer := time.NewTimer(time.Millisecond)
	for !ob.isClose {
		select {
		case <-blockTimer.C:
			cp := ob.cs.cn.Provider()
			ob.Lock()
			hasItem := false
			TargetHeight := uint64(cp.Height() + 1)
			Count := 0
			item := ob.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := ob.cs.cn.ConnectBlock(b, nil); err != nil {
					rlog.Println(err)
					panic(err)
					break
				}
				if debug.DEBUG {
					rlog.Println(cp.Height(), "BlockConnectedQ", b.Header.Generator.String(), ob.round.RoundState, b.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				}
				TargetHeight++
				Count++
				if Count > 100 {
					break
				}
				item = ob.blockQ.PopUntil(TargetHeight)
				hasItem = true
			}
			ob.Unlock()

			if hasItem {
				ob.broadcastStatus()
				blockTimer.Reset(50 * time.Millisecond)
			} else {
				blockTimer.Reset(200 * time.Millisecond)
			}
		case <-queueTimer.C:
			v := ob.messageQueue.Pop()
			i := 0
			for v != nil {
				i++
				item := v.(*messageItem)
				ob.Lock()
				ob.handleObserverMessage(item.PublicHash, item.Message, item.Packet)
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
				if ob.round.MinRoundVoteAck != nil {
					if debug.DEBUG {
						rlog.Println(cp.Height(), "Current State", ob.round.MinRoundVoteAck.Formulator.String(), ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				} else {
					if debug.DEBUG {
						rlog.Println(cp.Height(), "Current State", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				}
				if ob.round.RoundState == RoundVoteState {
					ob.sendRoundVote()
					ob.broadcastStatus()
				} else if ob.round.RoundState == BlockVoteState {
					br, has := ob.round.BlockRoundMap[ob.round.TargetHeight]
					if has {
						ob.sendBlockVote(br.BlockGenMessage)
						if debug.DEBUG {
							rlog.Println(cp.Height(), "sendBlockVote", ob.round.MinRoundVoteAck.Formulator.String(), encoding.Hash(br.BlockGenMessage.Block.Header), ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
						}
						IsFailable = false
					}
				}
				if IsFailable {
					ob.round.VoteFailCount++
					if ob.round.VoteFailCount > 20 {
						if ob.round.MinRoundVoteAck != nil {
							addr := ob.round.MinRoundVoteAck.Formulator
							if _, has := ob.ignoreMap[addr]; has {
								ob.fs.RemovePeer(string(addr[:]))
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(120*time.Second)
							} else {
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(30*time.Second)
							}
							if debug.DEBUG {
								rlog.Println(cp.Height(), "Failure", ob.round.MinRoundVoteAck.Formulator.String(), ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
							}
						} else {
							if debug.DEBUG {
								rlog.Println(cp.Height(), "Failure", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
							}
						}
						ob.resetVoteRound(true)
					}
				}
			} else {
				if debug.DEBUG {
					rlog.Println(cp.Height(), "No Formulator", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				}
			}
			ob.Unlock()

			voteTimer.Reset(100 * time.Millisecond)
		}
	}
}

// OnTimerExpired called when rquest expired
func (ob *ObserverNode) OnTimerExpired(height uint32, value string) {
	if height > ob.cs.cn.Provider().Height() {
		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(value))
		list := ob.ms.Peers()
		for _, p := range list {
			var pubhash common.PublicHash
			copy(pubhash[:], []byte(p.ID()))
			if pubhash != ob.myPublicHash && pubhash != TargetPublicHash {
				ob.sendRequestBlockTo(pubhash, height, 1)
				break
			}
		}
	}
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
			return chain.ErrFoundForkedBlock
		}
	} else {
		if item := ob.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
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

func (ob *ObserverNode) syncVoteRound() {
	TargetHeight := ob.cs.cn.Provider().Height() + 1
	if ob.round.TargetHeight < TargetHeight {
		IsContinue := false
		Top, err := ob.cs.rt.TopRank(0)
		if err != nil {
			panic(err)
		}
		if ob.round.MinRoundVoteAck != nil && Top.Address == ob.round.MinRoundVoteAck.Formulator {
			if br, has := ob.round.BlockRoundMap[TargetHeight]; has {
				ob.round.TargetHeight = TargetHeight
				ob.round.RoundState = BlockWaitState
				if br.BlockGenMessageWait != nil && br.BlockGenMessage == nil {
					ob.messageQueue.Push(&messageItem{
						Message: br.BlockGenMessageWait,
					})
					br.BlockGenMessageWait = nil
				}
				IsContinue = true
			}
		}
		if !IsContinue {
			if debug.DEBUG {
				rlog.Println(ob.cs.cn.Provider().Height(), "Turn Over", ob.round.RoundState, len(ob.adjustFormulatorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}
			ob.resetVoteRound(false)
		}
	}
}

func (ob *ObserverNode) resetVoteRound(resetStat bool) {
	ob.round = NewVoteRound(ob.cs.cn.Provider().Height()+1, ob.cs.maxBlocksPerFormulator)
	ob.prevRoundEndTime = time.Now().UnixNano()
	if resetStat {
		ob.roundFirstTime = 0
		ob.roundFirstHeight = 0
	}
}
