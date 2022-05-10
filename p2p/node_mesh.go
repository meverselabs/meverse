package p2p

import (
	"log"
	"math/big"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/p2p/nodepoolmanage"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

// Handler is a interface for connection events
type Handler interface {
	OnConnected(p peer.Peer)
	OnDisconnected(p peer.Peer)
	OnRecv(p peer.Peer, bs []byte) error
}

// NodeMesh is a mesh for networking between nodes
type NodeMesh struct {
	sync.Mutex
	BindAddress     string
	chainID         *big.Int
	key             key.Key
	handler         Handler
	myPublicKey     common.PublicKey
	nodeSet         map[common.PublicKey]string
	peerIDs         []string
	badPointMap     map[string]int
	clientPeerMap   map[string]peer.Peer
	serverPeerMap   map[string]peer.Peer
	nodePoolManager nodepoolmanage.Manager
}

// NewNodeMesh returns a NodeMesh
func NewNodeMesh(ChainID *big.Int, key key.Key, SeedNodeMap map[common.PublicKey]string, handler Handler, peerStorePath string) *NodeMesh {
	ms := &NodeMesh{
		chainID:       ChainID,
		key:           key,
		handler:       handler,
		myPublicKey:   key.PublicKey(),
		nodeSet:       map[common.PublicKey]string{},
		peerIDs:       []string{},
		badPointMap:   map[string]int{},
		clientPeerMap: map[string]peer.Peer{},
		serverPeerMap: map[string]peer.Peer{},
	}
	manager, err := nodepoolmanage.NewNodePoolManage(peerStorePath, ms, ms.myPublicKey)
	if err != nil {
		panic(err)
	}
	ms.nodePoolManager = manager
	ms.nodePoolManager.Ban(string(ms.myPublicKey[:]))

	for PubHash, v := range SeedNodeMap {
		ms.nodeSet[PubHash] = v
	}
	return ms
}

// Run starts the node mesh
func (ms *NodeMesh) Run(BindAddress string) {
	ms.BindAddress = BindAddress
	for PubHash, v := range ms.nodeSet {
		if PubHash != ms.myPublicKey {
			go func(pubhash common.PublicKey, NetAddr string) {
				time.Sleep(1 * time.Second)
				for {
					ID := string(pubhash[:])
					ms.Lock()
					_, hasInSet := ms.nodeSet[pubhash]
					_, hasC := ms.clientPeerMap[ID]
					_, hasS := ms.serverPeerMap[ID]
					ms.Unlock()
					if !hasInSet {
						return
					}
					if !hasC && !hasS {
						if err := ms.client(NetAddr, pubhash); err != nil {
							log.Printf("[client] %v %+v\n", NetAddr, err)
						}
					}
					time.Sleep(30 * time.Second)
				}
			}(PubHash, v)
		}
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			ms.Lock()
			for ID, point := range ms.badPointMap {
				if point <= 1 {
					delete(ms.badPointMap, ID)
				} else {
					ms.badPointMap[ID] = point - 1
				}
			}
			ms.Unlock()
		}
	}()
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

func (ms *NodeMesh) HasPeer() bool {
	ms.Lock()
	defer ms.Unlock()

	return len(ms.clientPeerMap) > 0 || len(ms.serverPeerMap) > 0
}

// Peers returns peers of the node mesh
func (ms *NodeMesh) Peers() []peer.Peer {
	peerMap := map[string]peer.Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.ID()] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.ID()] = p
	}
	ms.Unlock()

	peers := []peer.Peer{}
	for _, p := range peerMap {
		peers = append(peers, p)
	}
	return peers
}

// AddBadPoint adds bad points to to the peer
func (ms *NodeMesh) AddBadPoint(ID string, Point int) {
	ms.Lock()
	ms.badPointMap[ID] += Point
	ms.Unlock()

	if ms.badPointMap[ID] >= 100 {
		ms.RemovePeer(ID)
	}
}

// RemovePeer removes peers from the mesh
func (ms *NodeMesh) RemovePeer(ID string) {
	ms.Lock()
	pc, hasClient := ms.clientPeerMap[ID]
	if hasClient {
		delete(ms.clientPeerMap, ID)
	}
	ps, hasServer := ms.serverPeerMap[ID]
	if hasServer {
		delete(ms.serverPeerMap, ID)
	}
	if hasClient || hasServer {
		ms.updatePeerIDs()
		delete(ms.badPointMap, ID)
	}
	ms.Unlock()

	if hasClient {
		pc.Close()
	}
	if hasServer {
		ps.Close()
	}
}

