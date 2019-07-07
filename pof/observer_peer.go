package pof

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"
	"net"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
)

// Peer is a observer peer
type Peer struct {
	id         string
	netAddr    string
	conn       net.Conn
	pubhash    common.PublicHash
	startTime  uint64
	readTotal  uint64
	writeTotal uint64
	writeChan  chan []byte
}

// NewPeer returns a Peer
func NewPeer(conn net.Conn, pubhash common.PublicHash) *Peer {
	p := &Peer{
		id:        pubhash.String(),
		netAddr:   conn.RemoteAddr().String(),
		conn:      conn,
		pubhash:   pubhash,
		startTime: uint64(time.Now().UnixNano()),
		writeChan: make(chan []byte, 100),
	}
	go func() {
		defer p.conn.Close()

		for {
			select {
			case bs := <-p.writeChan:
				var buffer bytes.Buffer
				buffer.Write(bs[:2])          // message type
				buffer.Write(make([]byte, 4)) //size of gzip
				if len(bs) > 2 {
					zw := gzip.NewWriter(&buffer)
					zw.Write(bs[8:])
					zw.Flush()
					zw.Close()
				}
				wbs := buffer.Bytes()
				binary.LittleEndian.PutUint32(wbs[2:], uint32(len(wbs)-12))
				if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				_, err := p.conn.Write(wbs)
				if err != nil {
					return
				}
				atomic.AddUint64(&p.writeTotal, uint64(len(wbs)))
			}
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *Peer) ID() string {
	return p.id
}

// NetAddr returns the network address of the peer
func (p *Peer) NetAddr() string {
	return p.netAddr
}

// ReadMessageData returns a message data
func (p *Peer) ReadMessageData() (interface{}, []byte, error) {
	var t uint16
	if v, _, err := p2p.ReadUint16(p.conn); err != nil {
		return nil, nil, err
	} else {
		t = v
	}

	fc := encoding.Factory("pof.message")
	if Len, _, err := p2p.ReadUint32(p.conn); err != nil {
		return nil, nil, err
	} else if Len == 0 {
		return nil, nil, nil
	} else {
		zbs := make([]byte, Len)
		if _, err := p2p.FillBytes(p.conn, zbs); err != nil {
			return nil, nil, err
		}
		zr, err := gzip.NewReader(bytes.NewReader(zbs))
		if err != nil {
			return nil, nil, err
		}
		defer zr.Close()

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

// Send sends a message to the peer
func (p *Peer) Send(m interface{}) error {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	fc := encoding.Factory("pof.message")
	t, err := fc.TypeOf(m)
	if err != nil {
		return err
	}
	if err := enc.EncodeUint16(t); err != nil {
		return err
	}
	if err := enc.Encode(m); err != nil {
		return err
	}
	return p.SendRaw(buffer.Bytes())
}

// SendRaw sends bytes to the peer
func (p *Peer) SendRaw(bs []byte) error {
	p.writeChan <- bs
	return nil
}
