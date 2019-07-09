package types

// Event defines common event functions
type Event interface {
	Height() uint32
	Index() uint16
	N() uint16
	SetN(n uint16)
}
