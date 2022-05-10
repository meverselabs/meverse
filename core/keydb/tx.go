// Package keydb came from buntdb that implements a low-level in-memory key/value store in pure Go.
// It persists to disk, is ACID compliant, and uses locking for multiple
// readers and a single writer. Bunt is ideal for projects that need
// a dependable database, and favor speed over data size.
package keydb

import (
	"bytes"
	"time"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/pkg/errors"
	"github.com/tidwall/btree"
)

// Tx represents a transaction on the database. This transaction can either be
// read-only or read/write. Read-only transactions can be used for retrieving
// values for keys and iterating through keys and values. Read/write
// transactions can set and delete keys.
//
// All transactions must be committed or rolled-back when done.
type Tx struct {
	db       *DB             // the underlying database.
	writable bool            // when false mutable operations fail.
	funcd    bool            // when true Commit and Rollback panic.
	wc       *txWriteContext // context for writable transactions.
}

type txWriteContext struct {
	// rollback when deleteAll is called
	rbkeys *btree.BTree // a tree of all item ordered by key

	rollbackItems map[string]*dbItem // details for rolling back tx.
	commitItems   map[string]*dbItem // details for committing tx.
	itercount     int                // stack of iterators
}

// DeleteAll deletes all items from the database.
func (tx *Tx) DeleteAll() error {
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	} else if !tx.writable {
		return errors.WithStack(ErrTxNotWritable)
	} else if tx.wc.itercount > 0 {
		return errors.WithStack(ErrTxIterating)
	}

	// check to see if we've already deleted everything
	if tx.wc.rbkeys == nil {
		// we need to backup the live data in case of a rollback.
		tx.wc.rbkeys = tx.db.keys
	}

	// now reset the live database trees
	tx.db.keys = btree.New(btreeDegrees, nil)

	// always clear out the commits
	tx.wc.commitItems = make(map[string]*dbItem)

	return nil
}

// Begin opens a new transaction.
// Multiple read-only transactions can be opened at the same time but there can
// only be one read/write transaction at a time. Attempting to open a read/write
// transactions while another one is in progress will result in blocking until
// the current read/write transaction is completed.
//
// All transactions must be closed by calling Commit() or Rollback() when done.
func (db *DB) Begin(writable bool) (*Tx, error) {
	tx := &Tx{
		db:       db,
		writable: writable,
	}
	tx.lock()
	if db.closed {
		tx.unlock()
		return nil, errors.WithStack(ErrDatabaseClosed)
	}
	if writable {
		// writable transactions have a writeContext object that
		// contains information about changes to the database.
		tx.wc = &txWriteContext{}
		tx.wc.rollbackItems = make(map[string]*dbItem)
		tx.wc.commitItems = make(map[string]*dbItem)
	}
	return tx, nil
}

// lock locks the database based on the transaction type.
func (tx *Tx) lock() {
	if tx.writable {
		tx.db.mu.Lock()
	} else {
		tx.db.mu.RLock()
	}
}

// unlock unlocks the database based on the transaction type.
func (tx *Tx) unlock() {
	if tx.writable {
		tx.db.mu.Unlock()
	} else {
		tx.db.mu.RUnlock()
	}
}

// rollbackInner handles the underlying rollback logic.
// Intended to be called from Commit() and Rollback().
func (tx *Tx) rollbackInner() {
	// rollback the deleteAll if needed
	if tx.wc.rbkeys != nil {
		tx.db.keys = tx.wc.rbkeys
	}
	for key, item := range tx.wc.rollbackItems {
		tx.db.deleteFromDatabase(&dbItem{key: []byte(key)})
		if item != nil {
			// When an item is not nil, we will need to reinsert that item
			// into the database overwriting the current one.
			tx.db.insertIntoDatabase(item)
		}
	}
}

