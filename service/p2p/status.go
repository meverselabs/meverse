package p2p

import "github.com/fletaio/fleta/common/hash"

// Status represents the status of the peer
type Status struct {
	Version  uint16
	Height   uint32
	LastHash hash.Hash256
}
