package bank

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

func (s *Bank) SendTx(tx types.Transaction, sigs []common.Signature) (hash.Hash256, error) {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return hash.Hash256{}, err
	}

	TxHash := chain.HashTransactionByType(s.cn.ChainID(), t, tx)

	ch := make(chan string)
	s.Lock()
	s.waitTxMap[TxHash] = &ch
	s.Unlock()

	// if s.hs != nil {
	// 	if err := s.hs.SendTx(tx, sigs); err != nil {
	// 		s.Lock()
	// 		if c, has := s.waitTxMap[TxHash]; has {
	// 			close(*c)
	// 		}
	// 		delete(s.waitTxMap, TxHash)
	// 		s.Unlock()
	// 		return hash.Hash256{}, err
	// 	}
	// } else
	if err := s.nd.AddTx(tx, sigs); err != nil {
		s.Lock()
		if c, has := s.waitTxMap[TxHash]; has {
			close(*c)
		}
		delete(s.waitTxMap, TxHash)
		s.Unlock()
		return hash.Hash256{}, err
	}
	return TxHash, nil
}

func (s *Bank) WaitTx(TxHash hash.Hash256, wait time.Duration) (string, error) {
	s.Lock()
	pCh, has := s.waitTxMap[TxHash]
	s.Unlock()
	if !has {
		return "", ErrInvalidTransactionHash
	}

	timer := time.NewTimer(wait)
	select {
	case <-timer.C:
		s.Lock()
		delete(s.waitTxMap, TxHash)
		s.Unlock()
		return "", ErrTransactionTimeout
	case TXID := <-(*pCh):
		s.Lock()
		delete(s.waitTxMap, TxHash)
		s.Unlock()
		if len(TXID) == 0 {
			return "", ErrTransactionFailed
		}
		return TXID, nil
	}
}
