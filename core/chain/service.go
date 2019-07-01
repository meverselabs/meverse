package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Service is a interface of the chain service
type Service interface {
	Name() string
	Init(cn *Chain) error
	OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) error
}

// ServiceBase is a base handler of the chain service
type ServiceBase struct{}

// OnBlockConnected called when a block is connected to the chain
func (m *ServiceBase) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) error {
	return nil
}
