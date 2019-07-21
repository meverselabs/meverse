package vault

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

type FeeTransaction interface {
	From() common.Address
	Fee(lw types.LoaderWrapper) *amount.Amount
}