// Commit writes all changes to disk.
// An error is returned when a write error occurs, or when a Commit() is called
// from a read-only transaction.
func (tx *Tx) Commit() error {
	if tx.funcd {
		panic("managed tx commit not allowed")
	}
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	} else if !tx.writable {
		return errors.WithStack(ErrTxNotWritable)
	}
	var err error
	if len(tx.wc.commitItems) > 0 || tx.wc.rbkeys != nil {
		tx.db.buf = tx.db.buf[:0]
		// write a flushdb if a deleteAll was called.
		if tx.wc.rbkeys != nil {
			tx.db.buf = append(tx.db.buf, tagFlushDB)
		}
		// Each committed record is written to disk
		for key, item := range tx.wc.commitItems {
			if item == nil {
				tx.db.buf = (&dbItem{key: []byte(key)}).writeDeleteTo(tx.db.buf)
			} else {
				tx.db.buf = item.writeSetTo(tx.db.buf)
			}
		}
		tx.db.buf = append(tx.db.buf, tagTxEnd)
		// Flushing the buffer only once per transaction.
		// If this operation fails then the write did failed and we must
		// rollback.
		if _, err := tx.db.file.Seek(0, 2); err != nil {
			tx.rollbackInner()
		}
		if _, err := tx.db.file.Write(tx.db.buf); err != nil {
			tx.rollbackInner()
		}
		if tx.db.config.SyncPolicy == Always {
			_ = tx.db.file.Sync()
		}
		// Increment the number of flushes. The background syncing uses this.
		tx.db.flushes++
	}
	// Unlock the database and allow for another writable transaction.
	tx.unlock()
	// Clear the db field to disable this transaction from future use.
	tx.db = nil
	return errors.WithStack(err)
}

// Rollback closes the transaction and reverts all mutable operations that
// were performed on the transaction such as Set() and Delete().
//
// Read-only transactions can only be rolled back, not committed.
func (tx *Tx) Rollback() error {
	if tx.funcd {
		panic("managed tx rollback not allowed")
	}
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	}
	// The rollback func does the heavy lifting.
	if tx.writable {
		tx.rollbackInner()
	}
	// unlock the database for more transactions.
	tx.unlock()
	// Clear the db field to disable this transaction from future use.
	tx.db = nil
	return nil
}

type dbItem struct {
	key     []byte
	value   interface{}
	data    []byte
	keyless bool // keyless item for scanning
}

// writeSetTo writes an item as a single SET record to the a bufio Writer.
func (dbi *dbItem) writeSetTo(buf []byte) []byte {
	buf = append(buf, tagItemWriteSet)
	buf = append(buf, []byte{0, 0, 0, 0}...)
	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.key)))
	buf = append(buf, dbi.key...)

	buf = append(buf, []byte{0, 0, 0, 0}...)
	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.data)))
	buf = append(buf, dbi.data...)
	return buf
}

// writeDeleteTo writes an item as a single DEL record to the a bufio Writer.
func (dbi *dbItem) writeDeleteTo(buf []byte) []byte {
	buf = append(buf, tagItemWriteDel)
	buf = append(buf, []byte{0, 0, 0, 0}...)
	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.key)))
	buf = append(buf, dbi.key...)
	return buf
}

// MaxTime from http://stackoverflow.com/questions/25065055#32620397
// This is a long time in the future. It's an imaginary number that is
// used for b-tree ordering.
var maxTime = time.Unix(1<<63-62135596801, 999999999)

// Less determines if a b-tree item is less than another. This is required
// for ordering, inserting, and deleting items from a b-tree. It's important
// to note that the ctx parameter is used to help with determine which
// formula to use on an item. Each b-tree should use a different ctx when
// sharing the same item.
func (dbi *dbItem) Less(item btree.Item, value interface{}) bool {
	dbi2 := item.(*dbItem)
	// Always fall back to the key comparison. This creates absolute uniqueness.
	if dbi.keyless {
		return false
	} else if dbi2.keyless {
		return true
	}
	return bytes.Compare(dbi.key, dbi2.key) < 0
}

