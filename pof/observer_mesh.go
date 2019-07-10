package pof

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

type ObserverNodeMesh struct {
	sync.Mutex
	ob            *ObserverNode
	key           key.Key
	netAddressMap map[common.PublicHash]string
	clientPeerMap map[string]p2p.Peer
	serverPeerMap map[string]p2p.Peer
}

func NewObserverNodeMesh(key key.Key, NetAddressMap map[common.PublicHash]string, ob *ObserverNode) *ObserverNodeMesh {
	ms := &ObserverNodeMesh{
		key:           key,
		netAddressMap: NetAddressMap,
		clientPeerMap: map[string]p2p.Peer{},
		serverPeerMap: map[string]p2p.Peer{},
		ob:            ob,
	}
	return ms
}

// Run starts the observer mesh
func (ms *ObserverNodeMesh) Run(BindAddress string) {
	myPublicHash := common.NewPublicHash(ms.key.PublicKey())
	for PubHash, v := range ms.netAddressMap {
		if PubHash != myPublicHash {
			go func(pubhash common.PublicHash, NetAddr string) {
				time.Sleep(1 * time.Second)
				for {
					ID := string(pubhash[:])
					ms.Lock()
					_, hasC := ms.clientPeerMap[ID]
					_, hasS := ms.serverPeerMap[ID]
					ms.Unlock()
					if !hasC && !hasS {
						if err := ms.client(NetAddr, pubhash); err != nil {
							log.Println("[client]", err, NetAddr)
						}
					}
					time.Sleep(1 * time.Second)
				}
			}(PubHash, v)
		}
	}
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

// Peers returns peers of the observer mesh
func (ms *ObserverNodeMesh) Peers() []p2p.Peer {
	peerMap := map[string]p2p.Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.ID()] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.ID()] = p
	}
	ms.Unlock()

	peers := []p2p.Peer{}
	for _, p := range peerMap {
		peers = append(peers, p)
	}
	return peers
}

// RemovePeer removes peers from the mesh
func (ms *ObserverNodeMesh) RemovePeer(ID string) {
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

func (ms *ObserverNodeMesh) removePeerInMap(ID string, peerMap map[string]p2p.Peer) {
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

// SendTo sends a message to the observer
func (ms *ObserverNodeMesh) SendTo(pubhash common.PublicHash, m interface{}) error {
	ID := string(pubhash[:])

	ms.Lock()
	var p p2p.Peer
	if cp, has := ms.clientPeerMap[ID]; has {
		p = cp
	} else if sp, has := ms.serverPeerMap[ID]; has {
		p = sp
	}
	ms.Unlock()
	if p == nil {
		return ErrNotExistObserverPeer
	}

	if err := p.Send(m); err != nil {
		log.Println(err)
		ms.RemovePeer(p.ID())
	}
	return nil
}

// BroadcastRaw sends a message to all peers
func (ms *ObserverNodeMesh) BroadcastRaw(bs []byte) {
	peerMap := map[string]p2p.Peer{}
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
func (ms *ObserverNodeMesh) BroadcastMessage(m interface{}) error {
	var buffer bytes.Buffer
	fc := encoding.Factory("message")
	t, err := fc.TypeOf(m)
	if err != nil {
		return err
	}
	buffer.Write(util.Uint16ToBytes(t))
	enc := encoding.NewEncoder(&buffer)
	if err := enc.Encode(m); err != nil {
		return err
	}
	data := buffer.Bytes()

	peerMap := map[string]p2p.Peer{}
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

func (ms *ObserverNodeMesh) client(Address string, TargetPubHash common.PublicHash) error {
	conn, err := net.DialTimeout("tcp", Address, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := ms.recvHandshake(conn); err != nil {
		log.Println("[recvHandshake]", err)
		return err
	}
	pubhash, err := ms.sendHandshake(conn)
	if err != nil {
		log.Println("[sendHandshake]", err)
		return err
	}
	if pubhash != TargetPubHash {
		return common.ErrInvalidPublicHash
	}
	if _, has := ms.netAddressMap[pubhash]; !has {
		return ErrInvalidObserverKey
	}

	ID := string(pubhash[:])
	p := p2p.NewTCPPeer(conn, ID, pubhash.String())
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

func (ms *ObserverNodeMesh) server(BindAddress string) error {
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

			pubhash, err := ms.sendHandshake(conn)
			if err != nil {
				log.Println("[sendHandshake]", err)
				return
			}
			if _, has := ms.netAddressMap[pubhash]; !has {
				log.Println("ErrInvalidPublicHash")
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				log.Println("[recvHandshakeAck]", err)
				return
			}

			ID := string(pubhash[:])
			p := p2p.NewTCPPeer(conn, ID, pubhash.String())
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

func (ms *ObserverNodeMesh) handleConnection(p p2p.Peer) error {
	log.Println("Observer", common.NewPublicHash(ms.key.PublicKey()).String(), "Observer Connected", p.Name())

	var pingCount uint64
	pingCountLimit := uint64(3)
	pingTicker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-pingTicker.C:
				if err := p.Send(&p2p.PingMessage{}); err != nil {
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
		if _, is := m.(*p2p.PingMessage); is {
			continue
		} else if m == nil {
			return p2p.ErrUnknownMessage
		}

		if err := ms.ob.onObserverRecv(p, m); err != nil {
			return err
		}
	}
}

func (ms *ObserverNodeMesh) recvHandshake(conn net.Conn) error {
	//log.Println("recvHandshake")
	req := make([]byte, 40)
	if _, err := p2p.FillBytes(conn, req); err != nil {
		return err
	}
	timestamp := binary.LittleEndian.Uint64(req[32:])
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return p2p.ErrInvalidHandshake
	}
	//log.Println("sendHandshakeAck")
	h := hash.Hash(req)
	if sig, err := ms.key.Sign(h); err != nil {
		return err
	} else if _, err := conn.Write(sig[:]); err != nil {
		return err
	}
	return nil
}

func (ms *ObserverNodeMesh) sendHandshake(conn net.Conn) (common.PublicHash, error) {
	//log.Println("sendHandshake")
	req := make([]byte, 40)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, err
	}
	binary.LittleEndian.PutUint64(req[32:], uint64(time.Now().UnixNano()))
	if _, err := conn.Write(req); err != nil {
		return common.PublicHash{}, err
	}
	//log.Println("recvHandshakeAsk")
	h := hash.Hash(req)
	bs := make([]byte, common.SignatureSize)
	if _, err := p2p.FillBytes(conn, bs); err != nil {
		return common.PublicHash{}, err
	}
	var sig common.Signature
	copy(sig[:], bs)
	pubkey, err := common.RecoverPubkey(h, sig)
	if err != nil {
		return common.PublicHash{}, err
	}
	pubhash := common.NewPublicHash(pubkey)
	return pubhash, nil
}
