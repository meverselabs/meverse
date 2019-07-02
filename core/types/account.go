package types

import (
	"encoding/json"

	"github.com/fletaio/fleta/common"
)

// Account is a interface that defines common account functions
type Account interface {
	json.Marshaler
	Address() common.Address
	Name() string
	Clone() Account
	Validate(loader LoaderProcess, signers []common.PublicHash) error
}
