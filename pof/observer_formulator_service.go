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

// FormulatorService provides connectivity with formulators
type FormulatorService struct {
	sync.Mutex
	key     key.Key
	ob      *Observer
	peerMap map[common.Address]*Peer
}

// NewFormulatorService returns a FormulatorService
func NewFormulatorService(ob *Observer) *FormulatorService {
	ms := &FormulatorService{
		key:     ob.key,
		ob:      ob,
		peerMap: map[common.Address]*Peer{},
	}
	return ms
}

// Run provides a server
func (ms *FormulatorService) Run(BindAddress string) {
	if err := ms.server(BindAddress); err != nil {
		panic(err)
	}
}

// PeerCount returns a number of the peer
func (ms *FormulatorService) PeerCount() int {
	ms.Lock()
	defer ms.Unlock()

	return len(ms.peerMap)
}

// RemovePeer removes peers from the mesh
func (ms *FormulatorService) RemovePeer(addr common.Address) {
	ms.Lock()
	p, has := ms.peerMap[addr]
	if has {
		delete(ms.peerMap, addr)
	}
	ms.Unlock()

	if has {
		p.Close()
	}
}

// SendTo sends a message to the formulator
func (ms *FormulatorService) SendTo(Formulator common.Address, m interface{}) error {
	ms.Lock()
	p, has := ms.peerMap[Formulator]
	ms.Unlock()
	if !has {
		return ErrNotExistFormulatorPeer
	}

	if err := p.Send(m); err != nil {
		log.Println(err)
		ms.RemovePeer(p.address)
	}
	return nil
}

// BroadcastMessage sends a message to all peers
func (ms *FormulatorService) BroadcastMessage(m interface{}) error {
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

	peers := []*Peer{}
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

// GuessHeightCountMap returns a number of the guess height from all peers
func (ms *FormulatorService) GuessHeightCountMap() map[uint32]int {
	CountMap := map[uint32]int{}
	ms.Lock()
	for _, p := range ms.peerMap {
		CountMap[p.GuessHeight()]++
	}
	ms.Unlock()
	return CountMap
}

// GuessHeight returns the guess height of the fomrulator
func (ms *FormulatorService) GuessHeight(Formulator common.Address) (uint32, error) {
	ms.Lock()
	p, has := ms.peerMap[Formulator]
	ms.Unlock()
	if !has {
		return 0, ErrNotExistFormulatorPeer
	}
	return p.GuessHeight(), nil
}

// UpdateGuessHeight updates the guess height of the fomrulator
func (ms *FormulatorService) UpdateGuessHeight(Formulator common.Address, height uint32) {
	ms.Lock()
	p, has := ms.peerMap[Formulator]
	ms.Unlock()
	if has {
		p.UpdateGuessHeight(height)
	}
}

func (ms *FormulatorService) server(BindAddress string) error {
	lstn, err := net.Listen("tcp", BindAddress)
	if err != nil {
		return err
	}
	log.Println("FormulatorService", common.NewPublicHash(ms.key.PublicKey()), "Start to Listen", BindAddress)

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
			Formulator, err := ms.recvHandshake(conn)
			if err != nil {
				log.Println("[recvHandshakeAck]", err)
				return
			}
			if !ms.ob.cs.rt.IsFormulator(Formulator, pubhash) {
				log.Println("[IsFormulator]", Formulator.String(), pubhash.String())
				return
			}

			p := NewPeer(conn)
			p.pubhash = pubhash
			p.address = Formulator
			ms.Lock()
			old, has := ms.peerMap[Formulator]
			ms.peerMap[Formulator] = p
			ms.Unlock()
			if has {
				ms.RemovePeer(old.address)
			}
			defer ms.RemovePeer(p.address)

			if err := ms.handleConnection(p); err != nil {
				log.Println("[handleConnection]", err)
			}
		}()
	}
}

func (ms *FormulatorService) handleConnection(p *Peer) error {
	log.Println("Observer", common.NewPublicHash(ms.key.PublicKey()).String(), "Fromulator Connected", p.address.String())

	cp := ms.ob.cs.cn.Provider()
	p.Send(&p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   cp.Height(),
		LastHash: cp.LastHash(),
	})

	var pingCount uint64
	pingCountLimit := uint64(3)
	pingTimer := time.NewTimer(10 * time.Second)
	go func() {
		for {
			select {
			case <-pingTimer.C:
				if err := p.Send(&p2p.PingMessage{}); err != nil {
					p.Close()
					return
				}
				if atomic.AddUint64(&pingCount, 1) > pingCountLimit {
					p.Close()
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
		if _, is := m.(*p2p.PingMessage); is {
			continue
		}

		if err := ms.ob.onFormulatorRecv(p, m, bs); err != nil {
			return err
		}
	}
}

func (ms *FormulatorService) recvHandshake(conn net.Conn) (common.Address, error) {
	//log.Println("recvHandshake")
	req := make([]byte, 60)
	if _, err := p2p.FillBytes(conn, req); err != nil {
		return common.Address{}, err
	}
	var Formulator common.Address
	copy(Formulator[:], req[32:])
	timestamp := binary.LittleEndian.Uint64(req[52:])
	diff := time.Duration(uint64(time.Now().UnixNano()) - timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second*30 {
		return common.Address{}, p2p.ErrInvalidHandshake
	}
	//log.Println("sendHandshakeAck")
	h := hash.Hash(req)
	if sig, err := ms.key.Sign(h); err != nil {
		return common.Address{}, err
	} else if _, err := conn.Write(sig[:]); err != nil {
		return common.Address{}, err
	}
	return Formulator, nil
}

func (ms *FormulatorService) sendHandshake(conn net.Conn) (common.PublicHash, error) {
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

// FormulatorMap returns a formulator list as a map
func (ms *FormulatorService) FormulatorMap() map[common.Address]bool {
	ms.Lock()
	defer ms.Unlock()

	FormulatorMap := map[common.Address]bool{}
	for _, p := range ms.peerMap {
		FormulatorMap[p.address] = true
	}
	return FormulatorMap
}
