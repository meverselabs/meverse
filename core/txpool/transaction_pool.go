package txpool

import (
	"bytes"
	"math"
	"strconv"
	"sync"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/queue"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// TransactionPool provides a transaction queue
// User can push transaction regardless of UTXO model based transactions or account model based transactions
type TransactionPool struct {
	sync.Mutex
	slotQue queue.IDoubleKeyQueue
}

// NewTransactionPool returns a TransactionPool
func NewTransactionPool() *TransactionPool {
	tp := &TransactionPool{
		slotQue: queue.NewDoubleKeyPriorityQueue(queue.LOWEST),
	}
	return tp
}

// IsExist checks that the transaction hash is inserted or not
func (tp *TransactionPool) IsExist(TxHash hash.Hash256) bool {
	tp.Lock()
	defer tp.Unlock()

	return tp.slotQue.Get(TxHash) != nil
}

// Size returns the size of TxPool
func (tp *TransactionPool) Size() int {
	tp.Lock()
	defer tp.Unlock()

	return tp.slotQue.Size()
}

// UnsafeSize returns the size of TxPool without mutex
func (tp *TransactionPool) UnsafeSize() int {
	return tp.slotQue.Size()
}

// Size returns the size of TxPool
func (tp *TransactionPool) GasLevel() (glv uint16) {
	l := tp.slotQue.Size()
	if 65535 < l {
		glv = glv - 1
	} else {
		glv = uint16(l)
	}
	switch true {
	case glv < 100:
		return 10
	case glv < 500:
		return 15
	case glv < 1000:
		return 100
	case glv < 2000:
		return 500
	default:
		return 5000
	}
}

// Push inserts the transaction and signatures of it by base model
// An UTXO model based transaction will be handled by FIFO
func (tp *TransactionPool) Push(TxHash hash.Hash256, tx *types.Transaction, sig common.Signature, signer common.Address) error {
	tp.Lock()
	defer tp.Unlock()

	if tp.slotQue.Get(TxHash) != nil {
		return errors.WithStack(ErrExistTransaction)
	}

	slot := types.ToTimeSlot(tx.Timestamp)
	item := &PoolItem{
		TxHash:      TxHash,
		Transaction: tx,
		Signature:   sig,
		Signer:      signer,
		Slot:        slot,
	}

	tp.slotQue.Push(item)
	return nil
}

// Get returns the pool item of the hash
func (tp *TransactionPool) Get(TxHash hash.Hash256) *PoolItem {
	tp.Lock()
	defer tp.Unlock()

	i := tp.slotQue.Get(TxHash)
	if i == nil {
		return nil
	}
	return i.(*PoolItem)
}

// Remove deletes the target transaction from the queue
func (tp *TransactionPool) Remove(TxHash hash.Hash256, tx *types.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	tp.slotQue.RemoveKey(TxHash)
}

// Clean removes outdated slot queue
func (tp *TransactionPool) Clean(currentSlot uint32) []*types.Transaction {
	tp.Lock()
	defer tp.Unlock()

	items := []*types.Transaction{}

	tp.slotQue.Iter(func(i queue.IDoubleKeyQueueItem) bool {
		if i.DKey().(uint32) < currentSlot-1 {
			items = append(items, i.(*PoolItem).Transaction)
			return true
		}
		return false
	})
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
	item := tp.slotQue.Pop()
	if item == nil {
		return nil
	}
	return item.(*PoolItem)
}

// PoolItem represents the item of the queue
type PoolItem struct {
	TxHash      hash.Hash256
	Transaction *types.Transaction
	Signature   common.Signature
	Signer      common.Address
	Slot        uint32
}

func (pi *PoolItem) DKey() interface{} {
	return pi.Slot
}
func (pi *PoolItem) DPriority() uint64 {
	return uint64(pi.Slot)
}
func (pi *PoolItem) Key() interface{} {
	return pi.TxHash
}

func (pi *PoolItem) Priority() uint64 {
	if pi.Transaction.IsEtherType {
		gasFactor := uint64(math.MaxUint32 - pi.Transaction.GasPrice.Uint64())
		return uint64(gasFactor<<47) /*less then int64 max*/ + pi.Transaction.Seq
	}
	if pi.Transaction.GasPrice == nil {
		return uint64(math.MaxUint32)
	}
	return uint64(math.MaxUint32 - pi.Transaction.GasPrice.Uint64())
}

// List return txpool list
func (tp *TransactionPool) List() []*PoolItem {
	tp.Lock()
	defer tp.Unlock()

	pis := []*PoolItem{}

	list := tp.slotQue.List()

	for _, i := range list {
		item := i.(*PoolItem)
		pis = append(pis, &PoolItem{
			TxHash:      item.TxHash,
			Transaction: item.Transaction,
			Signature:   item.Signature,
			Signer:      item.Signer,
		})
	}
	return pis
}

// Dump do dump
func (tp *TransactionPool) Dump() string {
	tp.Lock()
	defer tp.Unlock()

	var buffer bytes.Buffer
	list := tp.slotQue.List()

	if tp.slotQue.Size() > 0 {
		buffer.WriteString("slotQue\n")
		for _, i := range list {
			item := i.(*PoolItem)
			buffer.WriteString(strconv.FormatUint(item.DKey().(uint64), 10))
			buffer.WriteString(":")
			buffer.WriteString(item.Key().(string))
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}
