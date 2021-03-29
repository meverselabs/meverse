package pof

import (
	crand "crypto/rand"
	"net"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/debug"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

type ObserverNodeMesh struct {
	sync.Mutex
	ob            *ObserverNode
	key           key.Key
	netAddressMap map[common.PublicHash]string
	clientPeerMap map[string]peer.Peer
	serverPeerMap map[string]peer.Peer
}

func NewObserverNodeMesh(key key.Key, NetAddressMap map[common.PublicHash]string, ob *ObserverNode) *ObserverNodeMesh {
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
							rlog.Println("[client]", err, NetAddr)
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
func (ms *ObserverNodeMesh) SendTo(pubhash common.PublicHash, bs []byte) error {
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
		return ErrNotExistObserverPeer
	}

	p.SendPacket(bs)
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
		return ErrNotExistObserverPeer
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
}

func (ms *ObserverNodeMesh) client(Address string, TargetPubHash common.PublicHash) error {
	d := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.Dial("tcp", Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	start := time.Now()
	if err := ms.recvHandshake(conn); err != nil {
		rlog.Println("[recvHandshake]", err)
		return err
	}
	pubhash, err := ms.sendHandshake(conn)
	if err != nil {
		rlog.Println("[sendHandshake]", err)
		return err
	}
	if pubhash != TargetPubHash {
		return common.ErrInvalidPublicHash
	}
	if _, has := ms.netAddressMap[pubhash]; !has {
		return ErrInvalidObserverKey
	}

	ID := string(pubhash[:])
	p := p2p.NewTCPAsyncPeer(conn, ID, pubhash.String(), start.UnixNano())
	ms.removePeerInMap(ID, ms.clientPeerMap)
	ms.Lock()
	ms.clientPeerMap[ID] = p
	ms.Unlock()
	defer ms.removePeerInMap(p.ID(), ms.clientPeerMap)

	if err := ms.handleConnection(p); err != nil {
		rlog.Println("[handleConnection]", err)
	}
	return nil
}

func (ms *ObserverNodeMesh) server(BindAddress string) error {
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
			pubhash, err := ms.sendHandshake(conn)
			if err != nil {
				rlog.Println("[sendHandshake]", err)
				return
			}
			if _, has := ms.netAddressMap[pubhash]; !has {
				rlog.Println("ErrInvalidPublicHash", pubhash)
				return
			}
			if err := ms.recvHandshake(conn); err != nil {
				rlog.Println("[recvHandshakeAck]", err)
				return
			}

			ID := string(pubhash[:])
			p := p2p.NewTCPAsyncPeer(conn, ID, pubhash.String(), start.UnixNano())
			ms.removePeerInMap(ID, ms.serverPeerMap)
			ms.Lock()
			ms.serverPeerMap[ID] = p
			ms.Unlock()
			defer ms.removePeerInMap(p.ID(), ms.serverPeerMap)

			if err := ms.handleConnection(p); err != nil {
				rlog.Println("[handleConnection]", err)
			}
		}()
	}
}

func (ms *ObserverNodeMesh) handleConnection(p peer.Peer) error {
	if debug.DEBUG {
		rlog.Println("Observer", common.NewPublicHash(ms.key.PublicKey()).String(), "Observer Connected", p.Name())
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
	//rlog.Println("recvHandshake")
	req := make([]byte, 40)
	if _, err := p2p.FillBytes(conn, req); err != nil {
		return err
	}
	ChainID := req[0]
	if ChainID != ms.ob.cs.cn.Provider().ChainID() {
		return chain.ErrInvalidChainID
	}
	timestamp := binutil.LittleEndian.Uint64(req[32:])
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return p2p.ErrInvalidHandshake
	}
	//rlog.Println("sendHandshakeAck")
	if sig, err := ms.key.Sign(hash.Hash(req)); err != nil {
		return err
	} else if _, err := conn.Write(sig[:]); err != nil {
		return err
	}
	return nil
}

func (ms *ObserverNodeMesh) sendHandshake(conn net.Conn) (common.PublicHash, error) {
	//rlog.Println("sendHandshake")
	req := make([]byte, 40)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, err
	}
	req[0] = ms.ob.cs.cn.Provider().ChainID()
	binutil.LittleEndian.PutUint64(req[32:], uint64(time.Now().UnixNano()))
	if _, err := conn.Write(req); err != nil {
		return common.PublicHash{}, err
	}
	//rlog.Println("recvHandshakeAsk")
	var sig common.Signature
	if _, err := p2p.FillBytes(conn, sig[:]); err != nil {
		return common.PublicHash{}, err
	}
	pubkey, err := common.RecoverPubkey(hash.Hash(req), sig)
	if err != nil {
		return common.PublicHash{}, err
	}
	pubhash := common.NewPublicHash(pubkey)
	return pubhash, nil
}
