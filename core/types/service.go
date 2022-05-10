package types

// Service defines service functions
type Service interface {
	Name() string
	OnLoadChain(loader Loader) error
	OnBlockConnected(b *Block, loader Loader)
	OnTransactionInPoolExpired(txs []*Transaction)
	OnTransactionFail(height uint32, txs []*Transaction, err []error)
}

// ServiceBase is a base handler of the chain service
type ServiceBase struct{}

// OnLoadChain called when the chain loaded
func (s *ServiceBase) OnLoadChain(loader Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *ServiceBase) OnBlockConnected(b *Block, loader Loader) {
}

// OnTransactionInPoolExpired called when a transaction in pool is expired
func (s *ServiceBase) OnTransactionInPoolExpired(txs []*Transaction) {
}

// OnTransactionFail called when a transaction in pool is expired
func (s *ServiceBase) OnTransactionFail(height uint32, txs []*Transaction, err []error) {
}
