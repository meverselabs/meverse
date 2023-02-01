package itxsearch

import (
	etypes "github.com/ethereum/go-ethereum/core/types"
	mtypes "github.com/meverselabs/meverse/core/types"

	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

type BlockInfo struct {
	Height    uint32
	Hash      string
	TxLen     uint16
	Timestamp uint64
}

type TxID struct {
	Height uint32
	Index  uint16
	Err    error
}

type TxList struct {
	TxID     string
	From     string
	To       string
	Contract string
	Method   string
	Amount   string
}

type MethodCallEvent struct {
	types.MethodCallEvent
	ToName string
}

type ContractName interface {
	Name() string
}
type ContractNameCC interface {
	Name(cc types.ContractLoader) string
}

type BloomInterface interface {
	FindTransactionsEvents(txs []*types.Transaction, evs []*types.Event, idx int) ([]*types.Event, error)
	CreateEventBloom(ctx *types.Context, events []*types.Event) (etypes.Bloom, error)
	EventsToLogs(chain *chain.Chain, header *mtypes.Header, tx *mtypes.Transaction, evs []*mtypes.Event, idx int) ([]*etypes.Log, error)
}
