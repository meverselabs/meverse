package nodepoolmanage

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/service/p2p/peer"
	"github.com/fletaio/fleta/service/p2p/peermessage"
	"github.com/fletaio/fleta/service/p2p/storage"
)

// peer errors
var (
	ErrNotFoundPeer = errors.New("not found peer")
)

//Manager manages peer-connected networks.
type Manager interface {
	NewNode(addr string, hash string, ping time.Duration) error
	AddPeerList(ips []string, hashs []string)
	GetPeerList() (ips []string, hashs []string)
}

type nodeMesh interface {
	RequestConnect(Address string, TargetPubHash common.PublicHash)
	Peers() []peer.Peer
	RemovePeer(ID string)
	GetPeer(ID string) peer.Peer
	RequestPeerList(targetHash string)
	SendPeerList(targetHash string)
}

type nodePoolManage struct {
	nodes              *nodeStore
	nodeRotateIndex    int
	requestRotateIndex int
	peerStorage        storage.PeerStorage
	nodeMesh           nodeMesh
	BanPeerInfos       *ByTime
	myPublicHash       common.PublicHash

	putPeerListLock sync.Mutex
}

type candidateState int

const (
	csRequestWait  candidateState = 1
	csPeerListWait candidateState = 2
)

//NewNodePoolManage is the peerManager creator.
//Apply messages necessary for peer management.
func NewNodePoolManage(StorePath string, nodeMesh nodeMesh, pubhash common.PublicHash) (Manager, error) {
	ns, err := newNodeStore(StorePath)
	if err != nil {
		return nil, err
	}
	pm := &nodePoolManage{
		nodes:        ns,
		nodeMesh:     nodeMesh,
		myPublicHash: pubhash,
		BanPeerInfos: NewByTime(),
	}
	pm.peerStorage = storage.NewPeerStorage(pm.checkClosePeer)
	go pm.rotatePeer()

	return pm, nil
}

func (pm *nodePoolManage) checkClosePeer(ID string) bool {
	peer := pm.nodeMesh.GetPeer(ID)
	if peer == nil {
		return true
	}
	return false
}

// AddNode is used to register additional peers from outside.
func (pm *nodePoolManage) NewNode(addr string, hash string, ping time.Duration) error {
	pm.kickOutPeerStorage()
	ci := peermessage.NewConnectInfo(addr, hash, ping)
	pm.nodes.Store(hash, ci)
	pm.updateScoreBoard(ping, ci)

	pm.nodeMesh.SendPeerList(hash)

	return nil
}

func (pm *nodePoolManage) GetPeerList() ([]string, []string) {
	var ips []string
	var hashs []string
	pm.nodes.Range(func(key string, ci peermessage.ConnectInfo) bool {
		ips = append(ips, ci.Address)
		hashs = append(hashs, ci.Hash)
		return true
	})

	return ips, hashs
}

func (pm *nodePoolManage) AddPeerList(ips []string, hashs []string) {
	pm.putPeerListLock.Lock()
	defer pm.putPeerListLock.Unlock()

	for i, ip := range ips {
		pm.nodes.Store(hashs[i], peermessage.NewConnectInfo(ip, hashs[i], time.Second*5))
	}
}

func (pm *nodePoolManage) updateScoreBoard(duration time.Duration, ci peermessage.ConnectInfo) {
	node := pm.nodes.LoadOrStore(ci.Hash, peermessage.NewConnectInfo(ci.Address, ci.Hash, duration))
	node.PingScoreBoard.Store(ci.Address, duration)
}

func (pm *nodePoolManage) rotatePeer() {
	for {
		if pm.peerStorage.NotEnoughPeer() {
			time.Sleep(time.Second * 5)
		} else {
			time.Sleep(time.Minute * 20)
			pm.kickOutPeerStorage()
		}

		if pm.nodes.Len() < 4 {
			pm.reqPeerList()
		}
		log.Println("peermanager nodes len", pm.nodes.Len())
		pm.appendPeerStorage()
	}
}

func (pm *nodePoolManage) reqPeerList() {
	ps := pm.nodeMesh.Peers()
	for i := pm.requestRotateIndex; i < len(ps); i++ {
		pm.requestRotateIndex = i + 1
		pm.nodeMesh.RequestPeerList(ps[i].ID())
		break
	}
	if pm.requestRotateIndex >= len(ps)-1 {
		pm.requestRotateIndex = 0
	}

}

