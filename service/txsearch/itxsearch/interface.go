package itxsearch

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type ITxSearch interface {
	BlockHeight(bh hash.Hash256) (uint32, error)
	BlockList(index int) []*BlockInfo
	Block(i uint32) (*types.Block, error)
	TxIndex(th hash.Hash256) (TxID, error)
	TxList(index, size int) ([]TxList, error)
	Tx(height uint32, index uint16) (map[string]interface{}, error)
	AddressTxList(From common.Address, index, size int) ([]TxList, error)
	TokenTxList(From common.Address, index, size int) ([]TxList, error)
	Reward(cont, rewarder common.Address) (*amount.Amount, error)
}
