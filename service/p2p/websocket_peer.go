package p2p

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"
	"sync"
	"time"

	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/encoding"
	"github.com/gorilla/websocket"
)

// WebsocketPeer manages send and recv of the connection
type WebsocketPeer struct {
	sync.Mutex
	conn          *websocket.Conn
	id            string
	name          string
	guessHeight   uint32
	writeQueue    *queue.Queue
	isClose       bool
	connectedTime int64
	pingTime      time.Duration
}

// NewWebsocketPeer returns a WebsocketPeer
func NewWebsocketPeer(conn *websocket.Conn, ID string, Name string, connectedTime int64, pingTime time.Duration) *WebsocketPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &WebsocketPeer{
		conn:          conn,
		id:            ID,
		name:          Name,
		writeQueue:    queue.NewQueue(),
		connectedTime: connectedTime,
		pingTime:      pingTime,
	}
	go func() {
		defer p.conn.Close()

		for {
			if p.isClose {
				return
			}
			v := p.writeQueue.Pop()
			if v == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			bs := v.([]byte)
			var buffer bytes.Buffer
			buffer.Write(bs[:2])
			buffer.Write(make([]byte, 4))
			if len(bs) > 2 {
				zw := gzip.NewWriter(&buffer)
				zw.Write(bs[2:])
				zw.Flush()
				zw.Close()
			}
			wbs := buffer.Bytes()
			binary.LittleEndian.PutUint32(wbs[2:], uint32(len(wbs)-6))
			if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return
			}
			if err := p.conn.WriteMessage(websocket.BinaryMessage, wbs); err != nil {
				return
			}
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
	p.conn.Close()
	p.isClose = true
}

// ReadMessageData returns a message data
func (p *WebsocketPeer) ReadMessageData() (interface{}, []byte, error) {
	_, bs, err := p.conn.ReadMessage()
	if err != nil {
		return nil, nil, err
	}
	if len(bs) < 6 {
		return nil, nil, ErrInvalidLength
	}

	t := util.BytesToUint16(bs)
	Len := util.BytesToUint32(bs[2:])
	if Len == 0 {
		return nil, nil, ErrUnknownMessage
	} else if len(bs) != 6+int(Len) {
		return nil, nil, ErrInvalidLength
	} else {
		zbs := bs[6:]
		zr, err := gzip.NewReader(bytes.NewReader(zbs))
		if err != nil {
			return nil, nil, err
		}
		defer zr.Close()

		fc := encoding.Factory("message")
		m, err := fc.Create(t)
		if err != nil {
			return nil, nil, err
		}
		bs, err := ioutil.ReadAll(zr)
		if err != nil {
			return nil, nil, err
		}
		if err := encoding.Unmarshal(bs, &m); err != nil {
			return nil, nil, err
		}
		return m, bs, nil
	}
}

// Send sends a message to the WebsocketPeer
func (p *WebsocketPeer) Send(m interface{}) error {
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
	p.SendRaw(buffer.Bytes())
	return nil
}

// SendRaw sends bytes to the WebsocketPeer
func (p *WebsocketPeer) SendRaw(bs []byte) {
	p.writeQueue.Push(bs)
}

// UpdateGuessHeight updates the guess height of the WebsocketPeer
func (p *WebsocketPeer) UpdateGuessHeight(height uint32) {
	p.Lock()
	defer p.Unlock()

	p.guessHeight = height
}

// GuessHeight updates the guess height of the WebsocketPeer
func (p *WebsocketPeer) GuessHeight() uint32 {
	return p.guessHeight
}

// ConnectedTime returns peer connected time
func (p *WebsocketPeer) ConnectedTime() int64 {
	return p.connectedTime
}

// PingTime returns peer ping time
func (p *WebsocketPeer) PingTime() time.Duration {
	return p.pingTime
}
