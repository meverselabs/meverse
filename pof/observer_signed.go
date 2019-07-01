package pof

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// ObserverSigned is signatures from observers
type ObserverSigned struct {
	types.BlockSign
	ObserverSignatures []common.Signature
}