func (ms *NodeMesh) updatePeerIDs() {
	peerIDMap := map[string]bool{}
	for _, p := range ms.clientPeerMap {
		peerIDMap[p.ID()] = true
	}
	for _, p := range ms.serverPeerMap {
		peerIDMap[p.ID()] = true
	}
	peerIDs := []string{}
	for k := range peerIDMap {
		peerIDs = append(peerIDs, k)
	}
	sort.Strings(peerIDs)
	ms.peerIDs = peerIDs
}

func (ms *NodeMesh) removePeerInMap(ID string, peerMap map[string]peer.Peer) {
	ms.Lock()
	p, has := ms.clientPeerMap[ID]
	if has {
		delete(ms.clientPeerMap, ID)
		ms.updatePeerIDs()
	}
	ms.Unlock()

	if has {
		p.Close()
	}
}

// GetPeer returns the peer of the id
func (ms *NodeMesh) GetPeer(ID string) peer.Peer {
	ms.Lock()
	defer ms.Unlock()

	if cp, has := ms.clientPeerMap[ID]; has {
		return cp
	} else if sp, has := ms.serverPeerMap[ID]; has {
		return sp
	}
	return nil
}

// SendTo sends a message to the node
func (ms *NodeMesh) SendTo(pubhash common.PublicKey, bs []byte) {
	ID := string(pubhash[:])

	ms.Lock()
	var p peer.Peer
	if cp, has := ms.clientPeerMap[ID]; has {
		p = cp
	} else if sp, has := ms.serverPeerMap[ID]; has {
		p = sp
	}
	ms.Unlock()

	if p != nil {
		p.SendPacket(bs)
	}
}

// ExceptCast sends a message except the peer
func (ms *NodeMesh) ExceptCast(ID string, bs []byte) {
	peerMap := map[string]peer.Peer{}

	ms.Lock()
	for _, p := range ms.clientPeerMap {
		if p.ID() != ID {
			peerMap[p.ID()] = p
		}
	}
	for _, p := range ms.serverPeerMap {
		if p.ID() != ID {
			peerMap[p.ID()] = p
		}
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendPacket(bs)
	}
}

// BroadcastPacket sends a packet to all peers
func (ms *NodeMesh) BroadcastPacket(bs []byte) {
	peerMap := map[string]peer.Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.ID()] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.ID()] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendPacket(bs)
	}
}

func (ms *NodeMesh) RequestConnect(Address string, TargetPubHash common.PublicKey) {
	go ms.client(Address, TargetPubHash)
}

func (ms *NodeMesh) RequestPeerList(targetHash string) {
	pm := &RequestPeerListMessage{}

	var ph common.PublicKey
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, MessageToPacket(pm))
}

func (ms *NodeMesh) SendPeerList(targetHash string) {
	ips, hashs := ms.nodePoolManager.GetPeerList()
	pm := &PeerListMessage{
		Ips:   ips,
		Hashs: hashs,
	}

	var ph common.PublicKey
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, MessageToPacket(pm))
}

func (ms *NodeMesh) AddPeerList(ips []string, hashs []string) {
	go ms.nodePoolManager.AddPeerList(ips, hashs)
}

func (ms *NodeMesh) client(Address string, TargetPubKey common.PublicKey) error {
	log.Println("ConnectingTo", Address, TargetPubKey.Address().String())

	if TargetPubKey == ms.myPublicKey {
		ms.nodePoolManager.Ban(string(TargetPubKey[:]))
		return errors.WithStack(ErrSelfConnection)
	}

	conn, err := net.DialTimeout("tcp", Address, 10*time.Second)
	if err != nil {
		return errors.WithStack(err)
	}
	defer conn.Close()

	start := time.Now()
	if err := ms.recvHandshake(conn); err != nil {
		log.Printf("[recvHandshake] %+v\n", err)
		return err
	}
	pubkey, bindAddress, err := ms.sendHandshake(conn)
	if err != nil {
		log.Printf("[sendHandshake] %+v\n", err)
		return err
	}
	if pubkey == ms.myPublicKey {
		ms.nodePoolManager.Ban(string(TargetPubKey[:]))
		ms.nodePoolManager.Ban(string(pubkey[:]))
		return errors.WithStack(ErrSelfConnection)
	}
	if pubkey != TargetPubKey {
		return errors.WithStack(common.ErrInvalidPublicKey)
	}
	//duration := time.Since(start)
	var ipAddress string
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddress = addr.IP.String()
	}
	ipAddress += bindAddress

	ID := string(pubkey[:])
	//ms.nodePoolManager.NewNode(ipAddress, ID, duration)
	p := NewTCPAsyncPeer(conn, ID, pubkey.String(), start.UnixNano())

	ms.Lock()
	old, has := ms.clientPeerMap[ID]
	ms.clientPeerMap[ID] = p
	if !has {
		ms.updatePeerIDs()
	}
	ms.Unlock()
	if has {
		ms.removePeerInMap(old.ID(), ms.clientPeerMap)
	}
	defer ms.removePeerInMap(p.ID(), ms.clientPeerMap)

	if err := ms.handleConnection(p); err != nil {
		log.Printf("[handleConnection] %+v\n", err)
	}
	return nil
}

