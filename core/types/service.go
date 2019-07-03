package types

// Service is a interface of the chain service
type Service interface {
	Name() string
	Init(pm ProcessManager, cn Provider) error
	OnBlockConnected(b *Block, events []Event, loader Loader) error
}

// ServiceBase is a base handler of the chain service
type ServiceBase struct{}

// OnBlockConnected called when a block is connected to the chain
func (m *ServiceBase) OnBlockConnected(b *Block, events []Event, loader Loader) error {
	return nil
}
