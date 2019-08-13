package txpool

import (
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
)

// TransactionPool provides a transaction queue
// User can push transaction regardless of UTXO model based transactions or account model based transactions
// If the sequence of the account model based transaction is not reached to the next of the last sequence, it doens't poped
type TransactionPool struct {
	sync.Mutex
	turnQ        *queue.Queue
	numberQ      *queue.Queue
	utxoQ        *queue.LinkedQueue
	txhashMap    map[hash.Hash256]bool
	turnOutMap   map[bool]int
	numberOutMap map[common.Address]int
	bucketMap    map[common.Address]*queue.SortedQueue
}

// NewTransactionPool returns a TransactionPool
func NewTransactionPool() *TransactionPool {
	tp := &TransactionPool{
		turnQ:        queue.NewQueue(),
		numberQ:      queue.NewQueue(),
		utxoQ:        queue.NewLinkedQueue(),
		txhashMap:    map[hash.Hash256]bool{},
		turnOutMap:   map[bool]int{},
		numberOutMap: map[common.Address]int{},
		bucketMap:    map[common.Address]*queue.SortedQueue{},
	}
	return tp
}

// IsExist checks that the transaction hash is inserted or not
func (tp *TransactionPool) IsExist(TxHash hash.Hash256) bool {
	tp.Lock()
	defer tp.Unlock()

	return tp.txhashMap[TxHash]
}

// Size returns the size of TxPool
func (tp *TransactionPool) Size() int {
	tp.Lock()
	defer tp.Unlock()

	sum := 0
	for _, v := range tp.turnOutMap {
		sum += v
	}
	return tp.turnQ.Size() - sum
}

// Push inserts the transaction and signatures of it by base model and sequence
// An UTXO model based transaction will be handled by FIFO
// An account model based transaction will be sorted by the sequence value
func (tp *TransactionPool) Push(ChainID uint8, t uint16, TxHash hash.Hash256, tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	tp.Lock()
	defer tp.Unlock()

	if tp.txhashMap[TxHash] {
		return ErrExistTransaction
	}

	item := &PoolItem{
		ChainID:     ChainID,
		TxType:      t,
		TxHash:      TxHash,
		Transaction: tx,
		Signatures:  sigs,
		Signers:     signers,
	}
	atx, is := tx.(chain.AccountTransaction)
	if !is {
		tp.utxoQ.Push(TxHash, item)
		tp.turnQ.Push(true)
	} else {
		addr := atx.From()
		q, has := tp.bucketMap[addr]
		if !has {
			q = queue.NewSortedQueue()
			tp.bucketMap[addr] = q
		}
		q.Insert(item, atx.Seq())
		tp.numberQ.Push(addr)
		tp.turnQ.Push(false)
	}
	tp.txhashMap[TxHash] = true
	return nil
}

// Remove deletes the target transaction from the queue
// If it is an account model based transaction, it will be sorted by the sequence in the address
func (tp *TransactionPool) Remove(TxHash hash.Hash256, t types.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	tx, is := t.(chain.AccountTransaction)
	if !is {
		if tp.utxoQ.Remove(TxHash) != nil {
			tp.turnOutMap[true]++
			delete(tp.txhashMap, TxHash)
		}
	} else {
		addr := tx.From()
		if q, has := tp.bucketMap[addr]; has {
			for {
				if q.Size() == 0 {
					break
				}
				v, _ := q.Peek()
				item := v.(*PoolItem)
				if tx.Seq() < item.Transaction.(chain.AccountTransaction).Seq() {
					break
				}
				q.Pop()
				delete(tp.txhashMap, chain.HashTransaction(item.ChainID, item.Transaction))
				tp.turnOutMap[false]++
				tp.numberOutMap[addr]++
			}
			if q.Size() == 0 {
				delete(tp.bucketMap, addr)
			}
		}
	}
}

// Pop returns and removes the proper transaction
func (tp *TransactionPool) Pop(SeqCache SeqCache) *PoolItem {
	tp.Lock()
	defer tp.Unlock()

	return tp.UnsafePop(SeqCache)
}

// UnsafePop returns and removes the proper transaction without mutex locking
func (tp *TransactionPool) UnsafePop(SeqCache SeqCache) *PoolItem {
	var bTurn bool
	for {
		turn := tp.turnQ.Pop()
		if turn == nil {
			return nil
		}
		bTurn = turn.(bool)
		tout := tp.turnOutMap[bTurn]
		if tout > 0 {
			tp.turnOutMap[bTurn] = tout - 1
			continue
		}
		break
	}
	if bTurn {
		item := tp.utxoQ.Pop().(*PoolItem)
		delete(tp.txhashMap, chain.HashTransaction(item.ChainID, item.Transaction))
		return item
	} else {
		remain := tp.numberQ.Size()
		ignoreMap := map[common.Address]bool{}
		for {
			var addr common.Address
			for {
				if tp.numberQ.Size() == 0 {
					return nil
				}
				addr = tp.numberQ.Pop().(common.Address)
				remain--
				nout := tp.numberOutMap[addr]
				if nout > 0 {
					nout--
					if nout == 0 {
						delete(tp.numberOutMap, addr)
					} else {
						tp.numberOutMap[addr] = nout
					}
					continue
				}
				break
			}
			if ignoreMap[addr] {
				tp.numberQ.Push(addr)
				if remain > 0 {
					continue
				} else {
					return nil
				}
			}
			q := tp.bucketMap[addr]
			v, _ := q.Peek()
			item := v.(*PoolItem)
			lastSeq := SeqCache.Seq(addr)
			if item.Transaction.(chain.AccountTransaction).Seq() != lastSeq+1 {
				ignoreMap[addr] = true
				tp.numberQ.Push(addr)
				if remain > 0 {
					continue
				} else {
					return nil
				}
			}
			q.Pop()
			delete(tp.txhashMap, chain.HashTransaction(item.ChainID, item.Transaction))
			if q.Size() == 0 {
				delete(tp.bucketMap, addr)
			}
			return item
		}
	}
}

// PoolItem represents the item of the queue
type PoolItem struct {
	ChainID     uint8
	TxType      uint16
	TxHash      hash.Hash256
	Transaction types.Transaction
	Signatures  []common.Signature
	Signers     []common.PublicHash
}
