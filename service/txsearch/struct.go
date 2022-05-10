package txsearch

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/syndtr/goleveldb/leveldb"
)

type TxID struct {
	Height uint32
	Index  uint16
	Err    error
}

type ITxSearch interface {
	BlockHeight(bh hash.Hash256) (uint32, error)
	BlockList(index int) []*BlockInfo
	Block(i uint32) (*types.Block, error)
	TxIndex(th hash.Hash256) (TxID, error)
	TxList(index int) ([]TxList, error)
	Tx(height uint32, index uint16) (map[string]interface{}, error)
	AddressTxList(From common.Address, index int) ([]TxList, error)
	TokenTxList(From common.Address, index int) ([]TxList, error)
	Reward(cont, rewarder common.Address) (*amount.Amount, error)
}

type TxSearch struct {
	db  *leveldb.DB
	st  *chain.Store
	cn  *chain.Chain
	api *apiserver.APIServer
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
