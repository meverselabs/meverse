package chain

import (
	"github.com/fletaio/fleta/core/types"
)

// Application is a interface of the chain application
type Application interface {
	Process
	InitGenesis(ctp *types.ContextProcess) error
}
