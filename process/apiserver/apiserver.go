package apiserver

import (
	"sync"

	"github.com/fletaio/fleta/core/types"
	"github.com/labstack/echo"
)

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

// Init called when initialize service
func (s *APIServer) Init(pm types.ProcessManager, cn types.Provider) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (s *APIServer) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *APIServer) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) error {
	return nil
}
