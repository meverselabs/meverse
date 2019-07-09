package p2p

import (
	"log"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Node receives a block by the consensus
type Node struct {
	sync.Mutex
	key          key.Key
	ms           *NodeMesh
	cn           *chain.Chain
	myPublicHash common.PublicHash
	requestTimer *RequestTimer
	blockQ       *queue.SortedQueue
	isRunning    bool
	closeLock    sync.RWMutex
	runEnd       chan struct{}
	isClose      bool
}

// NewNode returns a Node
func NewNode(key key.Key, SeedNodeMap map[common.PublicHash]string, cn *chain.Chain) *Node {
	nd := &Node{
		key:          key,
		cn:           cn,
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		blockQ:       queue.NewSortedQueue(),
	}
	nd.ms = NewNodeMesh(key, SeedNodeMap, nd)
	nd.requestTimer = NewRequestTimer(nd)
	return nd
}

// Init initializes node
func (nd *Node) Init() error {
	fc := encoding.Factory("pof.message")
	fc.Register(types.DefineHashedType("PingMessage"), &PingMessage{})
	fc.Register(types.DefineHashedType("StatusMessage"), &StatusMessage{})
	fc.Register(types.DefineHashedType("BlockMessage"), &BlockMessage{})
	fc.Register(types.DefineHashedType("RequestMessage"), &RequestMessage{})
	return nil
}

// Close terminates the node
func (nd *Node) Close() {
	nd.closeLock.Lock()
	defer nd.closeLock.Unlock()

	nd.Lock()
	defer nd.Unlock()

	nd.isClose = true
	nd.cn.Close()
	nd.runEnd <- struct{}{}
}

// Run starts the node
func (nd *Node) Run(BindAddress string) {
	nd.Lock()
	if nd.isRunning {
		nd.Unlock()
		return
	}
	nd.isRunning = true
	nd.Unlock()

	go nd.ms.Run(BindAddress)
	go nd.requestTimer.Run()

	blockTimer := time.NewTimer(time.Millisecond)
	for !nd.isClose {
		select {
		case <-blockTimer.C:
			cp := nd.cn.Provider()
			nd.Lock()
			TargetHeight := uint64(cp.Height() + 1)
			item := nd.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := nd.cn.ConnectBlock(b); err != nil {
					log.Println(err)
					panic(err)
					break
				}
				log.Println("Node", nd.myPublicHash.String(), cp.Height(), "BlockConnected", b.Header.Height)
				nd.broadcastStatus()
				TargetHeight++
				item = nd.blockQ.PopUntil(TargetHeight)
			}
			nd.Unlock()
			blockTimer.Reset(50 * time.Millisecond)
		case <-nd.runEnd:
			return
		}
	}
}

// OnTimerExpired called when rquest expired
func (nd *Node) OnTimerExpired(height uint32, value interface{}) {
	TargetPublicHash := value.(common.PublicHash)
	list := nd.ms.Peers()
	for _, p := range list {
		var pubhash common.PublicHash
		copy(pubhash[:], []byte(p.ID()))
		if pubhash != nd.myPublicHash && pubhash != TargetPublicHash {
			nd.sendRequestBlockTo(pubhash, height)
			break
		}
	}
}

// OnRecv called when message received
func (nd *Node) OnRecv(p Peer, m interface{}) error {
	cp := nd.cn.Provider()

	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))

	switch msg := m.(type) {
	case *RequestMessage:
		b, err := cp.Block(msg.Height)
		if err != nil {
			return err
		}
		sm := &BlockMessage{
			Block: b,
		}
		if err := nd.ms.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *StatusMessage:
		Height := cp.Height()
		if Height < msg.Height {
			for i := Height + 1; i <= Height+100 && i <= msg.Height; i++ {
				if !nd.requestTimer.Exist(i) {
					nd.sendRequestBlockTo(SenderPublicHash, i)
				}
			}
		} else {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				log.Println(p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *BlockMessage:
		if err := nd.addBlock(msg.Block); err != nil {
			return err
		}
		nd.requestTimer.Remove(msg.Block.Header.Height)
	default:
		panic(ErrUnknownMessage) //TEMP
		return ErrUnknownMessage
	}
	return nil
}

func (nd *Node) addBlock(b *types.Block) error {
	cp := nd.cn.Provider()
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
		if item := nd.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				panic(chain.ErrFoundForkedBlock)
			}
		}
	}
	return nil
}
