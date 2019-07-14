package p2p

import (
	crand "crypto/rand"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/service/p2p/nodepoolmanage"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/service/p2p/peer"
)

type Handler interface {
	OnConnected(p peer.Peer)
	OnDisconnected(p peer.Peer)
	OnRecv(p peer.Peer, m interface{}) error
}

type NodeConfig struct {
}

type NodeMesh struct {
	sync.Mutex
	BindAddress     string
	key             key.Key
	handler         Handler
	nodeSet         map[common.PublicHash]string
	clientPeerMap   map[string]peer.Peer
	serverPeerMap   map[string]peer.Peer
	nodePoolManager nodepoolmanage.Manager
}

func NewNodeMesh(key key.Key, SeedNodeMap map[common.PublicHash]string, handler Handler, peerStorePath string) *NodeMesh {
	ms := &NodeMesh{
		key:           key,
		handler:       handler,
		nodeSet:       map[common.PublicHash]string{},
		clientPeerMap: map[string]peer.Peer{},
		serverPeerMap: map[string]peer.Peer{},
	}
	manager, err := nodepoolmanage.NewNodePoolManage(peerStorePath, ms)
	if err != nil {
		panic(err)
	}
	ms.nodePoolManager = manager

	for PubHash, v := range SeedNodeMap {
		ms.nodeSet[PubHash] = v
	}
	return ms
}

// Run starts the node mesh
func (ms *NodeMesh) Run(BindAddress string) {
	ms.BindAddress = BindAddress
	myPublicHash := common.NewPublicHash(ms.key.PublicKey())
	for PubHash, v := range ms.nodeSet {
		if PubHash != myPublicHash {
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
							log.Println("[client]", err, NetAddr)
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
	ms.Unlock()

	if hasClient {
		pc.Close()
	}
	if hasServer {
		ps.Close()
	}
}

func (ms *NodeMesh) removePeerInMap(ID string, peerMap map[string]peer.Peer) {
	ms.Lock()
	p, has := ms.clientPeerMap[ID]
	if has {
		delete(ms.clientPeerMap, ID)
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
func (ms *NodeMesh) SendTo(pubhash common.PublicHash, m interface{}) error {
	ID := string(pubhash[:])

	ms.Lock()
	var p peer.Peer
	if cp, has := ms.clientPeerMap[ID]; has {
		p = cp
	} else if sp, has := ms.serverPeerMap[ID]; has {
		p = sp
	}
	ms.Unlock()
	if p == nil {
		return ErrNotExistPeer
	}

	if err := p.Send(m); err != nil {
		log.Println(err)
		ms.RemovePeer(p.ID())
	}
	return nil
}

// ExceptCastLimit sends a message within the given number except the peer
func (ms *NodeMesh) ExceptCastLimit(ID string, m interface{}, Limit int) error {
	data, err := MessageToBytes(m)
	if err != nil {
		return err
	}

	peerMap := map[string]peer.Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		if len(peerMap) >= Limit {
			break
		}
		if p.ID() != ID {
			peerMap[p.ID()] = p
		}
	}
	for _, p := range ms.serverPeerMap {
		if len(peerMap) >= Limit {
			break
		}
		if p.ID() != ID {
			peerMap[p.ID()] = p
		}
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendRaw(data)
	}
	return nil
}

// BroadcastRaw sends a message to all peers
func (ms *NodeMesh) BroadcastRaw(bs []byte) {
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
		p.SendRaw(bs)
	}
}

// BroadcastMessage sends a message to all peers
func (ms *NodeMesh) BroadcastMessage(m interface{}) error {
	data, err := MessageToBytes(m)
	if err != nil {
		return err
	}

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
		p.SendRaw(data)
	}
	return nil
}

func (ms *NodeMesh) RequestConnect(Address string, TargetPubHash common.PublicHash) {
	go ms.client(Address, TargetPubHash)
}

func (ms *NodeMesh) RequestPeerList(targetHash string) {
	pm := &RequestPeerListMessage{}

	var ph common.PublicHash
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, pm)
}

func (ms *NodeMesh) SendPeerList(targetHash string) {
	ips, hashs := ms.nodePoolManager.GetPeerList()
	pm := &PeerListMessage{
		Ips:   ips,
		Hashs: hashs,
	}

	var ph common.PublicHash
	copy(ph[:], []byte(targetHash))
	ms.SendTo(ph, pm)
}

func (ms *NodeMesh) AddPeerList(ips []string, hashs []string) {
	go ms.nodePoolManager.AddPeerList(ips, hashs)
}

