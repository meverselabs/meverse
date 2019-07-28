package types

import "encoding/json"

// Event defines common event functions
type Event interface {
	json.Marshaler
	Height() uint32
	Index() uint16
	N() uint16
	SetN(n uint16)
}
