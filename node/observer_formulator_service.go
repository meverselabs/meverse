package node

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/p2p/peer"
	"github.com/pkg/errors"
)

// GeneratorService provides connectivity with generators
type GeneratorService struct {
	sync.Mutex
	key     key.Key
	ob      *ObserverNode
	peerMap map[string]peer.Peer
}

// NewGeneratorService returns a GeneratorService
func NewGeneratorService(ob *ObserverNode) *GeneratorService {
	ms := &GeneratorService{
		key:     ob.key,
		ob:      ob,
		peerMap: map[string]peer.Peer{},
	}
	return ms
}

// Run provides a server
func (ms *GeneratorService) Run(BindAddress string) {
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

// PeerCount returns a number of the peer
func (ms *GeneratorService) PeerCount() int {
	ms.Lock()
	defer ms.Unlock()

	return len(ms.peerMap)
}

// RemovePeer removes peers from the mesh
func (ms *GeneratorService) RemovePeer(ID string) {
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

// Peer returns the peer
func (ms *GeneratorService) Peer(ID string) (peer.Peer, bool) {
	ms.Lock()
	p, has := ms.peerMap[ID]
	ms.Unlock()

	return p, has
}

// Peers returns peer list
func (ms *GeneratorService) Peers() []peer.Peer {
	peers := []peer.Peer{}
	ms.Lock()
	for _, p := range ms.peerMap {
		peers = append(peers, p)
	}
	ms.Unlock()

	return peers
}

// SendTo sends a message to the generator
func (ms *GeneratorService) SendTo(addr common.Address, bs []byte) error {
	ms.Lock()
	p, has := ms.peerMap[string(addr[:])]
	ms.Unlock()
	if !has {
		return errors.WithStack(ErrNotExistGeneratorPeer)
	}

	p.SendPacket(bs)
	return nil
}

func (ms *GeneratorService) server(BindAddress string) error {
	if DEBUG {
		log.Println("GeneratorService", ms.key.PublicKey().String(), "Start to Listen", BindAddress)
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return errors.WithStack(err)
		}
		defer conn.Close()

		Generator, err := ms.sendHandshake(conn)
		if err != nil {
			log.Printf("[sendHandshake] %+v\n", err)
			return err
		}
		if err := ms.recvHandshake(conn); err != nil {
			log.Printf("[recvHandshakeAck] %+v\n", err)
			return err
		}
		ctx := ms.ob.cn.NewContext()
		if !ctx.IsGenerator(Generator) {
			log.Printf("[IsGenerator] %+v\n", Generator.String())
			return err
		}

		ID := string(Generator[:])
		p := p2p.NewWebsocketPeer(conn, ID, Generator.String(), time.Now().UnixNano())
		ms.RemovePeer(ID)
		ms.Lock()
		ms.peerMap[ID] = p
		ms.Unlock()
		defer ms.RemovePeer(p.ID())

		if err := ms.handleConnection(p); err != nil {
			log.Printf("[handleConnection] %+v\n", err)
			return nil
		}
		return nil
	})
	return e.Start(BindAddress)
}

func (ms *GeneratorService) handleConnection(p peer.Peer) error {
	if DEBUG {
		log.Println("Observer", ms.key.PublicKey().String(), "Fromulator Connected", p.Name())
	}

	ms.ob.OnGeneratorConnected(p)
	defer ms.ob.OnGeneratorDisconnected(p)

	for {
		bs, err := p.ReadPacket()
		if err != nil {
			return err
		}
		if err := ms.ob.onGeneratorRecv(p, bs); err != nil {
			return err
		}
	}
}

func (ms *GeneratorService) recvHandshake(conn *websocket.Conn) error {
	//log.Println("recvHandshake")
	_, req, err := conn.ReadMessage()
	if err != nil {
		return errors.WithStack(err)
	}
	ChainID, _, timestamp, err := p2p.RecoveryHandSharkBs(req)
	if err != nil {
		return err
	}
	if ChainID.Cmp(ms.ob.ChainID) != 0 {
		log.Println("not match chainin from :", ChainID.String(), "has:", ms.ob.ChainID.String())
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
		return errors.WithStack(err)
	}
	return nil
}

func (ms *GeneratorService) sendHandshake(conn *websocket.Conn) (common.Address, error) {
	//log.Println("sendHandshake")
	rn := rand.Uint64()
	req, err := p2p.MakeHandSharkBs(ms.ob.ChainID, rn, uint64(time.Now().UnixNano()))
	if err != nil {
		return common.Address{}, err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, req); err != nil {
		return common.Address{}, errors.WithStack(err)
	}
	//log.Println("recvHandshakeAsk")
	_, sig, err := conn.ReadMessage()
	if err != nil {
		return common.Address{}, errors.WithStack(err)
	}
	pubkey, err := common.RecoverPubkey(ms.ob.ChainID, hash.Hash(req), sig)
	if err != nil {
		return common.Address{}, err
	}
	return pubkey.Address(), nil
}

// GeneratorMap returns a Generator list as a map
func (ms *GeneratorService) GeneratorMap() map[common.Address]bool {
	ms.Lock()
	defer ms.Unlock()

	GeneratorMap := map[common.Address]bool{}
	for _, p := range ms.peerMap {
		var addr common.Address
		copy(addr[:], []byte(p.ID()))
		GeneratorMap[addr] = true
	}
	return GeneratorMap
}
