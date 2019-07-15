package types

// Service defines service functions
type Service interface {
	Name() string
	Init(pm ProcessManager, cn Provider) error
	OnLoadChain(loader Loader) error
	OnBlockConnected(b *Block, events []Event, loader Loader) error
}

// ServiceBase is a base handler of the chain service
type ServiceBase struct{}

// OnLoadChain called when the chain loaded
func (s *ServiceBase) OnLoadChain(loader Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *ServiceBase) OnBlockConnected(b *Block, events []Event, loader Loader) error {
	return nil
}