// Set inserts or replaces an item in the database based on the key.
// The opt params may be used for additional functionality such as forcing
// the item to be evicted at a specified time. When the return value
// for err is nil the operation succeeded. When the return value of
// replaced is true, then the operaton replaced an existing item whose
// value will be returned through the previousValue variable.
// The results of this operation will not be available to other
// transactions until the current transaction has successfully committed.
//
// Only a writable transaction can be used with this operation.
// This operation is not allowed during iterations such as Ascend* & Descend*.
func (tx *Tx) Set(key []byte, value interface{}, data []byte) error {
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	} else if !tx.writable {
		return errors.WithStack(ErrTxNotWritable)
	} else if tx.wc.itercount > 0 {
		return errors.WithStack(ErrTxIterating)
	}
	skey := string(key)
	item := &dbItem{key: key, value: value, data: data}
	// Insert the item into the keys tree.
	prev := tx.db.insertIntoDatabase(item)

	// insert into the rollback map if there has not been a deleteAll.
	if tx.wc.rbkeys == nil {
		if prev == nil {
			// An item with the same key did not previously exist. Let's
			// create a rollback entry with a nil value. A nil value indicates
			// that the entry should be deleted on rollback. When the value is
			// *not* nil, that means the entry should be reverted.
			tx.wc.rollbackItems[skey] = nil
		} else {
			// A previous item already exists in the database. Let's create a
			// rollback entry with the item as the value. We need to check the
			// map to see if there isn't already an item that matches the
			// same key.
			if _, ok := tx.wc.rollbackItems[skey]; !ok {
				tx.wc.rollbackItems[skey] = prev
			}
		}
	}
	// For commits we simply assign the item to the map. We use this map to
	// write the entry to disk.
	tx.wc.commitItems[skey] = item
	return nil
}

// Get returns a value for a key. If the item does not exist then ErrNotFound is returned
func (tx *Tx) Get(key []byte) (val interface{}, err error) {
	if tx.db == nil {
		return "", errors.WithStack(ErrTxClosed)
	}
	item := tx.db.get(key)
	if item == nil {
		// The item does not exists.
		return "", errors.WithStack(ErrNotFound)
	}
	return item.value, nil
}

// Delete removes an item from the database based on the item's key. If the item
// does not exist then ErrNotFound is returned.
//
// Only a writable transaction can be used for this operation.
// This operation is not allowed during iterations such as Ascend* & Descend*.
func (tx *Tx) Delete(key []byte) error {
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	} else if !tx.writable {
		return errors.WithStack(ErrTxNotWritable)
	} else if tx.wc.itercount > 0 {
		return errors.WithStack(ErrTxIterating)
	}
	item := tx.db.deleteFromDatabase(&dbItem{key: key})
	if item == nil {
		return errors.WithStack(ErrNotFound)
	}
	// create a rollback entry if there has not been a deleteAll call.
	skey := string(key)
	if tx.wc.rbkeys == nil {
		if _, ok := tx.wc.rollbackItems[skey]; !ok {
			tx.wc.rollbackItems[skey] = item
		}
	}
	tx.wc.commitItems[skey] = nil
	return nil
}

