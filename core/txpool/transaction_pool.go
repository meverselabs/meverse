package txpool

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/core/types"
)

// TransactionPool provides a transaction queue
// User can push transaction regardless of UTXO model based transactions or account model based transactions
type TransactionPool struct {
	sync.Mutex
	slotMap   map[uint32]*queue.LinkedQueue
	txhashMap map[hash.Hash256]*PoolItem
}

// NewTransactionPool returns a TransactionPool
func NewTransactionPool() *TransactionPool {
	tp := &TransactionPool{
		slotMap:   map[uint32]*queue.LinkedQueue{},
		txhashMap: map[hash.Hash256]*PoolItem{},
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

	return len(tp.txhashMap)
}

// Push inserts the transaction and signatures of it by base model
// An UTXO model based transaction will be handled by FIFO
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
	slot := types.ToTimeSlot(tx.Timestamp())
	q, has := tp.slotMap[slot]
	if !has {
		q = queue.NewLinkedQueue()
		tp.slotMap[slot] = q
	}
	q.Push(TxHash, item)
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
func (tp *TransactionPool) Remove(TxHash hash.Hash256, tx types.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	slot := types.ToTimeSlot(tx.Timestamp())
	q, has := tp.slotMap[slot]
	if has {
		q.Remove(TxHash)
		delete(tp.txhashMap, TxHash)
	}
}

// Clean removes outdated slot queue
func (tp *TransactionPool) Clean(currentSlot uint32) []types.Transaction {
	tp.Lock()
	defer tp.Unlock()

	deletes := []uint32{}
	for slot := range tp.slotMap {
		if slot < currentSlot-1 {
			deletes = append(deletes, slot)
		}
	}
	items := []types.Transaction{}
	for _, v := range deletes {
		if q, has := tp.slotMap[v]; has {
			q.Iter(func(key hash.Hash256, value interface{}) {
				delete(tp.txhashMap, key)
				items = append(items, value.(*PoolItem).Transaction)
			})
			delete(tp.slotMap, v)
		}
	}
	return items
}

// Pop returns and removes the proper transaction
func (tp *TransactionPool) Pop(currentSlot uint32) *PoolItem {
	tp.Lock()
	defer tp.Unlock()

	return tp.UnsafePop(currentSlot)
}

// UnsafePop returns and removes the proper transaction without mutex locking
func (tp *TransactionPool) UnsafePop(currentSlot uint32) *PoolItem {
	slots := []uint32{}
	for slot := range tp.slotMap {
		slots = append(slots, slot)
	}

	if q, has := tp.slotMap[currentSlot-1]; has {
		v := q.Pop()
		if v != nil {
			return v.(*PoolItem)
		} else {
			delete(tp.slotMap, currentSlot-1)
		}
	}
	if q, has := tp.slotMap[currentSlot]; has {
		v := q.Pop()
		if v != nil {
			return v.(*PoolItem)
		}
	}
	return nil
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
	if len(tp.slotMap) > 0 {
		buffer.WriteString("slotMap\n")
		for k, v := range tp.slotMap {
			buffer.WriteString(strconv.FormatUint(uint64(k), 10))
			buffer.WriteString(":")
			v.Iter(func(key hash.Hash256, value interface{}) {
				buffer.WriteString(key.String())
				buffer.WriteString("\n")
			})
			buffer.WriteString("\n")
			buffer.WriteString("\n")
		}
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
	return buffer.String()
}
