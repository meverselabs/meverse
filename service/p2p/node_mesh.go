package p2p

import (
	crand "crypto/rand"
	"log"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/service/p2p/nodepoolmanage"
	"github.com/fletaio/fleta/service/p2p/peer"
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
	BindAddress      string
	chainID          uint8
	key              key.Key
	handler          Handler
	myPublicHash     common.PublicHash
	nodeSet          map[common.PublicHash]string
	peerIDs          []string
	clientPeerMap    map[string]peer.Peer
	serverPeerMap    map[string]peer.Peer
	nodePoolManager  nodepoolmanage.Manager
	limitCastIndexes []int
	limitCastHead    int
}

// NewNodeMesh returns a NodeMesh
func NewNodeMesh(ChainID uint8, key key.Key, SeedNodeMap map[common.PublicHash]string, handler Handler, peerStorePath string) *NodeMesh {
	ms := &NodeMesh{
		chainID:          ChainID,
		key:              key,
		handler:          handler,
		myPublicHash:     common.NewPublicHash(key.PublicKey()),
		nodeSet:          map[common.PublicHash]string{},
		peerIDs:          []string{},
		clientPeerMap:    map[string]peer.Peer{},
		serverPeerMap:    map[string]peer.Peer{},
		limitCastIndexes: rand.Perm(2000),
	}
	manager, err := nodepoolmanage.NewNodePoolManage(peerStorePath, ms, ms.myPublicHash)
	if err != nil {
		panic(err)
	}
	ms.nodePoolManager = manager
	ms.nodePoolManager.Ban(string(ms.myPublicHash[:]))

	for PubHash, v := range SeedNodeMap {
		ms.nodeSet[PubHash] = v
	}
	return ms
}

// Run starts the node mesh
func (ms *NodeMesh) Run(BindAddress string) {
	ms.BindAddress = BindAddress
	for PubHash, v := range ms.nodeSet {
		if PubHash != ms.myPublicHash {
			go func(pubhash common.PublicHash, NetAddr string) {
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
							rlog.Println("[client]", err, NetAddr)
						}
					}
					time.Sleep(30 * time.Second)
				}
			}(PubHash, v)
		}
	}
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
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
func (ms *NodeMesh) SendTo(pubhash common.PublicHash, bs []byte) {
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

// ExceptCastLimit sends a message within the given number except the peer
func (ms *NodeMesh) ExceptCastLimit(ID string, bs []byte, Limit int) {
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

	if len(peerMap) <= Limit {
		for _, p := range peerMap {
			p.SendPacket(bs)
		}
	} else {
		Remain := Limit
		for len(peerMap) > 0 && Remain > 0 {
			id := ms.peerIDs[ms.limitCastIndexes[ms.limitCastHead]%len(ms.peerIDs)]
			ms.limitCastHead = (ms.limitCastHead + 1) % len(ms.limitCastIndexes)
			if p, has := peerMap[id]; has {
				p.SendPacket(bs)
				delete(peerMap, id)
				Remain--
			}
		}
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

func (ms *NodeMesh) RequestConnect(Address string, TargetPubHash common.PublicHash) {
	go ms.client(Address, TargetPubHash)
}

func (ms *NodeMesh) RequestPeerList(targetHash string) {
	pm := &RequestPeerListMessage{}

	var ph common.PublicHash
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, MessageToPacket(pm))
}

func (ms *NodeMesh) SendPeerList(targetHash string) {
	ips, hashs := ms.nodePoolManager.GetPeerList()
	pm := &PeerListMessage{
		Ips:   ips,
		Hashs: hashs,
	}

	var ph common.PublicHash
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, MessageToPacket(pm))
}

func (ms *NodeMesh) AddPeerList(ips []string, hashs []string) {
	go ms.nodePoolManager.AddPeerList(ips, hashs)
}

func (ms *NodeMesh) client(Address string, TargetPubHash common.PublicHash) error {
	log.Println("ConnectingTo", Address, TargetPubHash.String())

	if TargetPubHash == ms.myPublicHash {
		ms.nodePoolManager.Ban(string(TargetPubHash[:]))
		return ErrSelfConnection
	}

	conn, err := net.DialTimeout("tcp", Address, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	start := time.Now()
	if err := ms.recvHandshake(conn); err != nil {
		rlog.Println("[recvHandshake]", err)
		return err
	}
	pubhash, bindAddress, err := ms.sendHandshake(conn)
	if err != nil {
		rlog.Println("[sendHandshake]", err)
		return err
	}
	if pubhash == ms.myPublicHash {
		ms.nodePoolManager.Ban(string(TargetPubHash[:]))
		ms.nodePoolManager.Ban(string(pubhash[:]))
		return ErrSelfConnection
	}
	if pubhash != TargetPubHash {
		return common.ErrInvalidPublicHash
	}
	//duration := time.Since(start)
	var ipAddress string
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddress = addr.IP.String()
	}
	ipAddress += bindAddress

	ID := string(pubhash[:])
	//ms.nodePoolManager.NewNode(ipAddress, ID, duration)
	p := NewTCPPeer(conn, ID, pubhash.String(), start.UnixNano())

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
		rlog.Println("[handleConnection]", err)
	}
	return nil
}