// Iterate iterates all elements has prefix
func (tx *Tx) Iterate(prefix []byte, fn func(key []byte, value interface{}) error) error {
	if len(prefix) > 0 {
		end := make([]byte, len(prefix))
		copy(end, prefix)
		for i := len(end) - 1; i >= 0; i-- {
			end[i]++
			if end[i] != 0 {
				break
			}
		}
		if bytes.Compare(prefix, end) > 0 {
			return nil
		}
		var inErr error
		tx.AscendRange(prefix, end, func(key []byte, value interface{}) bool {
			if err := fn(key, value); err != nil {
				inErr = errors.WithStack(err)
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	} else {
		var inErr error
		tx.Ascend(func(key []byte, value interface{}) bool {
			if err := fn(key, value); err != nil {
				inErr = errors.WithStack(err)
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
	}
	return nil
}

// scan iterates through a specified index and calls user-defined iterator
// function for each item encountered.
// The desc param indicates that the iterator should descend.
// The gt param indicates that there is a greaterThan limit.
// The lt param indicates that there is a lessThan limit.
// The index param tells the scanner to use the specified index tree. An
// empty string for the index means to scan the keys, not the values.
// The start and stop params are the greaterThan, lessThan limits. For
// descending order, these will be lessThan, greaterThan.
// An error will be returned if the tx is closed or the index is not found.
func (tx *Tx) scan(desc, gt, lt bool, start, stop []byte,
	iterator func(key []byte, value interface{}) bool) error {
	if tx.db == nil {
		return errors.WithStack(ErrTxClosed)
	}
	// wrap a btree specific iterator around the user-defined iterator.
	iter := func(item btree.Item) bool {
		dbi := item.(*dbItem)
		return iterator(dbi.key, dbi.value)
	}
	var tr *btree.BTree
	tr = tx.db.keys
	// create some limit items
	var itemA, itemB *dbItem
	if gt || lt {
		itemA = &dbItem{key: start}
		itemB = &dbItem{key: stop}
	}
	// execute the scan on the underlying tree.
	if tx.wc != nil {
		tx.wc.itercount++
		defer func() {
			tx.wc.itercount--
		}()
	}
	if desc {
		if gt {
			if lt {
				tr.DescendRange(itemA, itemB, iter)
			} else {
				tr.DescendGreaterThan(itemA, iter)
			}
		} else if lt {
			tr.DescendLessOrEqual(itemA, iter)
		} else {
			tr.Descend(iter)
		}
	} else {
		if gt {
			if lt {
				tr.AscendRange(itemA, itemB, iter)
			} else {
				tr.AscendGreaterOrEqual(itemA, iter)
			}
		} else if lt {
			tr.AscendLessThan(itemA, iter)
		} else {
			tr.Ascend(iter)
		}
	}
	return nil
}

// Ascend calls the iterator for every item in the database within the range
// [first, last], until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) Ascend(iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(false, false, false, nil, nil, iterator)
}

// AscendGreaterOrEqual calls the iterator for every item in the database within
// the range [pivot, last], until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) AscendGreaterOrEqual(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(false, true, false, pivot, nil, iterator)
}

// AscendLessThan calls the iterator for every item in the database within the
// range [first, pivot), until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) AscendLessThan(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(false, false, true, pivot, nil, iterator)
}

// AscendRange calls the iterator for every item in the database within
// the range [greaterOrEqual, lessThan), until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) AscendRange(greaterOrEqual, lessThan []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(
		false, true, true, greaterOrEqual, lessThan, iterator,
	)
}

// Descend calls the iterator for every item in the database within the range
// [last, first], until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) Descend(iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(true, false, false, nil, nil, iterator)
}

// DescendGreaterThan calls the iterator for every item in the database within
// the range [last, pivot), until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) DescendGreaterThan(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(true, true, false, pivot, nil, iterator)
}

// DescendLessOrEqual calls the iterator for every item in the database within
// the range [pivot, first], until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) DescendLessOrEqual(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(true, false, true, pivot, nil, iterator)
}

// DescendRange calls the iterator for every item in the database within
// the range [lessOrEqual, greaterThan), until iterator returns false.
// The results will be ordered by the item key.
func (tx *Tx) DescendRange(lessOrEqual, greaterThan []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.scan(
		true, true, true, lessOrEqual, greaterThan, iterator,
	)
}

// AscendEqual calls the iterator for every item in the database that equals
// pivot, until iterator returns false.
func (tx *Tx) AscendEqual(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.AscendGreaterOrEqual(pivot, func(key []byte, value interface{}) bool {
		if bytes.Compare(key, pivot) != 0 {
			return false
		}
		return iterator(key, value)
	})
}

// DescendEqual calls the iterator for every item in the database that equals
// pivot, until iterator returns false.
func (tx *Tx) DescendEqual(pivot []byte,
	iterator func(key []byte, value interface{}) bool) error {
	return tx.DescendLessOrEqual(pivot, func(key []byte, value interface{}) bool {
		if bytes.Compare(key, pivot) != 0 {
			return false
		}
		return iterator(key, value)
	})
}

// Len returns the number of items in the database
func (tx *Tx) Len() (int, error) {
	if tx.db == nil {
		return 0, errors.WithStack(ErrTxClosed)
	}
	return tx.db.keys.Len(), nil
}
