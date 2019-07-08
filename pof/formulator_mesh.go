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

type FormulatorMesh struct {
	sync.Mutex
	fr            *Formulator
	key           key.Key
	netAddressMap map[common.PublicHash]string
	peerMap       map[common.PublicHash]*Peer
}

func NewFormulatorMesh(key key.Key, NetAddressMap map[common.PublicHash]string, fr *Formulator) *FormulatorMesh {
	ms := &FormulatorMesh{
		key:           key,
		netAddressMap: NetAddressMap,
		peerMap:       map[common.PublicHash]*Peer{},
		fr:            fr,
	}
	return ms
}

// Run starts the formulator mesh
func (ms *FormulatorMesh) Run() {
	for PubHash, v := range ms.netAddressMap {
		go func(pubhash common.PublicHash, NetAddr string) {
			time.Sleep(1 * time.Second)
			for {
				ms.Lock()
				_, has := ms.peerMap[pubhash]
				ms.Unlock()
				if !has {
					if err := ms.client(NetAddr, pubhash); err != nil {
						log.Println("[client]", err, NetAddr)
					}
				}
				time.Sleep(1 * time.Second)
			}
		}(PubHash, v)
	}
}

// Peers returns peers of the formulator mesh
func (ms *FormulatorMesh) Peers() []*Peer {
	ms.Lock()
	defer ms.Unlock()

	peers := []*Peer{}
	for _, p := range ms.peerMap {
		peers = append(peers, p)
	}
	return peers
}

// RemovePeer removes peers from the mesh
func (ms *FormulatorMesh) RemovePeer(p *Peer) {
	ms.Lock()
	p, has := ms.peerMap[p.pubhash]
	if has {
		delete(ms.peerMap, p.pubhash)
	}
	ms.Unlock()

	if has {
		p.Close()
	}
}

// RemovePeerInMap removes peers from the mesh in the map
func (ms *FormulatorMesh) RemovePeerInMap(p *Peer, peerMap map[common.PublicHash]*Peer) {
	ms.Lock()
	delete(peerMap, p.pubhash)
	ms.Unlock()

	p.Close()
}

// SendTo sends a message to the observer
func (ms *FormulatorMesh) SendTo(PublicHash common.PublicHash, m interface{}) error {
	ms.Lock()
	p, has := ms.peerMap[PublicHash]
	ms.Unlock()
	if !has {
		return ErrNotExistObserverPeer
	}

	if err := p.Send(m); err != nil {
		log.Println(err)
		ms.RemovePeer(p)
	}
	return nil
}

// BroadcastRaw sends a message to all peers
func (ms *FormulatorMesh) BroadcastRaw(bs []byte) {
	peerMap := map[common.PublicHash]*Peer{}
	ms.Lock()
	for _, p := range ms.peerMap {
		peerMap[p.pubhash] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendRaw(bs)
	}
}

// BroadcastMessage sends a message to all peers
func (ms *FormulatorMesh) BroadcastMessage(m interface{}) error {
	var buffer bytes.Buffer
	fc := encoding.Factory("pof.message")
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

	peerMap := map[common.PublicHash]*Peer{}
	ms.Lock()
	for _, p := range ms.peerMap {
		peerMap[p.pubhash] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendRaw(data)
	}
	return nil
}

func (ms *FormulatorMesh) client(Address string, TargetPubHash common.PublicHash) error {
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

	p := NewPeer(conn)
	p.pubhash = pubhash
	ms.Lock()
	old, has := ms.peerMap[pubhash]
	ms.peerMap[pubhash] = p
	ms.Unlock()
	if has {
		ms.RemovePeerInMap(old, ms.peerMap)
	}
	defer ms.RemovePeerInMap(p, ms.peerMap)

	if err := ms.handleConnection(p); err != nil {
		log.Println("[handleConnection]", err)
	}
	return nil
}

func (ms *FormulatorMesh) handleConnection(p *Peer) error {
	log.Println(common.NewPublicHash(ms.key.PublicKey()).String(), "Connected", p.pubhash.String())

	ms.fr.OnObserverConnected(p)
	defer ms.fr.OnObserverDisconnected(p)

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
		m, _, err := p.ReadMessageData()
		if err != nil {
			return err
		}
		atomic.SwapUint64(&pingCount, 0)
		if _, is := m.(*p2p.PingMessage); is {
			continue
		}

		if err := ms.fr.onRecv(p, m); err != nil {
			return err
		}
	}
}

func (ms *FormulatorMesh) recvHandshake(conn net.Conn) error {
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

func (ms *FormulatorMesh) sendHandshake(conn net.Conn) (common.PublicHash, error) {
	//log.Println("sendHandshake")
	req := make([]byte, 60)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, err
	}
	copy(req[32:], ms.fr.Config.Formulator[:])
	binary.LittleEndian.PutUint64(req[52:], uint64(time.Now().UnixNano()))
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
