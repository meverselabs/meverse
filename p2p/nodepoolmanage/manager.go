package nodepoolmanage

import (
	"sync"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/meverselabs/meverse/p2p/peermessage"
	"github.com/meverselabs/meverse/p2p/storage"
	"github.com/pkg/errors"
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
	RemovePeer(hash string)
	Ban(hash string)
	Unban(Hash string)
}

type nodeMesh interface {
	RequestConnect(Address string, TargetPubKey common.PublicKey)
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
	BanPeerInfos       *BanAlways
	myPublicKey        common.PublicKey

	putPeerListLock sync.Mutex
}

type candidateState int

const (
	csRequestWait  candidateState = 1
	csPeerListWait candidateState = 2
)

//NewNodePoolManage is the peerManager creator.
//Apply messages necessary for peer management.
func NewNodePoolManage(StorePath string, nodeMesh nodeMesh, pubkey common.PublicKey) (Manager, error) {
	ns, err := newNodeStore(StorePath)
	if err != nil {
		return nil, err
	}
	pm := &nodePoolManage{
		nodes:        ns,
		nodeMesh:     nodeMesh,
		myPublicKey:  pubkey,
		BanPeerInfos: NewBanAlways(),
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

func (pm *nodePoolManage) RemovePeer(hash string) {
	pm.nodes.Delete(hash)
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
		// log.Println("peermanager nodes len", pm.nodes.Len())
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
		if !pm.BanPeerInfos.IsBan(p.Hash) {
			var ph common.PublicKey
			copy(ph[:], []byte(p.Hash))
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

// BanAlways implements sort.Interface for []BanPeerInfo on the Timeout field.
type BanAlways struct {
	Map map[string]bool
}

func NewBanAlways() *BanAlways {
	return &BanAlways{
		Map: map[string]bool{},
	}
}

func (a *BanAlways) Add(Hash string) {
	a.Map[Hash] = true
}

func (a *BanAlways) Delete(Hash string) {
	delete(a.Map, Hash)
}

func (a *BanAlways) IsBan(hash string) bool {
	return a.Map[hash]
}

func (pm *nodePoolManage) Ban(hash string) {
	pm.BanPeerInfos.Add(hash)
	pm.nodeMesh.RemovePeer(hash)
}

func (pm *nodePoolManage) Unban(Hash string) {
	pm.BanPeerInfos.Delete(Hash)
}