func (ms *NodeMesh) server(BindAddress string) error {
	lstn, err := net.Listen("tcp", BindAddress)
	if err != nil {
		return err
	}
	rlog.Println(common.NewPublicHash(ms.key.PublicKey()), "Start to Listen", BindAddress)
	for {
		conn, err := lstn.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()

			start := time.Now()
			pubhash, bindAddress, err := ms.sendHandshake(conn)
			if err != nil {
				rlog.Println("[sendHandshake]", err)
				return
			}
			if pubhash == ms.myPublicHash {
				ms.nodePoolManager.RemovePeer(string(pubhash[:]))
				ms.nodePoolManager.Ban(string(pubhash[:]))
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				rlog.Println("[recvHandshakeAck]", err)
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
			p := NewTCPPeer(conn, ID, pubhash.String(), start.UnixNano())

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
				rlog.Println("[handleConnection]", err)
			}
		}()
	}
}

func (ms *NodeMesh) handleConnection(p peer.Peer) error {
	// rlog.Println("Node", common.NewPublicHash(ms.key.PublicKey()).String(), "Node Connected", p.Name())

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
	//rlog.Println("recvHandshake")
	req := make([]byte, 40)
	if _, err := FillBytes(conn, req); err != nil {
		return err
	}
	ChainID := req[0]
	if ChainID != ms.chainID {
		return chain.ErrInvalidChainID
	}
	timestamp := binutil.LittleEndian.Uint64(req[32:])
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return ErrInvalidHandshake
	}
	//rlog.Println("sendHandshakeAck")
	h := hash.Hash(req)
	if sig, err := ms.key.Sign(h); err != nil {
		return err
	} else if _, err := conn.Write(sig[:]); err != nil {
		return err
	}

	ba := []byte(ms.BindAddress)
	length := byte(uint8(len(ba)))
	if _, err := conn.Write([]byte{length}); err != nil {
		return err
	}
	if _, err := conn.Write(ba); err != nil {
		return err
	}
	return nil
}

func (ms *NodeMesh) sendHandshake(conn net.Conn) (common.PublicHash, string, error) {
	//rlog.Println("sendHandshake")
	req := make([]byte, 40)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, "", err
	}
	req[0] = ms.chainID
	binutil.LittleEndian.PutUint64(req[32:], uint64(time.Now().UnixNano()))
	if _, err := conn.Write(req); err != nil {
		return common.PublicHash{}, "", err
	}
	//rlog.Println("recvHandshakeAsk")
	var sig common.Signature
	if _, err := FillBytes(conn, sig[:]); err != nil {
		return common.PublicHash{}, "", err
	}
	pubkey, err := common.RecoverPubkey(hash.Hash(req), sig)
	if err != nil {
		return common.PublicHash{}, "", err
	}
	pubhash := common.NewPublicHash(pubkey)

	bs := make([]byte, 1)
	if _, err := FillBytes(conn, bs); err != nil {
		return common.PublicHash{}, "", err
	}
	length := uint8(bs[0])
	bs = make([]byte, length)
	if _, err := FillBytes(conn, bs); err != nil {
		return common.PublicHash{}, "", err
	}
	bindAddres := string(bs)

	return pubhash, bindAddres, nil
}
