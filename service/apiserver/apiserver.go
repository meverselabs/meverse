package apiserver

import (
	"sync"

	"github.com/labstack/echo"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
)

type INode interface {
	AddTx(tx *types.Transaction, sig common.Signature) error
}

// APIServer provides json rpc and web service for the chain
type APIServer struct {
	types.ServiceBase
	sync.Mutex
	e      *echo.Echo
	subMap map[string]*JRPCSub
}

// NewAPIServer returns a APIServer
func NewAPIServer() *APIServer {
	s := &APIServer{
		e:      echo.New(),
		subMap: map[string]*JRPCSub{},
	}
	return s
}

// Name returns the name of the service
func (s *APIServer) Name() string {
	return "fleta.apiserver"
}

// OnLoadChain called when the chain loaded
func (s *APIServer) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *APIServer) OnBlockConnected(b *types.Block, loader types.Loader) {
}

// OnTransactionInPoolExpired called when the tx expired
func (s *APIServer) OnTransactionInPoolExpired(txs []*types.Transaction) {
}

// OnTransactionFail called when the tx fail
func (s *APIServer) OnTransactionFail(height uint32, txs []*types.Transaction, err []error) {
}
