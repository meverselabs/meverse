package txpool

import (
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
)

// AccountTransaction is an interface that defines common functions of account model based transactions
type AccountTransaction interface {
	Seq() uint64
}

// TransactionPool provides a transaction queue
// User can push transaction regardless of UTXO model based transactions or account model based transactions
// If the sequence of the account model based transaction is not reached to the next of the last sequence, it doens't poped
type TransactionPool struct {
	sync.Mutex
	turnQ        *queue.Queue
	numberQ      *queue.Queue
	utxoQ        *queue.LinkedQueue
	txidMap      map[hash.Hash256]bool
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
		txidMap:      map[hash.Hash256]bool{},
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

	return tp.txidMap[TxHash]
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
func (tp *TransactionPool) Push(t types.Transaction, sigs []common.Signature) error {
	tp.Lock()
	defer tp.Unlock()

	TxHash := chain.HashTransaction(t)
	if tp.txidMap[TxHash] {
		return ErrExistTransaction
	}

	item := &PoolItem{
		Transaction: t,
		TxHash:      TxHash,
		Signatures:  sigs,
	}
	tx, is := t.(AccountTransaction)
	if !is {
		tp.utxoQ.Push(TxHash, item)
		tp.turnQ.Push(true)
	} else {
		addr := tx.From()
		q, has := tp.bucketMap[addr]
		if !has {
			q = queue.NewSortedQueue()
			tp.bucketMap[addr] = q
		}
		q.Insert(item, tx.Seq())
		tp.numberQ.Push(addr)
		tp.turnQ.Push(false)
	}
	tp.txidMap[TxHash] = true
	return nil
}

// Remove deletes the target transaction from the queue
// If it is an account model based transaction, it will be sorted by the sequence in the address
func (tp *TransactionPool) Remove(TxHash hash.Hash256, t types.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	tx, is := t.(AccountTransaction)
	if !is {
		if tp.utxoQ.Remove(TxHash) != nil {
			tp.turnOutMap[true]++
			delete(tp.txidMap, TxHash)
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
				if tx.Seq() < item.Transaction.(AccountTransaction).Seq() {
					break
				}
				q.Pop()
				delete(tp.txidMap, chain.HashTransaction(item.Transaction))
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
		delete(tp.txidMap, chain.HashTransaction(item.Transaction))
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
			if item.Transaction.(AccountTransaction).Seq() != lastSeq+1 {
				ignoreMap[addr] = true
				tp.numberQ.Push(addr)
				if remain > 0 {
					continue
				} else {
					return nil
				}
			}
			q.Pop()
			delete(tp.txidMap, chain.HashTransaction(item.Transaction))
			if q.Size() == 0 {
				delete(tp.bucketMap, addr)
			}
			return item
		}
	}
}

// PoolItem represents the item of the queue
type PoolItem struct {
	Transaction types.Transaction
	TxHash      hash.Hash256
	Signatures  []common.Signature
}
