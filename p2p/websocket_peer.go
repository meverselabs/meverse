package p2p

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// WebsocketPeer manages send and recv of the connection
type WebsocketPeer struct {
	sync.Mutex
	conn          *websocket.Conn
	id            string
	name          string
	isClose       bool
	connectedTime int64
	pingCount     uint64
}

// NewWebsocketPeer returns a WebsocketPeer
func NewWebsocketPeer(conn *websocket.Conn, ID string, Name string, connectedTime int64) *WebsocketPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &WebsocketPeer{
		conn:          conn,
		id:            ID,
		name:          Name,
		connectedTime: connectedTime,
	}
	conn.EnableWriteCompression(false)
	conn.SetPingHandler(func(appData string) error {
		atomic.StoreUint64(&p.pingCount, 0)
		return nil
	})

	go func() {
		defer p.Close()

		pingCountLimit := uint64(3)
		for !p.isClose {
			p.Lock()
			if err := p.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				p.Unlock()
				return
			}
			p.Unlock()
			if atomic.AddUint64(&p.pingCount, 1) > pingCountLimit {
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *WebsocketPeer) ID() string {
	return p.id
}

// Name returns the name of the peer
func (p *WebsocketPeer) Name() string {
	return p.name
}

// Close closes WebsocketPeer
func (p *WebsocketPeer) Close() {
	p.isClose = true
	p.conn.Close()
}

// IsClosed returns it is closed or not
func (p *WebsocketPeer) IsClosed() bool {
	return p.isClose
}

// ReadPacket returns a packet data
func (p *WebsocketPeer) ReadPacket() ([]byte, error) {
	_, rb, err := p.conn.ReadMessage()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rb, nil
}

// SendPacket sends packet to the WebsocketPeer
func (p *WebsocketPeer) SendPacket(bs []byte) {
	if p.isClose {
		return
	}

	p.Lock()
	defer p.Unlock()

	if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("%v SendPacket %+v\n", p.name, err)
		p.Close()
		return
	}
	if err := p.conn.WriteMessage(websocket.BinaryMessage, bs); err != nil {
		log.Printf("%v SendPacket %+v\n", p.name, err)
		p.Close()
		return
	}
}

// ConnectedTime returns peer connected time
func (p *WebsocketPeer) ConnectedTime() int64 {
	return p.connectedTime
}
