package bank

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
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

func (s *Bank) addTransfer(txid string, tx *vault.Transfer) error {
	if _, err := s.db.RPush(toTransferListKey(tx.From()), []byte(txid)); err != nil {
		return err
	}
	if tx.From() != tx.To {
		if _, err := s.db.RPush(toTransferListKey(tx.To), []byte(txid)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Bank) addTransaction(txid string, t uint16, at chain.AccountTransaction, res uint8) error {
	if res == 1 {
		switch tx := at.(type) {
		case *vault.Transfer:
			if tx.From() != tx.To {
				if _, err := s.db.RPush(toTransactionListKey(tx.To), []byte(txid)); err != nil {
					return err
				}
			}
		case *formulator.Unstaking:
			if _, err := s.db.RPush(toTransactionListKey(tx.HyperFormulator), []byte(txid)); err != nil {
				return err
			}
		case *formulator.RevertUnstaking:
			if _, err := s.db.RPush(toTransactionListKey(tx.HyperFormulator), []byte(txid)); err != nil {
				return err
			}
		}
	}
	data, err := encoding.Marshal(at)
	if err != nil {
		return err
	}
	bs, err := encoding.Marshal(&Transaction{
		Type:   t,
		Data:   data,
		Result: res,
	})
	if err != nil {
		return err
	}
	if _, err := s.db.HSet(tagTransaction, []byte(txid), bs); err != nil {
		return err
	}
	return nil
}

func (s *Bank) addAccount(addr common.Address, KeyHash common.PublicHash) error {
	if _, err := s.db.HSet(tagAddressKeyHash, addr[:], KeyHash[:]); err != nil {
		return err
	}
	if _, err := s.db.HSet(toAccountKey(KeyHash), addr[:], []byte{1}); err != nil {
		return err
	}
	return nil
}

func (s *Bank) removeAccount(addr common.Address) error {
	bs, err := s.db.HGet(tagAddressKeyHash, addr[:])
	if err != nil {
		return err
	}
	var KeyHash common.PublicHash
	copy(KeyHash[:], bs)

	if _, err := s.db.HDel(tagAddressKeyHash, addr[:]); err != nil {
		return err
	}
	if _, err := s.db.HDel(toAccountKey(KeyHash), addr[:]); err != nil {
		return err
	}
	return nil
}

func (s *Bank) Unstaking(HyperAddr common.Address, addr common.Address, UnstakedHeight uint32) *amount.Amount {
	bs, err := s.db.HGet(toUnstakingKey(addr), toUnstakingSubKey(HyperAddr, UnstakedHeight))
	if err != nil {
		return amount.NewCoinAmount(0, 0)
	}
	if len(bs) == 0 {
		return amount.NewCoinAmount(0, 0)
	}
	return amount.NewAmountFromBytes(bs)
}

func (s *Bank) addUnstaking(HyperAddr common.Address, addr common.Address, UnstakedHeight uint32, am *amount.Amount) error {
	if _, err := s.db.HSet(toUnstakingKey(addr), toUnstakingSubKey(HyperAddr, UnstakedHeight), s.Unstaking(HyperAddr, addr, UnstakedHeight).Add(am).Bytes()); err != nil {
		return err
	}
	return nil
}

func (s *Bank) removeUnstaking(HyperAddr common.Address, addr common.Address, UnstakedHeight uint32, am *amount.Amount) error {
	rem := s.Unstaking(HyperAddr, addr, UnstakedHeight).Sub(am)
	if rem.IsZero() {
		if _, err := s.db.HDel(toUnstakingKey(addr), toUnstakingSubKey(HyperAddr, UnstakedHeight)); err != nil {
			return err
		}
		if _, err := s.db.HLen(toUnstakingKey(addr)); err != nil {
			return err
		}
	} else {
		if _, err := s.db.HSet(toUnstakingKey(addr), toUnstakingSubKey(HyperAddr, UnstakedHeight), rem.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Bank) getTransaction(txid string) (types.Transaction, uint8, error) {
	bs, err := s.db.HGet(tagTransaction, []byte(txid))
	if err != nil {
		return nil, 0, err
	}
	item := &Transaction{}
	if err := encoding.Unmarshal(bs, &item); err != nil {
		return nil, 0, err
	}
	fc := encoding.Factory("transaction")
	t, err := fc.Create(item.Type)
	if err != nil {
		return nil, 0, err
	}
	if err := encoding.Unmarshal(item.Data, &t); err != nil {
		return nil, 0, err
	}
	return t.(types.Transaction), item.Result, nil
}

func (s *Bank) getTransactionFromTail(addr common.Address, offset int32, count int32) ([]string, []types.Transaction, []uint8, error) {
	v, err := s.db.LLen(toTransactionListKey(addr))
	if err != nil {
		return nil, nil, nil, err
	}
	from := int32(v) - count - offset
	if from < 0 {
		from = 0
	}
	to := int32(v) - offset
	if to < 0 {
		to = 0
	}
	return s.getTransactions(addr, from, to, true)
}

func (s *Bank) getTransactions(addr common.Address, from int32, to int32, reverse bool) ([]string, []types.Transaction, []uint8, error) {
	values, err := s.db.LRange(toTransactionListKey(addr), from, to)
	if err != nil {
		return nil, nil, nil, err
	}
	txids := make([]string, len(values))
	txs := make([]types.Transaction, len(values))
	results := make([]uint8, len(values))
	for i, bs := range values {
		txid := string(bs)
		tx, result, err := s.getTransaction(txid)
		if err != nil {
			return nil, nil, nil, err
		}
		if reverse {
			txids[len(values)-1-i] = txid
			txs[len(values)-1-i] = tx
			results[len(values)-1-i] = result
		} else {
			txids[i] = txid
			txs[i] = tx
			results[i] = result
		}
	}
	return txids, txs, results, nil
}
