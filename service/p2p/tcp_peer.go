package p2p

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/encoding"
)

// TCPPeer manages send and recv of the connection
type TCPPeer struct {
	sync.Mutex
	conn          net.Conn
	id            string
	name          string
	guessHeight   uint32
	writeQueue    *queue.Queue
	isClose       bool
	connectedTime int64
	pingTime      time.Duration
}

// NewTCPPeer returns a TCPPeer
func NewTCPPeer(conn net.Conn, ID string, Name string, connectedTime int64, pingTime time.Duration) *TCPPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &TCPPeer{
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
			_, err := p.conn.Write(wbs)
			if err != nil {
				return
			}
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *TCPPeer) ID() string {
	return p.id
}

// Name returns the name of the peer
func (p *TCPPeer) Name() string {
	return p.name
}

// Close closes TCPPeer
func (p *TCPPeer) Close() {
	p.isClose = true
	p.conn.Close()
}

// ReadMessageData returns a message data
func (p *TCPPeer) ReadMessageData() (interface{}, []byte, error) {
	var t uint16
	if v, _, err := ReadUint16(p.conn); err != nil {
		return nil, nil, err
	} else {
		t = v
	}

	if Len, _, err := ReadUint32(p.conn); err != nil {
		return nil, nil, err
	} else if Len == 0 {
		return nil, nil, ErrUnknownMessage
	} else {
		zbs := make([]byte, Len)
		if _, err := FillBytes(p.conn, zbs); err != nil {
			return nil, nil, err
		}
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

// Send sends a message to the TCPPeer
func (p *TCPPeer) Send(m interface{}) error {
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

// SendRaw sends bytes to the TCPPeer
func (p *TCPPeer) SendRaw(bs []byte) {
	p.writeQueue.Push(bs)
}

// UpdateGuessHeight updates the guess height of the TCPPeer
func (p *TCPPeer) UpdateGuessHeight(height uint32) {
	p.Lock()
	defer p.Unlock()

	p.guessHeight = height
}

// GuessHeight updates the guess height of the TCPPeer
func (p *TCPPeer) GuessHeight() uint32 {
	return p.guessHeight
}

// ConnectedTime returns peer connected time
func (p *TCPPeer) ConnectedTime() int64 {
	return p.connectedTime
}

// PingTime returns peer ping time
func (p *TCPPeer) PingTime() time.Duration {
	return p.pingTime
}
