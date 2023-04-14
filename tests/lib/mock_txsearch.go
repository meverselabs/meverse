package testlib

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
)

// tsMock is the mock of txsearch service
// range로 검색할 경우 txsearch는 필요없기 때문에 mock으로 가능
type TsMock struct {
	provider types.Provider
	blocks   []*types.Block
}

func NewTsMocK(provider types.Provider) *TsMock {
	return &TsMock{provider: provider}
}

// Name returns the name of the service
func (t *TsMock) Name() string {
	return "tsMock"
}

// OnLoadChain called when the chain loaded
func (t *TsMock) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (t *TsMock) OnBlockConnected(b *types.Block, loader types.Loader) {
	t.AddBlock(b)
}

// OnLoadChain called when the chain loaded
func (t *TsMock) OnTransactionInPoolExpired(txs []*types.Transaction) {}

// OnTransactionFail called when the tx fail
func (t *TsMock) OnTransactionFail(height uint32, txs []*types.Transaction, err []error) {}

func (t *TsMock) AddBlock(b *types.Block) {
	t.blocks = append(t.blocks, b)
}

func (t *TsMock) BlockHeight(bh hash.Hash256) (uint32, error) {
	for _, b := range t.blocks {

		if hash, err := t.provider.Hash(b.Header.Height); err != nil {
			return 0, err
		} else if hash == bh {
			return b.Header.Height, nil
		}
	}
	return 0, errors.New("Block " + bh.String() + " Not round")
}
func (t *TsMock) BlockList(index int) []*itxsearch.BlockInfo { return nil }

func (t *TsMock) Block(i uint32) (*types.Block, error) { return nil, nil }

func (t *TsMock) TxIndex(th hash.Hash256) (itxsearch.TxID, error) {

	for _, b := range t.blocks {
		for i, tx := range b.Body.Transactions {
			if tx.HashSig() == th {
				return itxsearch.TxID{Height: b.Header.Height, Index: uint16(i)}, nil
			}
		}
	}
	return itxsearch.TxID{}, nil

}
func (t *TsMock) TxList(index, size int) ([]itxsearch.TxList, error)             { return nil, nil }
func (t *TsMock) Tx(height uint32, index uint16) (map[string]interface{}, error) { return nil, nil }
func (t *TsMock) AddressTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}

func (t *TsMock) TokenTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}
func (t *TsMock) Reward(cont, rewarder common.Address) (*amount.Amount, error) { return nil, nil }
