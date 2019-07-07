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
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

type ObserverMesh struct {
	sync.Mutex
	ob            *Observer
	key           key.Key
	netAddressMap map[common.PublicHash]string
	clientPeerMap map[common.PublicHash]*Peer
	serverPeerMap map[common.PublicHash]*Peer
}

func NewObserverMesh(key key.Key, NetAddressMap map[common.PublicHash]string, ob *Observer) *ObserverMesh {
	ms := &ObserverMesh{
		key:           key,
		netAddressMap: NetAddressMap,
		clientPeerMap: map[common.PublicHash]*Peer{},
		serverPeerMap: map[common.PublicHash]*Peer{},
		ob:            ob,
	}
	return ms
}

// Run starts the observer mesh
func (ms *ObserverMesh) Run(BindAddress string) {
	myPublicHash := common.NewPublicHash(ms.key.PublicKey())
	for PubHash, v := range ms.netAddressMap {
		if PubHash != myPublicHash {
			go func(pubhash common.PublicHash, NetAddr string) {
				time.Sleep(1 * time.Second)
				for {
					ms.Lock()
					_, hasC := ms.clientPeerMap[pubhash]
					_, hasS := ms.serverPeerMap[pubhash]
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
func (ms *ObserverMesh) Peers() []*Peer {
	peerMap := map[common.PublicHash]*Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.pubhash] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.pubhash] = p
	}
	ms.Unlock()

	peers := []*Peer{}
	for _, p := range peerMap {
		peers = append(peers, p)
	}
	return peers
}

// RemovePeer removes peers from the mesh
func (ms *ObserverMesh) RemovePeer(p *Peer) {
	ms.Lock()
	pc, hasClient := ms.clientPeerMap[p.pubhash]
	if hasClient {
		delete(ms.clientPeerMap, p.pubhash)
	}
	ps, hasServer := ms.serverPeerMap[p.pubhash]
	if hasServer {
		delete(ms.serverPeerMap, p.pubhash)
	}
	ms.Unlock()

	if hasClient {
		pc.Close()
	}
	if hasServer {
		ps.Close()
	}
}

// RemovePeerInMap removes peers from the mesh in the map
func (ms *ObserverMesh) RemovePeerInMap(p *Peer, peerMap map[common.PublicHash]*Peer) {
	ms.Lock()
	delete(peerMap, p.pubhash)
	ms.Unlock()

	p.Close()
}

// SendTo sends a message to the observer
func (ms *ObserverMesh) SendTo(PublicHash common.PublicHash, m interface{}) error {
	ms.Lock()
	var p *Peer
	if cp, has := ms.clientPeerMap[PublicHash]; has {
		p = cp
	} else if sp, has := ms.serverPeerMap[PublicHash]; has {
		p = sp
	}
	ms.Unlock()
	if p == nil {
		return ErrNotExistObserverPeer
	}

	if err := p.Send(m); err != nil {
		log.Println(err)
		ms.RemovePeer(p)
	}
	return nil
}

// BroadcastRaw sends a message to all peers
func (ms *ObserverMesh) BroadcastRaw(bs []byte) {
	peerMap := map[common.PublicHash]*Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.pubhash] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.pubhash] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendRaw(bs)
	}
}

// BroadcastMessage sends a message to all peers
func (ms *ObserverMesh) BroadcastMessage(m interface{}) error {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	fc := encoding.Factory("pof.message")
	t, err := fc.TypeOf(m)
	if err != nil {
		return err
	}
	if err := enc.EncodeUint16(t); err != nil {
		return err
	}
	if err := enc.Encode(m); err != nil {
		return err
	}
	data := buffer.Bytes()

	peerMap := map[common.PublicHash]*Peer{}
	ms.Lock()
	for _, p := range ms.clientPeerMap {
		peerMap[p.pubhash] = p
	}
	for _, p := range ms.serverPeerMap {
		peerMap[p.pubhash] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendRaw(data)
	}
	return nil
}

func (ms *ObserverMesh) client(Address string, TargetPubHash common.PublicHash) error {
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

	p := NewPeer(conn, pubhash)
	ms.Lock()
	old, has := ms.clientPeerMap[pubhash]
	ms.clientPeerMap[pubhash] = p
	ms.Unlock()
	if has {
		ms.RemovePeerInMap(old, ms.clientPeerMap)
	}
	defer ms.RemovePeerInMap(p, ms.clientPeerMap)

	if err := ms.handleConnection(p); err != nil {
		log.Println("[handleConnection]", err)
	}
	return nil
}

func (ms *ObserverMesh) server(BindAddress string) error {
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

			p := NewPeer(conn, pubhash)
			ms.Lock()
			old, has := ms.serverPeerMap[pubhash]
			ms.serverPeerMap[pubhash] = p
			ms.Unlock()
			if has {
				ms.RemovePeerInMap(old, ms.serverPeerMap)
			}
			defer ms.RemovePeerInMap(p, ms.serverPeerMap)

			if err := ms.handleConnection(p); err != nil {
				log.Println("[handleConnection]", err)
			}
		}()
	}
}

func (ms *ObserverMesh) handleConnection(p *Peer) error {
	log.Println(common.NewPublicHash(ms.key.PublicKey()).String(), "Connected", p.pubhash.String())

	var pingCount uint64
	pingCountLimit := uint64(3)
	pingTicker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-pingTicker.C:
				if err := p.Send(&p2p.PingMessage{}); err != nil {
					ms.RemovePeer(p)
					return
				}
				if atomic.AddUint64(&pingCount, 1) > pingCountLimit {
					ms.RemovePeer(p)
					return
				}
			}
		}
	}()
	for {
		m, bs, err := p.ReadMessageData()
		if err != nil {
			return err
		}
		atomic.SwapUint64(&pingCount, 0)
		if m == nil {
			// Because a Message is zero size, so do not need to consume the body
			continue
		}

		if err := ms.ob.onObserverRecv(p, m, bs); err != nil {
			return err
		}
	}
}

func (ms *ObserverMesh) recvHandshake(conn net.Conn) error {
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

func (ms *ObserverMesh) sendHandshake(conn net.Conn) (common.PublicHash, error) {
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
	if err := encoding.Unmarshal(bs, &sig); err != nil {
		return common.PublicHash{}, err
	}
	pubkey, err := common.RecoverPubkey(h, sig)
	if err != nil {
		return common.PublicHash{}, err
	}
	pubhash := common.NewPublicHash(pubkey)
	return pubhash, nil
}
