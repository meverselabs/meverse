package p2p

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common/queue"
	"github.com/gorilla/websocket"
)

// WebsocketPeer manages send and recv of the connection
type WebsocketPeer struct {
	conn          *websocket.Conn
	id            string
	name          string
	isClose       bool
	connectedTime int64
	pingCount     uint64
	writeQ        *queue.Queue
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
		writeQ:        queue.NewQueue(),
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
			if err := p.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
			if atomic.AddUint64(&p.pingCount, 1) > pingCountLimit {
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		defer p.Close()
		for !p.isClose {
			for !p.isClose {
				v := p.writeQ.Pop()
				if v == nil {
					break
				}
				bs := v.([]byte)
				if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
					log.Println(p.name, "SendPacket", err)
					p.Close()
					return
				}
				if err := p.conn.WriteMessage(websocket.BinaryMessage, bs); err != nil {
					log.Println(p.name, "SendPacket", err)
					p.Close()
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
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
		return nil, err
	}
	return rb, nil
}

// SendPacket sends packet to the WebsocketPeer
func (p *WebsocketPeer) SendPacket(bs []byte) {
	p.writeQ.Push(bs)
}

// ConnectedTime returns peer connected time
func (p *WebsocketPeer) ConnectedTime() int64 {
	return p.connectedTime
}
