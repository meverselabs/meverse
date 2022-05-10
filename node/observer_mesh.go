package node

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

type ObserverNodeMesh struct {
	sync.Mutex
	ob            *ObserverNode
	key           key.Key
	netAddressMap map[common.PublicKey]string
	clientPeerMap map[string]peer.Peer
	serverPeerMap map[string]peer.Peer
}

func NewObserverNodeMesh(key key.Key, NetAddressMap map[common.PublicKey]string, ob *ObserverNode) *ObserverNodeMesh {
	ms := &ObserverNodeMesh{
		key:           key,
		netAddressMap: NetAddressMap,
		clientPeerMap: map[string]peer.Peer{},
		serverPeerMap: map[string]peer.Peer{},
		ob:            ob,
	}
	return ms
}

// Run starts the observer mesh
func (ms *ObserverNodeMesh) Run(BindAddress string) {
	myPublicKey := ms.key.PublicKey()
	for PubKey, v := range ms.netAddressMap {
		if PubKey != myPublicKey {
			go func(pubkey common.PublicKey, NetAddr string) {
				time.Sleep(1 * time.Second)
				for {
					ID := string(pubkey[:])
					ms.Lock()
					_, hasC := ms.clientPeerMap[ID]
					_, hasS := ms.serverPeerMap[ID]
					ms.Unlock()
					if !hasC && !hasS {
						if err := ms.client(NetAddr, pubkey); err != nil {
							log.Printf("[client] %+v %v\n ", err, NetAddr)
						}
					}
					time.Sleep(1 * time.Second)
				}
			}(PubKey, v)
		}
	}
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

// Peers returns peers of the observer mesh
func (ms *ObserverNodeMesh) Peers() []peer.Peer {
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

func (ms *ObserverNodeMesh) removePeerInMap(ID string, peerMap map[string]peer.Peer) {
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
func (ms *ObserverNodeMesh) SendTo(PubKey common.PublicKey, bs []byte) error {
	ID := string(PubKey[:])

	ms.Lock()
	var p peer.Peer
	if cp, has := ms.clientPeerMap[ID]; has {
		p = cp
	} else if sp, has := ms.serverPeerMap[ID]; has {
		p = sp
	}
	ms.Unlock()
	if p == nil {
		return errors.WithStack(ErrNotExistObserverPeer)
	}

	p.SendPacket(bs)
	bs = nil
	return nil
}

// SendAnyone sends a message to the one of observers
func (ms *ObserverNodeMesh) SendAnyone(bs []byte) error {
	ms.Lock()
	var p peer.Peer
	for _, v := range ms.clientPeerMap {
		p = v
		break
	}
	if p == nil {
		for _, v := range ms.serverPeerMap {
			p = v
			break
		}
	}
	ms.Unlock()
	if p == nil {
		return errors.WithStack(ErrNotExistObserverPeer)
	}

	p.SendPacket(bs)
	return nil
}

// BroadcastPacket sends a packet to all peers
func (ms *ObserverNodeMesh) BroadcastPacket(bs []byte) {
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
	peerMap = nil
	bs = nil
}

func (ms *ObserverNodeMesh) client(Address string, TargetPubKey common.PublicKey) error {
	d := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.Dial("tcp", Address)
	if err != nil {
		return errors.WithStack(err)
	}
	defer conn.Close()

	start := time.Now()
	if err := ms.recvHandshake(conn); err != nil {
		log.Printf("[recvHandshake] %+v\n", err)
		return err
	}
	pubkey, err := ms.sendHandshake(conn)
	if err != nil {
		log.Printf("[sendHandshake] %+v\n", err)
		return err
	}
	if pubkey != TargetPubKey {
		return errors.WithStack(common.ErrInvalidPublicKey)
	}
	if _, has := ms.netAddressMap[pubkey]; !has {
		return errors.WithStack(ErrInvalidObserverKey)
	}

	ID := string(pubkey[:])
	p := p2p.NewTCPAsyncPeer(conn, ID, pubkey.String(), start.UnixNano())
	ms.removePeerInMap(ID, ms.clientPeerMap)
	ms.Lock()
	ms.clientPeerMap[ID] = p
	ms.Unlock()
	defer ms.removePeerInMap(p.ID(), ms.clientPeerMap)

	if err := ms.handleConnection(p); err != nil {
		log.Printf("[handleConnection client] %+v\n", err)
	}
	return nil
}

func (ms *ObserverNodeMesh) server(BindAddress string) error {
	lstn, err := net.Listen("tcp", BindAddress)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Println(ms.ob.obID, ms.key.PublicKey(), "Start to Listen", BindAddress)
	for {
		conn, err := lstn.Accept()
		if err != nil {
			return errors.WithStack(err)
		}
		go func() {
			defer conn.Close()

			start := time.Now()
			PubKey, err := ms.sendHandshake(conn)
			if err != nil {
				log.Printf("[sendHandshake] %+v\n", err)
				return
			}
			if _, has := ms.netAddressMap[PubKey]; !has {
				log.Println(ms.ob.obID, "ErrInvalidPublicKey", PubKey)
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				log.Printf("[recvHandshakeAck] %+v\n", err)
				return
			}

			ID := string(PubKey[:])
			p := p2p.NewTCPAsyncPeer(conn, ID, PubKey.String(), start.UnixNano())
			ms.removePeerInMap(ID, ms.serverPeerMap)
			ms.Lock()
			ms.serverPeerMap[ID] = p
			ms.Unlock()
			defer ms.removePeerInMap(p.ID(), ms.serverPeerMap)

			if err := ms.handleConnection(p); err != nil {
				log.Printf("[handleConnection] server %+v\n", err)
			}
		}()
	}
}

func (ms *ObserverNodeMesh) handleConnection(p peer.Peer) error {
	if DEBUG {
		log.Println(ms.ob.obID, "Observer", ms.key.PublicKey().String(), "Observer Connected", p.Name())
	}

	for {
		bs, err := p.ReadPacket()
		if err != nil {
			return err
		}
		if err := ms.ob.onObserverRecv(p, bs); err != nil {
			return err
		}
	}
}

func (ms *ObserverNodeMesh) recvHandshake(conn net.Conn) error {
	//log.Println(ms.ob.obID, "recvHandshake")
	req, _, err := bin.ReadBytes(conn)
	if err != nil {
		return err
	}
	ChainID, _, timestamp, err := p2p.RecoveryHandSharkBs(req)
	if err != nil {
		return err
	}
	if ChainID.Cmp(ms.ob.ChainID) != 0 {
		log.Println(ms.ob.obID, "not match chainin from :", ChainID.String(), "has:", ms.ob.ChainID.String())
		return errors.WithStack(chain.ErrInvalidChainID)
	}
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return errors.WithStack(p2p.ErrInvalidHandshake)
	}
	//log.Println(ms.ob.obID, "sendHandshakeAck")
	if sig, err := ms.key.Sign(hash.Hash(req)); err != nil {
		return err
	} else if _, err := bin.WriteBytes(conn, sig[:]); err != nil {
		return err
	}
	return nil
}

func (ms *ObserverNodeMesh) sendHandshake(conn net.Conn) (common.PublicKey, error) {
	//log.Println(ms.ob.obID, "sendHandshake")
	rn := rand.Uint64()
	req, err := p2p.MakeHandSharkBs(ms.ob.ChainID, rn, uint64(time.Now().UnixNano()))
	if err != nil {
		return common.PublicKey{}, err
	}
	_, err = bin.WriteBytes(conn, req)
	if err != nil {
		return common.PublicKey{}, err
	}

	//log.Println(ms.ob.obID, "recvHandshakeAsk")
	sig, _, err := bin.ReadBytes(conn)
	if err != nil {
		return common.PublicKey{}, err
	}
	pubkey, err := common.RecoverPubkey(ms.ob.ChainID, hash.Hash(req), sig)
	if err != nil {
		return common.PublicKey{}, err
	}
	return pubkey, nil
}