func (pm *nodePoolManage) appendPeerStorage() {
	for i := pm.nodeRotateIndex; i < pm.nodes.Len(); i++ {
		p := pm.nodes.Get(i)
		pm.nodeRotateIndex = i + 1

		peer := pm.nodeMesh.GetPeer(p.Hash)
		if peer != nil {
			if pm.peerStorage.Have(p.Address) {
				continue
			}
			pm.addConnectedConn(p)
			continue
		}
		var ph common.PublicHash
		copy(ph[:], []byte(p.Hash))
		if ph != pm.myPublicHash {
			pm.nodeMesh.RequestConnect(p.Address, ph)
		}
		break
	}
	if pm.nodeRotateIndex >= pm.nodes.Len()-1 {
		pm.nodeRotateIndex = 0
	}
}

func (pm *nodePoolManage) addConnectedConn(p peermessage.ConnectInfo) {
	pm.peerStorage.Add(p, func(addr string) (time.Duration, bool) {
		if node, has := pm.nodes.Load(addr); has {
			return node.PingScoreBoard.Load(addr)
		}
		return 0, false
	})
}

func (pm *nodePoolManage) kickOutPeerStorage() {
	ps := pm.nodeMesh.Peers()

	if len(ps) > storage.MaxPeerStorageLen()*2 {
		var closePeer peer.Peer
		for _, peer := range ps {
			if closePeer == nil || closePeer.ConnectedTime() > peer.ConnectedTime() {
				closePeer = peer
			}
		}
		if closePeer != nil {
			pm.nodeMesh.RemovePeer(closePeer.ID())
		}
	}
}

// BanPeerInfo is a banned peer information
type BanPeerInfo struct {
	Hash     string
	Timeout  int64
	OverTime int64
}

func (p BanPeerInfo) String() string {
	return fmt.Sprintf("%s Ban over %d", p.Hash, p.OverTime)
}

// ByTime implements sort.Interface for []BanPeerInfo on the Timeout field.
type ByTime struct {
	Arr []*BanPeerInfo
	Map map[string]*BanPeerInfo
}

func NewByTime() *ByTime {
	return &ByTime{
		Arr: []*BanPeerInfo{},
		Map: map[string]*BanPeerInfo{},
	}
}

func (a *ByTime) Len() int           { return len(a.Arr) }
func (a *ByTime) Swap(i, j int)      { a.Arr[i], a.Arr[j] = a.Arr[j], a.Arr[i] }
func (a *ByTime) Less(i, j int) bool { return a.Arr[i].Timeout < a.Arr[j].Timeout }

func (a *ByTime) Add(Hash string, Seconds int64) {
	b, has := a.Map[Hash]
	if !has {
		b = &BanPeerInfo{
			Hash:     Hash,
			Timeout:  time.Now().UnixNano() + (int64(time.Second) * Seconds),
			OverTime: Seconds,
		}
		a.Arr = append(a.Arr, b)
		a.Map[Hash] = b
	} else {
		b.Timeout = Seconds
	}
	sort.Sort(a)
}

func (a *ByTime) Delete(Hash string) {
	i := a.Search(Hash)
	if i < 0 {
		return
	}

	b := a.Arr[i]
	a.Arr = append(a.Arr[:i], a.Arr[i+1:]...)
	delete(a.Map, b.Hash)
}

func (a *ByTime) Search(Hash string) int {
	b, has := a.Map[Hash]
	if !has {
		return -1
	}

	i, j := 0, len(a.Arr)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if !(a.Arr[h].Timeout >= b.Timeout) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}

func (a *ByTime) IsBan(netAddr string) bool {
	now := time.Now().UnixNano()
	var slicePivot = 0
	for i, b := range a.Arr {
		if now < b.Timeout {
			slicePivot = i
			break
		}
		delete(a.Map, a.Arr[i].Hash)
	}

	a.Arr = a.Arr[slicePivot:]
	_, has := a.Map[netAddr]
	return has
}

func (pm *nodePoolManage) Ban(hash string, Seconds uint32) {
	pm.BanPeerInfos.Add(hash, int64(Seconds))
	pm.nodeMesh.RemovePeer(hash)
}

func (pm *nodePoolManage) Unban(Hash string) {
	pm.BanPeerInfos.Delete(Hash)
}
