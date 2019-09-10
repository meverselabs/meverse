package poa

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"net/http"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/debug"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

// ClientService provides connectivity with clients
type ClientService struct {
	sync.Mutex
	key     key.Key
	an      *AuthorityNode
	peerMap map[string]peer.Peer
}

// NewClientService returns a ClientService
func NewClientService(an *AuthorityNode) *ClientService {
	ms := &ClientService{
		key:     an.key,
		an:      an,
		peerMap: map[string]peer.Peer{},
	}
	return ms
}

// Run provides a server
func (ms *ClientService) Run(BindAddress string) {
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

// PeerCount returns a number of the peer
func (ms *ClientService) PeerCount() int {
	ms.Lock()
	defer ms.Unlock()

	return len(ms.peerMap)
}

// RemovePeer removes peers from the mesh
func (ms *ClientService) RemovePeer(ID string) {
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

// SendTo sends a message to the client
func (ms *ClientService) SendTo(pubhash common.PublicHash, m interface{}) error {
	ms.Lock()
	p, has := ms.peerMap[string(pubhash[:])]
	ms.Unlock()
	if !has {
		return ErrNotExistClientPeer
	}

	if err := p.Send(m); err != nil {
		rlog.Println(err)
		ms.RemovePeer(p.ID())
	}
	return nil
}

// BroadcastMessage sends a message to all peers
func (ms *ClientService) BroadcastMessage(m interface{}) error {
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

	peers := []peer.Peer{}
	ms.Lock()
	for _, p := range ms.peerMap {
		peers = append(peers, p)
	}
	ms.Unlock()

	for _, p := range peers {
		p.SendRaw(data)
	}
	return nil
}

func (ms *ClientService) server(BindAddress string) error {
	if debug.DEBUG {
		rlog.Println("ClientService", common.NewPublicHash(ms.key.PublicKey()), "Start to Listen", BindAddress)
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
			return err
		}
		defer conn.Close()

		pubhash, err := ms.sendHandshake(conn)
		if err != nil {
			rlog.Println("[sendHandshake]", err)
			return err
		}
		/*
			if !ms.an.IsAcceptedClient(pubhash) {
				rlog.Println("[IsAcceptedClient]", pubhash.String())
				return err
			}
		*/
		if err := ms.recvHandshake(conn); err != nil {
			rlog.Println("[recvHandshakeAck]", err)
			return err
		}

		ID := string(pubhash[:])
		p := p2p.NewWebsocketPeer(conn, ID, pubhash.String(), time.Now().UnixNano())
		ms.RemovePeer(ID)
		ms.Lock()
		ms.peerMap[ID] = p
		ms.Unlock()
		defer ms.RemovePeer(p.ID())

		if err := ms.handleConnection(p); err != nil {
			rlog.Println("[handleConnection]", err)
			return nil
		}
		return nil
	})
	return e.Start(BindAddress)
}

func (ms *ClientService) handleConnection(p peer.Peer) error {
	if debug.DEBUG {
		rlog.Println("anserver", common.NewPublicHash(ms.key.PublicKey()).String(), "Fromulator Connected", p.Name())
	}

	cp := ms.an.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	p.Send(&p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	})

	for {
		m, _, err := p.ReadMessageData()
		if err != nil {
			return err
		}
		if err := ms.an.OnRecv(p, m); err != nil {
			return err
		}
	}
}

func (ms *ClientService) recvHandshake(conn *websocket.Conn) error {
	//rlog.Println("recvHandshake")
	_, req, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if len(req) != 40 {
		return p2p.ErrInvalidHandshake
	}
	ChainID := req[0]
	if ChainID != ms.an.cs.cn.Provider().ChainID() {
		return chain.ErrInvalidChainID
	}
	timestamp := binary.LittleEndian.Uint64(req[32:])
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
	} else if err := conn.WriteMessage(websocket.BinaryMessage, sig[:]); err != nil {
		return err
	}
	return nil
}

func (ms *ClientService) sendHandshake(conn *websocket.Conn) (common.PublicHash, error) {
	//rlog.Println("sendHandshake")
	req := make([]byte, 40)
	if _, err := crand.Read(req[:32]); err != nil {
		return common.PublicHash{}, err
	}
	req[0] = ms.an.cs.cn.Provider().ChainID()
	binary.LittleEndian.PutUint64(req[32:], uint64(time.Now().UnixNano()))
	if err := conn.WriteMessage(websocket.BinaryMessage, req); err != nil {
		return common.PublicHash{}, err
	}
	//rlog.Println("recvHandshakeAsk")
	_, bs, err := conn.ReadMessage()
	if err != nil {
		return common.PublicHash{}, err
	}
	if len(bs) != common.SignatureSize {
		return common.PublicHash{}, p2p.ErrInvalidHandshake
	}
	var sig common.Signature
	copy(sig[:], bs)
	pubkey, err := common.RecoverPubkey(hash.Hash(req), sig)
	if err != nil {
		return common.PublicHash{}, err
	}
	pubhash := common.NewPublicHash(pubkey)
	return pubhash, nil
}

// clientMap returns a client list as a map
func (ms *ClientService) clientMap() map[common.Address]bool {
	ms.Lock()
	defer ms.Unlock()

	clientMap := map[common.Address]bool{}
	for _, p := range ms.peerMap {
		var addr common.Address
		copy(addr[:], []byte(p.ID()))
		clientMap[addr] = true
	}
	return clientMap
}
