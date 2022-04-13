package peer

// Peer manages send and recv of the connection
type Peer interface {
	ID() string
	Name() string
	Close()
	IsClosed() bool
	ReadPacket() ([]byte, error)
	SendPacket(bs []byte)
	ConnectedTime() int64
}
