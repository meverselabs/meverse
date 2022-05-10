package p2p

import (
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/queue"
)

// TCPAsyncPeer manages send and recv of the connection
type TCPAsyncPeer struct {
	conn          net.Conn
	id            string
	name          string
	isClose       bool
	connectedTime int64
	pingCount     uint64
	pingType      uint32
	writeQ        *queue.Queue
}

// NewTCPAsyncPeer returns a TCPAsyncPeer
func NewTCPAsyncPeer(conn net.Conn, ID string, Name string, connectedTime int64) *TCPAsyncPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &TCPAsyncPeer{
		conn:          conn,
		id:            ID,
		name:          Name,
		connectedTime: connectedTime,
		writeQ:        queue.NewQueue(),
		pingType:      0x12345678,
	}

	go func() {
		defer p.Close()

		pingCountLimit := uint64(3)
		for !p.isClose {
			if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return
			}
			_, err := p.conn.Write(bin.Uint32Bytes(p.pingType))
			if err != nil {
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
					log.Printf("%v SendPacket.SetWriteDeadline %+v\n", p.name, err)
					p.Close()
					return
				}
				if _, err := p.conn.Write(bs); err != nil {
					log.Printf("%v SendPacket.Write %+v\n", p.name, err)
					p.Close()
					return
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *TCPAsyncPeer) ID() string {
	return p.id
}

// Name returns the name of the peer
func (p *TCPAsyncPeer) Name() string {
	return p.name
}

// Close closes TCPAsyncPeer
func (p *TCPAsyncPeer) Close() {
	p.isClose = true
	p.conn.Close()
}

// IsClosed returns it is closed or not
func (p *TCPAsyncPeer) IsClosed() bool {
	return p.isClose
}

// ReadPacket returns a packet data
func (p *TCPAsyncPeer) ReadPacket() ([]byte, error) {
	for {
		if t, _, err := ReadUint32(p.conn); err != nil {
			return nil, err
		} else {
			atomic.StoreUint64(&p.pingCount, 0)
			if t == p.pingType {
				continue
			} else {
				if Len, _, err := ReadUint32(p.conn); err != nil {
					return nil, err
				} else {
					bs := make([]byte, 8+Len)
					bin.PutUint32(bs, t)
					bin.PutUint32(bs[4:], Len)
					if _, err := FillBytes(p.conn, bs[8:]); err != nil {
						return nil, err
					}
					return bs, nil
				}
			}
		}
	}
}

// SendPacket sends packet to the WebsocketPeer
func (p *TCPAsyncPeer) SendPacket(bs []byte) {
	p.writeQ.Push(bs)
}

// ConnectedTime returns peer connected time
func (p *TCPAsyncPeer) ConnectedTime() int64 {
	return p.connectedTime
}