func (ms *NodeMesh) client(Address string, TargetPubHash common.PublicHash) error {
	conn, err := net.DialTimeout("tcp", Address, 10*time.Second)
	if err != nil {
		return err
	}
	defer func(addr string) {
		conn.Close()
	}(Address)

	start := time.Now()
	if err := ms.recvHandshake(conn); err != nil {
		log.Println("[recvHandshake]", err)
		return err
	}
	pubhash, bindAddress, err := ms.sendHandshake(conn)
	if err != nil {
		log.Println("[sendHandshake]", err)
		return err
	}
	if pubhash != TargetPubHash {
		return common.ErrInvalidPublicHash
	}
	duration := time.Since(start)
	var ipAddress string
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddress = addr.IP.String()
	}
	ipAddress += bindAddress

	ID := string(pubhash[:])
	ms.nodePoolManager.NewNode(ipAddress, ID, duration)
	p := NewTCPPeer(conn, ID, pubhash.String(), start.UnixNano(), duration)

	ms.Lock()
	old, has := ms.clientPeerMap[ID]
	ms.clientPeerMap[ID] = p
	ms.Unlock()
	if has {
		ms.removePeerInMap(old.ID(), ms.clientPeerMap)
	}
	defer ms.removePeerInMap(p.ID(), ms.clientPeerMap)

	if err := ms.handleConnection(p); err != nil {
		log.Println("[handleConnection]", err)
	}
	return nil
}

func (ms *NodeMesh) server(BindAddress string) error {
	lstn, err := net.Listen("tcp", BindAddress)
	if err != nil {
		return err
	}
	log.Println(common.NewPublicHash(ms.key.PublicKey()), "Start to Listen", BindAddress)
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
				log.Println("[sendHandshake]", err)
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				log.Println("[recvHandshakeAck]", err)
				return
			}
			duration := time.Since(start)
			var ipAddress string
			if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
				ipAddress = addr.IP.String()
			}
			ipAddress += bindAddress

			ID := string(pubhash[:])
			ms.nodePoolManager.NewNode(ipAddress, ID, duration)
			p := NewTCPPeer(conn, ID, pubhash.String(), start.UnixNano(), duration)

			ms.Lock()
			old, has := ms.serverPeerMap[ID]
			ms.serverPeerMap[ID] = p
			ms.Unlock()
			if has {
				ms.removePeerInMap(old.ID(), ms.serverPeerMap)
			}
			defer ms.removePeerInMap(p.ID(), ms.serverPeerMap)

			if err := ms.handleConnection(p); err != nil {
				log.Println("[handleConnection]", err)
			}
		}()
	}
}

func (ms *NodeMesh) handleConnection(p peer.Peer) error {
	// log.Println("Node", common.NewPublicHash(ms.key.PublicKey()).String(), "Node Connected", p.Name())

	ms.handler.OnConnected(p)
	defer ms.handler.OnDisconnected(p)

	var pingCount uint64
	pingCountLimit := uint64(3)
	pingTicker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-pingTicker.C:
				if err := p.Send(&PingMessage{}); err != nil {
					ms.RemovePeer(p.ID())
					return
				}
				if atomic.AddUint64(&pingCount, 1) > pingCountLimit {
					ms.RemovePeer(p.ID())
					return
				}
			}
		}
	}()
	for {
		m, _, err := p.ReadMessageData()
		if err != nil {
			return err
		}
		atomic.StoreUint64(&pingCount, 0)
		if _, is := m.(*PingMessage); is {
			continue
		} else if m == nil {
			return ErrUnknownMessage
		}

		if err := ms.handler.OnRecv(p, m); err != nil {
			return err
		}
	}
}

func (ms *NodeMesh) recvHandshake(conn net.Conn) error {
	//log.Println("recvHandshake")
	req := make([]byte, 40)
	if _, err := FillBytes(conn, req); err != nil {
		return err
	}
	timestamp := binary.LittleEndian.Uint64(req[32:])
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return ErrInvalidHandshake
	}
	//log.Println("sendHandshakeAck")
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
	//log.Println("sendHandshake")
	req := make([]byte, 40)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, "", err
	}
	binary.LittleEndian.PutUint64(req[32:], uint64(time.Now().UnixNano()))
	if _, err := conn.Write(req); err != nil {
		return common.PublicHash{}, "", err
	}
	//log.Println("recvHandshakeAsk")
	h := hash.Hash(req)
	bs := make([]byte, common.SignatureSize)
	if _, err := FillBytes(conn, bs); err != nil {
		return common.PublicHash{}, "", err
	}
	var sig common.Signature
	copy(sig[:], bs)
	pubkey, err := common.RecoverPubkey(h, sig)
	if err != nil {
		return common.PublicHash{}, "", err
	}
	pubhash := common.NewPublicHash(pubkey)

	bs = make([]byte, 1)
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
