package types

import (
	"github.com/fletaio/fleta/common/hash"
)

// Header is validation informations
type Header struct {
	Version       uint16
	Height        uint32
	PrevHash      hash.Hash256
	LevelRootHash hash.Hash256
	ContextHash   hash.Hash256
	Timestamp     uint64
	ConsensusData []byte
}
