package txpool

import (
	"bytes"
	"strconv"
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
	txhashMap    map[hash.Hash256]*PoolItem
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
		txhashMap:    map[hash.Hash256]*PoolItem{},
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

	_, has := tp.txhashMap[TxHash]
	return has
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
func (tp *TransactionPool) Push(t uint16, TxHash hash.Hash256, tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	tp.Lock()
	defer tp.Unlock()

	if _, has := tp.txhashMap[TxHash]; has {
		return ErrExistTransaction
	}

	item := &PoolItem{
		TxType:      t,
		TxHash:      TxHash,
		Transaction: tx,
		Signatures:  sigs,
		Signers:     signers,
	}
	atx, is := tx.(chain.AccountTransaction)
	if !is {
		if tp.utxoQ.Push(TxHash, item) {
			tp.turnQ.Push(true)
		} else {
			return ErrExistTransaction
		}
	} else {
		addr := atx.From()
		q, has := tp.bucketMap[addr]
		if !has {
			q = queue.NewSortedQueue()
			tp.bucketMap[addr] = q
		}
		if q.FindOrInsert(item, atx.Seq()) != nil {
			return ErrExistTransactionSeq
		}
		tp.numberQ.Push(addr)
		tp.turnQ.Push(false)
	}
	tp.txhashMap[TxHash] = item
	return nil
}

// Get returns the pool item of the hash
func (tp *TransactionPool) Get(TxHash hash.Hash256) *PoolItem {
	tp.Lock()
	defer tp.Unlock()

	return tp.txhashMap[TxHash]
}

// Remove deletes the target transaction from the queue
// If it is an account model based transaction, it will be sorted by the sequence in the address
func (tp *TransactionPool) Remove(TxHash hash.Hash256, t types.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	if tx, is := t.(chain.AccountTransaction); !is {
		if tp.utxoQ.Remove(TxHash) != nil {
			turn := tp.turnQ.Peek()
			if turn != nil {
				if turn == true {
					tp.turnQ.Pop()
				} else {
					tp.turnOutMap[true]++
				}
			}
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
				if item.Transaction.(chain.AccountTransaction).Seq() > tx.Seq() {
					break
				}
				q.Pop()
				delete(tp.txhashMap, item.TxHash)
				turn := tp.turnQ.Peek()
				if turn != nil {
					if turn == false {
						tp.turnQ.Pop()
					} else {
						tp.turnOutMap[false]++
					}
				}
				number := tp.numberQ.Peek()
				if number != nil {
					if number.(common.Address) == addr {
						tp.numberQ.Pop()
					} else {
						tp.numberOutMap[addr]++
					}
				}
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
	var turn bool
	for {
		if tp.turnQ.Size() == 0 {
			return nil
		}
		turn = tp.turnQ.Pop().(bool)
		tout := tp.turnOutMap[turn]
		if tout > 0 {
			tout--
			if tout == 0 {
				delete(tp.turnOutMap, turn)
			} else {
				tp.turnOutMap[turn] = tout
			}
		} else {
			break
		}
	}
	if turn {
		item := tp.utxoQ.Pop().(*PoolItem)
		delete(tp.txhashMap, item.TxHash)
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
				} else {
					break
				}
			}
			if ignoreMap[addr] {
				tp.numberQ.Push(addr)
				if remain > 0 {
					continue
				} else {
					tp.turnQ.Push(false)
					return nil
				}
			}
			q := tp.bucketMap[addr]
			v, _ := q.Peek()
			item := v.(*PoolItem)
			lastSeq := SeqCache.Seq(addr)
			txSeq := item.Transaction.(chain.AccountTransaction).Seq()
			if txSeq < lastSeq+1 {
				q.Pop()
				delete(tp.txhashMap, item.TxHash)
				if q.Size() == 0 {
					delete(tp.bucketMap, addr)
				}
				if remain > 0 {
					turn := tp.turnQ.Peek()
					if turn != nil {
						if turn == false {
							tp.turnQ.Pop()
						} else {
							tp.turnOutMap[false]++
						}
					}
					continue
				} else {
					return nil
				}
			} else if txSeq > lastSeq+1 {
				ignoreMap[addr] = true
				tp.numberQ.Push(addr)
				if remain > 0 {
					continue
				} else {
					tp.turnQ.Push(false)
					return nil
				}
			}
			q.Pop()
			delete(tp.txhashMap, item.TxHash)
			if q.Size() == 0 {
				delete(tp.bucketMap, addr)
			}
			return item
		}
	}
}

// PoolItem represents the item of the queue
type PoolItem struct {
	TxType      uint16
	TxHash      hash.Hash256
	Transaction types.Transaction
	Signatures  []common.Signature
	Signers     []common.PublicHash
}

// List return txpool list
func (tp *TransactionPool) List() []*PoolItem {
	tp.Lock()
	defer tp.Unlock()

	pis := []*PoolItem{}

	for _, item := range tp.txhashMap {
		pis = append(pis, &PoolItem{
			TxType:      item.TxType,
			TxHash:      item.TxHash,
			Transaction: item.Transaction,
			Signatures:  item.Signatures,
			Signers:     item.Signers,
		})
	}
	return pis
}

// Dump do dump
func (tp *TransactionPool) Dump() string {
	tp.Lock()
	defer tp.Unlock()

	var buffer bytes.Buffer
	if tp.turnQ.Size() > 0 {
		buffer.WriteString("turnQ\n")
		tp.turnQ.Iter(func(value interface{}) {
			if value.(bool) {
				buffer.WriteString("true")
			} else {
				buffer.WriteString("false")
			}
			buffer.WriteString("\n")
		})
		buffer.WriteString("\n")
	}
	if tp.numberQ.Size() > 0 {
		buffer.WriteString("numberQ\n")
		tp.numberQ.Iter(func(value interface{}) {
			buffer.WriteString(value.(common.Address).String())
			buffer.WriteString("\n")
		})
		buffer.WriteString("\n")
	}
	if len(tp.txhashMap) > 0 {
		buffer.WriteString("txhashMap\n")
		for k := range tp.txhashMap {
			buffer.WriteString(k.String())
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	if len(tp.turnOutMap) > 0 {
		buffer.WriteString("turnOutMap\n")
		for k, v := range tp.turnOutMap {
			if k {
				buffer.WriteString("true")
			} else {
				buffer.WriteString("false")
			}
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(v))
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	if len(tp.numberOutMap) > 0 {
		buffer.WriteString("numberOutMap\n")
		for k, v := range tp.numberOutMap {
			buffer.WriteString(k.String())
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(v))
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	if len(tp.bucketMap) > 0 {
		buffer.WriteString("bucketMap\n")
		for k, v := range tp.bucketMap {
			buffer.WriteString(k.String())
			buffer.WriteString(":")
			buffer.WriteString("\n")
			v.Iter(func(value interface{}, priority uint64) {
				buffer.WriteString(strconv.FormatUint(priority, 10))
				buffer.WriteString(":")
				k := value.(*PoolItem)
				buffer.WriteString(k.TxHash.String())
				buffer.WriteString("\n")
			})
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}
