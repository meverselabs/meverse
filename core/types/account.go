package types

import (
	"github.com/fletaio/fleta/common"
)

// Account is a interface that defines common account functions
type Account interface {
	Address() common.Address
	Name() string
	Clone() Account
	Validate(loader Loader, signers []common.PublicHash) error
}
