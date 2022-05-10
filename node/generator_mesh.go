package node

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

type GeneratorNodeMesh struct {
	sync.Mutex
	fr            *GeneratorNode
	key           key.Key
	netAddressMap map[common.PublicKey]string
	peerMap       map[string]peer.Peer
}

func NewGeneratorNodeMesh(key key.Key, NetAddressMap map[common.PublicKey]string, fr *GeneratorNode) *GeneratorNodeMesh {
	ms := &GeneratorNodeMesh{
		key:           key,
		netAddressMap: NetAddressMap,
		peerMap:       map[string]peer.Peer{},
		fr:            fr,
	}
	return ms
}

// Run starts the generator mesh
func (ms *GeneratorNodeMesh) Run() {
	myPubKey := ms.key.PublicKey()

	for PubKey, v := range ms.netAddressMap {
		if PubKey != myPubKey {
			go func(pubkey common.PublicKey, NetAddr string) {
				time.Sleep(1 * time.Second)
				for {
					ctx := ms.fr.cn.NewContext()
					if ctx.IsGenerator(myPubKey.Address()) {
						ms.Lock()
						_, has := ms.peerMap[string(pubkey[:])]
						ms.Unlock()
						if !has {
							if err := ms.client(NetAddr, pubkey); err != nil {
								log.Printf("[client] %+v %v\n", err, NetAddr)
							}
						}
					}
					time.Sleep(10 * time.Second)
				}
			}(PubKey, v)
		}
	}
}

// Peers returns peers of the generator mesh
func (ms *GeneratorNodeMesh) Peers() []peer.Peer {
	ms.Lock()
	defer ms.Unlock()

	peers := []peer.Peer{}
	for _, p := range ms.peerMap {
		peers = append(peers, p)
	}
	return peers
}

// RemovePeer removes peers from the mesh
func (ms *GeneratorNodeMesh) RemovePeer(ID string) {
	ms.Lock()
	p, has := ms.peerMap[ID]
	if has {
		delete(ms.peerMap, ID)
	}
	ms.Unlock()

	if has {
		p.Close()
	}
}

// SendTo sends a message to the observer
func (ms *GeneratorNodeMesh) SendTo(ID string, m p2p.Serializable) error {
	ms.Lock()
	p, has := ms.peerMap[ID]
	ms.Unlock()
	if !has {
		return errors.WithStack(ErrNotExistObserverPeer)
	}

	p.SendPacket(p2p.MessageToPacket(m))
	return nil
}

// BroadcastPacket sends a packet to all peers
func (ms *GeneratorNodeMesh) BroadcastPacket(bs []byte) {
	peerMap := map[string]peer.Peer{}
	ms.Lock()
	for _, p := range ms.peerMap {
		peerMap[p.ID()] = p
	}
	ms.Unlock()

	for _, p := range peerMap {
		p.SendPacket(bs)
	}
}

func (ms *GeneratorNodeMesh) client(Address string, TargetPubKey common.PublicKey) error {
	d := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}
	conn, _, err := d.Dial(Address, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	defer conn.Close()

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
	p := p2p.NewWebsocketPeer(conn, ID, pubkey.String(), time.Now().UnixNano())
	ms.RemovePeer(ID)
	ms.Lock()
	ms.peerMap[ID] = p
	ms.Unlock()
	defer ms.RemovePeer(p.ID())

	if err := ms.handleConnection(p); err != nil {
		log.Printf("[handleConnection] %+v\n", err)
	}
	return nil
}

func (ms *GeneratorNodeMesh) handleConnection(p peer.Peer) error {
	if DEBUG {
		log.Println("Generatorlog", ms.key.PublicKey().Address().String(), "Observer Connected", p.Name())
	}

	ms.fr.OnObserverConnected(p)
	defer ms.fr.OnObserverDisconnected(p)

	for {
		bs, err := p.ReadPacket()
		if err != nil {
			return err
		}
		if err := ms.fr.onObserverRecv(p, bs); err != nil {
			return err
		}
	}
}

func (ms *GeneratorNodeMesh) recvHandshake(conn *websocket.Conn) error {
	//log.Println("recvHandshake")
	_, req, err := conn.ReadMessage()
	if err != nil {
		return errors.WithStack(err)
	}
	ChainID, _, timestamp, err := p2p.RecoveryHandSharkBs(req)
	if err != nil {
		return err
	}
	if ChainID.Cmp(ms.fr.ChainID) != 0 {
		log.Println("not match chainin from :", ChainID.String(), "has:", ms.fr.ChainID.String())
		return errors.WithStack(chain.ErrInvalidChainID)
	}
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return errors.WithStack(p2p.ErrInvalidHandshake)
	}
	//log.Println("sendHandshakeAck")
	if sig, err := ms.key.Sign(hash.Hash(req)); err != nil {
		return err
	} else if err := conn.WriteMessage(websocket.BinaryMessage, sig[:]); err != nil {
		return err
	}
	return nil
}

func (ms *GeneratorNodeMesh) sendHandshake(conn *websocket.Conn) (common.PublicKey, error) {
	//log.Println("sendHandshake")

	rn := rand.Uint64()
	req, err := p2p.MakeHandSharkBs(ms.fr.ChainID, rn, uint64(time.Now().UnixNano()))
	if err != nil {
		return common.PublicKey{}, err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, req); err != nil {
		return common.PublicKey{}, err
	}
	//log.Println("recvHandshakeAsk")
	_, sig, err := conn.ReadMessage()
	if err != nil {
		return common.PublicKey{}, err
	}
	pubkey, err := common.RecoverPubkey(ms.fr.ChainID, hash.Hash(req), sig)
	if err != nil {
		return common.PublicKey{}, err
	}
	return pubkey, nil
}