func (ms *NodeMesh) server(BindAddress string) error {
	lstn, err := net.Listen("tcp", BindAddress)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Println(ms.key.PublicKey(), "Start to Listen", BindAddress)
	for {
		conn, err := lstn.Accept()
		if err != nil {
			return errors.WithStack(err)
		}
		go func() {
			defer conn.Close()

			start := time.Now()
			pubhash, bindAddress, err := ms.sendHandshake(conn)
			if err != nil {
				log.Printf("[sendHandshake] %+v\n", err)
				return
			}
			if pubhash == ms.myPublicKey {
				ms.nodePoolManager.RemovePeer(string(pubhash[:]))
				ms.nodePoolManager.Ban(string(pubhash[:]))
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				log.Printf("[recvHandshakeAck] %+v\n", err)
				return
			}
			//duration := time.Since(start)
			var ipAddress string
			if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
				ipAddress = addr.IP.String()
			}
			ipAddress += bindAddress

			ID := string(pubhash[:])
			//ms.nodePoolManager.NewNode(ipAddress, ID, duration)
			p := NewTCPAsyncPeer(conn, ID, pubhash.String(), start.UnixNano())

			log.Println("ConnectedFrom", pubhash.String())

			ms.Lock()
			old, has := ms.serverPeerMap[ID]
			ms.serverPeerMap[ID] = p
			if !has {
				ms.updatePeerIDs()
			}
			ms.Unlock()
			if has {
				ms.removePeerInMap(old.ID(), ms.serverPeerMap)
			}
			defer ms.removePeerInMap(p.ID(), ms.serverPeerMap)

			if err := ms.handleConnection(p); err != nil {
				log.Printf("[handleConnection] %+v\n", err)
			}
		}()
	}
}

func (ms *NodeMesh) handleConnection(p peer.Peer) error {
	// log.Println("Node", ms.key.PublicKey().String(), "Node Connected", p.Name())

	ms.handler.OnConnected(p)
	defer ms.handler.OnDisconnected(p)

	for {
		bs, err := p.ReadPacket()
		if err != nil {
			return err
		}
		if err := ms.handler.OnRecv(p, bs); err != nil {
			return err
		}
	}
}

func (ms *NodeMesh) recvHandshake(conn net.Conn) error {
	//log.Println("recvHandshake")
	req, _, err := bin.ReadBytes(conn)
	if err != nil {
		return err
	}
	ChainID, _, timestamp, err := RecoveryHandSharkBs(req)
	if err != nil {
		return err
	}
	if ChainID.Cmp(ms.chainID) != 0 {
		log.Println("not match chainin from :", ChainID.String(), "has:", ms.chainID.String())
		return errors.WithStack(chain.ErrInvalidChainID)
	}
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return errors.WithStack(ErrInvalidHandshake)
	}
	//log.Println("sendHandshakeAck")
	if sig, err := ms.key.Sign(hash.Hash(req)); err != nil {
		return err
	} else if _, err := bin.WriteBytes(conn, sig); err != nil {
		return err
	}

	ba := []byte(ms.BindAddress)
	length := byte(uint8(len(ba)))
	if _, err := conn.Write([]byte{length}); err != nil {
		return errors.WithStack(err)
	}
	if _, err := conn.Write(ba); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (ms *NodeMesh) sendHandshake(conn net.Conn) (common.PublicKey, string, error) {
	//log.Println("sendHandshake")
	rn := rand.Uint64()
	req, err := MakeHandSharkBs(ms.chainID, rn, uint64(time.Now().UnixNano()))
	if err != nil {
		return common.PublicKey{}, "", err
	}
	_, err = bin.WriteBytes(conn, req)
	if err != nil {
		return common.PublicKey{}, "", err
	}

	//log.Println("recvHandshakeAsk")
	sig, _, err := bin.ReadBytes(conn)
	if err != nil {
		return common.PublicKey{}, "", err
	}
	pubkey, err := common.RecoverPubkey(ms.chainID, hash.Hash(req), sig)
	if err != nil {
		return common.PublicKey{}, "", err
	}
	bs := make([]byte, 1)
	if _, err := FillBytes(conn, bs); err != nil {
		return common.PublicKey{}, "", err
	}
	length := uint8(bs[0])
	bs = make([]byte, length)
	if _, err := FillBytes(conn, bs); err != nil {
		return common.PublicKey{}, "", err
	}
	bindAddres := string(bs)

	return pubkey, bindAddres, nil
}
