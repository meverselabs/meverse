package node

import (
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/common/queue"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
)

// BlockTime defines the block generation interval
const BlockTime = 500 * time.Millisecond

type messageItem struct {
	PublicKey common.PublicKey
	Message   interface{}
	Packet    []byte
}

// ObserverNode observes a block by the consensus
type ObserverNode struct {
	obID string
	sync.Mutex
	ChainID          *big.Int
	key              key.Key
	ms               *ObserverNodeMesh
	fs               *GeneratorService
	cn               *chain.Chain
	ct               chain.Committer
	observerKeyMap   map[common.PublicKey]bool
	round            *VoteRound
	roundFirstTime   uint64
	roundFirstHeight uint32
	ignoreMap        map[common.Address]int64
	myPublicKey      common.PublicKey
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
func NewObserverNode(ChainID *big.Int, key key.Key, NetAddressMap map[common.PublicKey]string, cn *chain.Chain, obID string) *ObserverNode {
	observerKeyMap := map[common.PublicKey]bool{}
	for k := range NetAddressMap {
		observerKeyMap[k] = true
	}
	ob := &ObserverNode{
		obID:           obID,
		ChainID:        ChainID,
		key:            key,
		cn:             cn,
		ct:             chain.NewChainCommiter(cn),
		observerKeyMap: observerKeyMap,
		round:          NewVoteRound(cn.Provider().Height()+1, prefix.MaxBlocksPerGenerator),
		ignoreMap:      map[common.Address]int64{},
		myPublicKey:    key.PublicKey(),
		statusMap:      map[string]*p2p.Status{},
		blockQ:         queue.NewSortedQueue(),
		messageQueue:   queue.NewQueue(),
		recvChan:       make(chan *p2p.RecvMessageItem, 1000),
		sendChan:       make(chan *p2p.SendMessageItem, 1000),
		singleCache:    gcache.New(500).LRU().Build(),
		batchCache:     gcache.New(500).LRU().Build(),
	}
	ob.ms = NewObserverNodeMesh(key, NetAddressMap, ob)
	ob.fs = NewGeneratorService(ob)
	ob.requestTimer = p2p.NewRequestTimer(ob)
	return ob
}

// Init initializes observer
func (ob *ObserverNode) Init() error {
	return nil
}

// Close terminates the observer
func (ob *ObserverNode) Close() {
	ob.closeLock.Lock()
	defer ob.closeLock.Unlock()

	ob.Lock()
	defer ob.Unlock()

	ob.isClose = true
	ob.cn.Close()
}

func (ob *ObserverNode) ResetRound() { //TEMP
	ob.Lock()
	defer ob.Unlock()

	ob.resetVoteRound(true)
}

// Run starts the pof consensus on the observer
func (ob *ObserverNode) Run(BindObserver string, BindGenerator string) {
	ob.Lock()
	if ob.isRunning {
		ob.Unlock()
		return
	}
	ob.isRunning = true
	ob.Unlock()

	go ob.ms.Run(BindObserver)
	go ob.fs.Run(BindGenerator)
	go ob.requestTimer.Run()

	for i := 0; i < 2; i++ {
		go func() {
			for item := range ob.recvChan {
				if ob.isClose {
					break
				}
				m, err := p2p.PacketToMessage(item.Packet)
				if err != nil {
					log.Printf("PacketToMessage %+v\n", err)
					ob.fs.RemovePeer(item.PeerID)
					continue
				}
				if p, has := ob.fs.Peer(item.PeerID); has {
					if err := ob.handleGeneratorMessage(p, m, item.Packet); err != nil {
						log.Printf("Generator Error %v  %+v\n", p.Name(), err)
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
			cp := ob.cn.Provider()
			ob.Lock()
			hasItem := false
			TargetHeight := uint64(cp.Height() + 1)
			Count := 0
			item := ob.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := ob.cn.ConnectBlock(b, nil); err != nil {
					fmt.Printf("%+v\n", err)
					panic(err)
					break
				}
				if DEBUG {
					log.Println(ob.obID, "observer", cp.Height(), "BlockConnectedQ", b.Header.Generator.String(), ob.round.RoundState, b.Header.Height, (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond), len(b.Body.Transactions))
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
				if err := ob.handleObserverMessage(item.PublicKey, item.Message, item.Packet); err != nil {
					if DEBUG {
						switch errors.Cause(err) {
						case ErrInvalidRoundState, ErrAlreadyVoted:
						default:
							fmt.Printf("observer %+v\n", err)
						}
					}
				}
				ob.Unlock()
				v = ob.messageQueue.Pop()
			}
			if ob.fs.PeerCount() == 0 || len(ob.adjustGeneratorMap()) == 0 {
				queueTimer.Reset(100 * time.Millisecond)
			} else {
				queueTimer.Reset(50 * time.Millisecond)
			}
		case <-voteTimer.C:
			ob.Lock()
			cp := ob.cn.Provider()
			ob.syncVoteRound()
			IsFailable := true
			if len(ob.adjustGeneratorMap()) > 0 {
				if ob.round.MinRoundVoteAck != nil {
					if DEBUG {
						log.Println(ob.obID, "observer", cp.Height(), "Current State", ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), ob.round.MinRoundVoteAck.Generator.String(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				} else {
					if DEBUG {
						log.Println(ob.obID, "observer", cp.Height(), "Current State", ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
					}
				}
				if ob.round.RoundState == RoundVoteState {
					ob.sendRoundVote()
					ob.broadcastStatus()
				} else if ob.round.RoundState == BlockVoteState {
					br, has := ob.round.BlockRoundMap[ob.round.TargetHeight]
					if has {
						ob.sendBlockVote(br.BlockGenMessage)
						if DEBUG {
							log.Println(ob.obID, "observer", cp.Height(), "sendBlockVote", ob.round.MinRoundVoteAck.Generator.String(), bin.MustWriterToHash(&br.BlockGenMessage.Block.Header), ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
						}
						IsFailable = false
					}
				}
				if IsFailable {
					ob.round.VoteFailCount++
					if ob.round.VoteFailCount > 20 {
						if ob.round.MinRoundVoteAck != nil {
							addr := ob.round.MinRoundVoteAck.Generator
							if _, has := ob.ignoreMap[addr]; has {
								ob.fs.RemovePeer(string(addr[:]))
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(10*time.Second)
							} else {
								ob.ignoreMap[addr] = time.Now().UnixNano() + int64(10*time.Second)
							}
							if DEBUG {
								log.Println(ob.obID, "observer", cp.Height(), "Failure", ob.round.MinRoundVoteAck.Generator.String(), ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
							}
						} else {
							if DEBUG {
								log.Println(ob.obID, "observer", cp.Height(), "Failure", ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
							}
						}
						ob.resetVoteRound(true)
					}
				}
			} else {
				if DEBUG {
					log.Println(ob.obID, "observer", cp.Height(), "No Generator", ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
				}
			}
			ob.Unlock()

			voteTimer.Reset(100 * time.Millisecond)
		}
	}
}

// OnTimerExpired called when rquest expired
func (ob *ObserverNode) OnTimerExpired(height uint32, value string) {
	if height > ob.cn.Provider().Height() {
		var TargetPublicKey common.PublicKey
		copy(TargetPublicKey[:], []byte(value))
		list := ob.ms.Peers()
		for _, p := range list {
			var PubKey common.PublicKey
			copy(PubKey[:], []byte(p.ID()))
			if PubKey != ob.myPublicKey && PubKey != TargetPublicKey {
				ob.sendRequestBlockTo(PubKey, height, 1)
				break
			}
		}
	}
}

func (ob *ObserverNode) addBlock(b *types.Block) error {
	cp := ob.cn.Provider()
	if b.Header.Height <= cp.Height() {
		h, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if h != bin.MustWriterToHash(&b.Header) {
			//TODO : critical error signal
			return chain.ErrFoundForkedBlock
		}
	} else {
		if item := ob.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if bin.MustWriterToHash(&old.Header) != bin.MustWriterToHash(&b.Header) {
				//TODO : critical error signal
				return errors.WithStack(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}

func (ob *ObserverNode) adjustGeneratorMap() map[common.Address]bool {
	GeneratorMap := ob.fs.GeneratorMap()
	now := time.Now().UnixNano()
	for addr := range GeneratorMap {
		if now < ob.ignoreMap[addr] {
			delete(GeneratorMap, addr)
		}
	}
	return GeneratorMap
}

func (ob *ObserverNode) syncVoteRound() {
	TargetHeight := ob.cn.Provider().Height() + 1
	if ob.round.TargetHeight < TargetHeight {
		IsContinue := false
		Top, err := ob.cn.TopGenerator(0)
		if err != nil {
			panic(err)
		}
		if ob.round.MinRoundVoteAck != nil && Top == ob.round.MinRoundVoteAck.Generator {
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
			if DEBUG {
				log.Println(ob.obID, "observer", ob.cn.Provider().Height(), "Turn Over", ob.round.RoundState, len(ob.adjustGeneratorMap()), ob.fs.PeerCount(), (time.Now().UnixNano()-ob.prevRoundEndTime)/int64(time.Millisecond))
			}
			ob.resetVoteRound(false)
		}
	}
}

func (ob *ObserverNode) resetVoteRound(resetStat bool) {
	ob.round = NewVoteRound(ob.cn.Provider().Height()+1, prefix.MaxBlocksPerGenerator)
	ob.prevRoundEndTime = time.Now().UnixNano()
	if resetStat {
		ob.roundFirstTime = 0
		ob.roundFirstHeight = 0
	}
}
