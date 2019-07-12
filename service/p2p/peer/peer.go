package peer

import (
	"time"
)

// Peer manages send and recv of the connection
type Peer interface {
	ID() string
	Name() string
	Close()
	ReadMessageData() (interface{}, []byte, error)
	Send(m interface{}) error
	SendRaw(bs []byte)
	UpdateGuessHeight(height uint32)
	GuessHeight() uint32
	ConnectedTime() int64
	PingTime() time.Duration
}
