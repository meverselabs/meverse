// Package keydb came from buntdb that implements a low-level in-memory key/value store in pure Go.
// It persists to disk, is ACID compliant, and uses locking for multiple
// readers and a single writer. Bunt is ideal for projects that need
// a dependable database, and favor speed over data size.
package keydb2

// import (
// 	"bytes"
// 	"strconv"

// 	"github.com/meverselabs/meverse/common/bin"
// 	"github.com/pkg/errors"
// 	"github.com/tidwall/btree"
// 	"github.com/tidwall/gjson"
// 	"github.com/tidwall/grect"
// 	"github.com/tidwall/match"
// )

// // Tx represents a transaction on the database. This transaction can either be
// // read-only or read/write. Read-only transactions can be used for retrieving
// // values for keys and iterating through keys and values. Read/write
// // transactions can set and delete keys.
// //
// // All transactions must be committed or rolled-back when done.
// type Tx struct {
// 	db       *DB             // the underlying database.
// 	writable bool            // when false mutable operations fail.
// 	funcd    bool            // when true Commit and Rollback panic.
// 	wc       *txWriteContext // context for writable transactions.
// }

// type keydbKey []byte

// type txWriteContext struct {
// 	// rollback when deleteAll is called
// 	rbkeys *btree.BTree // a tree of all item ordered by key

// 	rollbackItems map[string]*dbItem // details for rolling back tx.
// 	commitItems   map[string]*dbItem // details for committing tx.
// 	itercount     int                // stack of iterators
// }

// // DeleteAll deletes all items from the database.
// func (tx *Tx) DeleteAll() error {
// 	if tx.db == nil {
// 		return ErrTxClosed
// 	} else if !tx.writable {
// 		return ErrTxNotWritable
// 	} else if tx.wc.itercount > 0 {
// 		return ErrTxIterating
// 	}

// 	// check to see if we've already deleted everything
// 	if tx.wc.rbkeys == nil {
// 		// we need to backup the live data in case of a rollback.
// 		tx.wc.rbkeys = tx.db.keys
// 	}

// 	// now reset the live database trees
// 	tx.db.keys = btreeNew(lessCtx())

// 	// always clear out the commits
// 	tx.wc.commitItems = make(map[string]*dbItem)

// 	return nil
// }

// // Begin opens a new transaction.
// // Multiple read-only transactions can be opened at the same time but there can
// // only be one read/write transaction at a time. Attempting to open a read/write
// // transactions while another one is in progress will result in blocking until
// // the current read/write transaction is completed.
// //
// // All transactions must be closed by calling Commit() or Rollback() when done.
// func (db *DB) Begin(writable bool) (*Tx, error) {
// 	tx := &Tx{
// 		db:       db,
// 		writable: writable,
// 	}
// 	tx.lock()
// 	if db.closed {
// 		tx.unlock()
// 		return nil, ErrDatabaseClosed
// 	}
// 	if writable {
// 		// writable transactions have a writeContext object that
// 		// contains information about changes to the database.
// 		tx.wc = &txWriteContext{}
// 		tx.wc.rollbackItems = make(map[string]*dbItem)
// 		if db.persist {
// 			tx.wc.commitItems = make(map[string]*dbItem)
// 		}
// 	}
// 	return tx, nil
// }

// // lock locks the database based on the transaction type.
// func (tx *Tx) lock() {
// 	if tx.writable {
// 		tx.db.mu.Lock()
// 	} else {
// 		tx.db.mu.RLock()
// 	}
// }

// // unlock unlocks the database based on the transaction type.
// func (tx *Tx) unlock() {
// 	if tx.writable {
// 		tx.db.mu.Unlock()
// 	} else {
// 		tx.db.mu.RUnlock()
// 	}
// }

// // rollbackInner handles the underlying rollback logic.
// // Intended to be called from Commit() and Rollback().
// func (tx *Tx) rollbackInner() {
// 	// rollback the deleteAll if needed
// 	if tx.wc.rbkeys != nil {
// 		tx.db.keys = tx.wc.rbkeys
// 	}
// 	for key, item := range tx.wc.rollbackItems {
// 		tx.db.deleteFromDatabase(&dbItem{key: []byte(key)})
// 		if item != nil {
// 			// When an item is not nil, we will need to reinsert that item
// 			// into the database overwriting the current one.
// 			tx.db.insertIntoDatabase(item)
// 		}
// 	}
// }

// // Commit writes all changes to disk.
// // An error is returned when a write error occurs, or when a Commit() is called
// // from a read-only transaction.
// func (tx *Tx) Commit() error {
// 	if tx.funcd {
// 		panic("managed tx commit not allowed")
// 	}
// 	if tx.db == nil {
// 		return ErrTxClosed
// 	} else if !tx.writable {
// 		return ErrTxNotWritable
// 	}
// 	var err error
// 	if tx.db.persist && (len(tx.wc.commitItems) > 0 || tx.wc.rbkeys != nil) {
// 		tx.db.buf = tx.db.buf[:0]
// 		// write a flushdb if a deleteAll was called.
// 		if tx.wc.rbkeys != nil {
// 			tx.db.buf = append(tx.db.buf, tagFlushDB)
// 		}
// 		// Each committed record is written to disk
// 		for key, item := range tx.wc.commitItems {
// 			if item == nil {
// 				tx.db.buf = (&dbItem{key: []byte(key)}).writeDeleteTo(tx.db.buf)
// 			} else {
// 				tx.db.buf = item.writeSetTo(tx.db.buf)
// 			}
// 		}
// 		tx.db.buf = append(tx.db.buf, tagTxEnd)
// 		// Flushing the buffer only once per transaction.
// 		// If this operation fails then the write did failed and we must
// 		// rollback.
// 		if _, err := tx.db.file.Seek(0, 2); err != nil {
// 			tx.rollbackInner()
// 		}
// 		if _, err := tx.db.file.Write(tx.db.buf); err != nil {
// 			tx.rollbackInner()
// 		}
// 		if tx.db.config.SyncPolicy == Always {
// 			_ = tx.db.file.Sync()
// 		}
// 		// Increment the number of flushes. The background syncing uses this.
// 		tx.db.flushes++
// 	}
// 	// Unlock the database and allow for another writable transaction.
// 	tx.unlock()
// 	// Clear the db field to disable this transaction from future use.
// 	tx.db = nil
// 	return err
// }

// // Rollback closes the transaction and reverts all mutable operations that
// // were performed on the transaction such as Set() and Delete().
// //
// // Read-only transactions can only be rolled back, not committed.
// func (tx *Tx) Rollback() error {
// 	if tx.funcd {
// 		panic("managed tx rollback not allowed")
// 	}
// 	if tx.db == nil {
// 		return ErrTxClosed
// 	}
// 	// The rollback func does the heavy lifting.
// 	if tx.writable {
// 		tx.rollbackInner()
// 	}
// 	// unlock the database for more transactions.
// 	tx.unlock()
// 	// Clear the db field to disable this transaction from future use.
// 	tx.db = nil
// 	return nil
// }

// type dbItem struct {
// 	key, data []byte
// 	value     interface{} // the binary key and value
// 	keyless   bool        // keyless item for scanning
// }

// // estIntSize returns the string representions size.
// // Has the same result as len(strconv.Itoa(x)).
// func estIntSize(x int) int {
// 	n := 1
// 	if x < 0 {
// 		n++
// 		x *= -1
// 	}
// 	for x >= 10 {
// 		n++
// 		x /= 10
// 	}
// 	return n
// }

// func estArraySize(count int) int {
// 	return 1 + estIntSize(count) + 2
// }

// func estBulkStringSize(s []byte) int {
// 	return 1 + estIntSize(len(s)) + 2 + len(s) + 2
// }

// // estAOFSetSize returns an estimated number of bytes that this item will use
// // when stored in the aof file.
// func (dbi *dbItem) estAOFSetSize() int {
// 	var n int
// 	n += estArraySize(3)
// 	n += estBulkStringSize([]byte("set"))
// 	n += estBulkStringSize(dbi.key)
// 	n += estBulkStringSize(dbi.data)
// 	return n
// }

// func appendArray(buf []byte, count int) []byte {
// 	buf = append(buf, '*')
// 	buf = strconv.AppendInt(buf, int64(count), 10)
// 	buf = append(buf, '\r', '\n')
// 	return buf
// }

// // writeSetTo writes an item as a single SET record to the a bufio Writer.
// func (dbi *dbItem) writeSetTo(buf []byte) []byte {
// 	buf = append(buf, tagItemWriteSet)
// 	buf = append(buf, []byte{0, 0, 0, 0}...)
// 	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.key)))
// 	buf = append(buf, dbi.key...)

// 	buf = append(buf, []byte{0, 0, 0, 0}...)
// 	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.data)))
// 	buf = append(buf, dbi.data...)
// 	return buf
// }

// // writeDeleteTo writes an item as a single DEL record to the a bufio Writer.
// func (dbi *dbItem) writeDeleteTo(buf []byte) []byte {
// 	buf = append(buf, tagItemWriteDel)
// 	buf = append(buf, []byte{0, 0, 0, 0}...)
// 	bin.PutUint32(buf[len(buf)-4:], uint32(len(dbi.key)))
// 	buf = append(buf, dbi.key...)
// 	return buf
// }

// // Less determines if a b-tree item is less than another. This is required
// // for ordering, inserting, and deleting items from a b-tree. It's important
// // to note that the ctx parameter is used to help with determine which
// // formula to use on an item. Each b-tree should use a different ctx when
// // sharing the same item.
// func (dbi *dbItem) Less(dbi2 *dbItem) bool {
// 	// Always fall back to the key comparison. This creates absolute uniqueness.
// 	if dbi.keyless {
// 		return false
// 	} else if dbi2.keyless {
// 		return true
// 	}
// 	return bytes.Compare(dbi.key, dbi2.key) < 0
// }

// func lessCtx() func(a, b interface{}) bool {
// 	return func(a, b interface{}) bool {
// 		return a.(*dbItem).Less(b.(*dbItem))
// 	}
// }

// // Set inserts or replaces an item in the database based on the key.
// // The opt params may be used for additional functionality such as forcing
// // the item to be evicted at a specified time. When the return value
// // for err is nil the operation succeeded. When the return value of
// // replaced is true, then the operaton replaced an existing item whose
// // value will be returned through the previousValue variable.
// // The results of this operation will not be available to other
// // transactions until the current transaction has successfully committed.
// //
// // Only a writable transaction can be used with this operation.
// // This operation is not allowed during iterations such as Ascend* & Descend*.
// func (tx *Tx) Set(key, value []byte) (err error) {
// 	if tx.db == nil {
// 		return ErrTxClosed
// 	} else if !tx.writable {
// 		return ErrTxNotWritable
// 	} else if tx.wc.itercount > 0 {
// 		return ErrTxIterating
// 	}
// 	item := &dbItem{key: key, data: value}
// 	// Insert the item into the keys tree.
// 	prev := tx.db.insertIntoDatabase(item)

// 	skey := string(key)

// 	// insert into the rollback map if there has not been a deleteAll.
// 	if tx.wc.rbkeys == nil {
// 		if prev == nil {
// 			// An item with the same key did not previously exist. Let's
// 			// create a rollback entry with a nil value. A nil value indicates
// 			// that the entry should be deleted on rollback. When the value is
// 			// *not* nil, that means the entry should be reverted.
// 			if _, ok := tx.wc.rollbackItems[skey]; !ok {
// 				tx.wc.rollbackItems[skey] = nil
// 			}
// 		} else {
// 			// A previous item already exists in the database. Let's create a
// 			// rollback entry with the item as the value. We need to check the
// 			// map to see if there isn't already an item that matches the
// 			// same key.
// 			if _, ok := tx.wc.rollbackItems[skey]; !ok {
// 				tx.wc.rollbackItems[skey] = prev
// 			}
// 		}
// 	}
// 	// For commits we simply assign the item to the map. We use this map to
// 	// write the entry to disk.
// 	if tx.db.persist {
// 		tx.wc.commitItems[skey] = item
// 	}
// 	return nil
// }

// // Get returns a value for a key. If the item does not exist or if the item
// // has expired then ErrNotFound is returned. If ignoreExpired is true, then
// // the found value will be returned even if it is expired.
// func (tx *Tx) Get(key []byte) (val interface{}, err error) {
// 	if tx.db == nil {
// 		return nil, ErrTxClosed
// 	}
// 	item := tx.db.get(key)
// 	if item == nil {
// 		// The item does not exists or has expired. Let's assume that
// 		// the caller is only interested in items that have not expired.
// 		return nil, ErrNotFound
// 	}
// 	return tx.db.unmarshaler(key, item.data)
// 	// return item.val, nil
// }

// // Delete removes an item from the database based on the item's key. If the item
// // does not exist or if the item has expired then ErrNotFound is returned.
// //
// // Only a writable transaction can be used for this operation.
// // This operation is not allowed during iterations such as Ascend* & Descend*.
// func (tx *Tx) Delete(key []byte) (val interface{}, err error) {
// 	if tx.db == nil {
// 		return nil, ErrTxClosed
// 	} else if !tx.writable {
// 		return nil, ErrTxNotWritable
// 	} else if tx.wc.itercount > 0 {
// 		return nil, ErrTxIterating
// 	}
// 	item := tx.db.deleteFromDatabase(&dbItem{key: key})
// 	if item == nil {
// 		return nil, ErrNotFound
// 	}
// 	skey := string(key)
// 	// create a rollback entry if there has not been a deleteAll call.
// 	if tx.wc.rbkeys == nil {
// 		if _, ok := tx.wc.rollbackItems[skey]; !ok {
// 			tx.wc.rollbackItems[skey] = item
// 		}
// 	}
// 	if tx.db.persist {
// 		tx.wc.commitItems[skey] = nil
// 	}
// 	return item.value, nil
// }

// // scan iterates through a specified index and calls user-defined iterator
// // function for each item encountered.
// // The desc param indicates that the iterator should descend.
// // The gt param indicates that there is a greaterThan limit.
// // The lt param indicates that there is a lessThan limit.
// // The index param tells the scanner to use the specified index tree. An
// // empty string for the index means to scan the keys, not the values.
// // The start and stop params are the greaterThan, lessThan limits. For
// // descending order, these will be lessThan, greaterThan.
// // An error will be returned if the tx is closed or the index is not found.
// func (tx *Tx) scan(desc, gt, lt bool, start, stop interface{}, iterator func(key []byte, data []byte) bool) error {
// 	if tx.db == nil {
// 		return ErrTxClosed
// 	}
// 	// wrap a btree specific iterator around the user-defined iterator.
// 	iter := func(item interface{}) bool {
// 		dbi := item.(*dbItem)
// 		return iterator(dbi.key, dbi.data)
// 	}
// 	var tr *btree.BTree
// 	tr = tx.db.keys
// 	// create some limit items
// 	var itemA, itemB *dbItem
// 	if gt || lt {
// 		itemA = &dbItem{key: start.([]byte)}
// 		itemB = &dbItem{key: stop.([]byte)}
// 	}
// 	// execute the scan on the underlying tree.
// 	if tx.wc != nil {
// 		tx.wc.itercount++
// 		defer func() {
// 			tx.wc.itercount--
// 		}()
// 	}
// 	if desc {
// 		if gt {
// 			if lt {
// 				btreeDescendRange(tr, itemA, itemB, iter)
// 			} else {
// 				btreeDescendGreaterThan(tr, itemA, iter)
// 			}
// 		} else if lt {
// 			btreeDescendLessOrEqual(tr, itemA, iter)
// 		} else {
// 			btreeDescend(tr, iter)
// 		}
// 	} else {
// 		if gt {
// 			if lt {
// 				btreeAscendRange(tr, itemA, itemB, iter)
// 			} else {
// 				btreeAscendGreaterOrEqual(tr, itemA, iter)
// 			}
// 		} else if lt {
// 			btreeAscendLessThan(tr, itemA, iter)
// 		} else {
// 			btreeAscend(tr, iter)
// 		}
// 	}
// 	return nil
// }

// // Match returns true if the specified key matches the pattern. This is a very
// // simple pattern matcher where '*' matches on any number characters and '?'
// // matches on any one character.
// func Match(key, pattern string) bool {
// 	return match.Match(key, pattern)
// }

// // AscendKeys allows for iterating through keys based on the specified pattern.
// // func (tx *Tx) AscendKeys(pattern []byte,
// // 	iterator func(key, value []byte) bool) error {
// // 	if len(pattern) == 0 {
// // 		return nil
// // 	}
// // 	if pattern[0] == '*' {
// // 		if string(pattern) == "*" {
// // 			return tx.Ascend(nil, iterator)
// // 		}
// // 		return tx.Ascend(nil, func(key, value []byte) bool {
// // 			if match.Match(string(key), string(pattern)) {
// // 				if !iterator(key, value) {
// // 					return false
// // 				}
// // 			}
// // 			return true
// // 		})
// // 	}
// // 	min, max := match.Allowable(string(pattern))
// // 	return tx.AscendGreaterOrEqual(nil, []byte(min), func(key, value []byte) bool {
// // 		if key > max {
// // 			return false
// // 		}
// // 		if match.Match(key, pattern) {
// // 			if !iterator(key, value) {
// // 				return false
// // 			}
// // 		}
// // 		return true
// // 	})
// // }

// // DescendKeys allows for iterating through keys based on the specified pattern.
// // func (tx *Tx) DescendKeys(pattern string,
// // 	iterator func(key, value string) bool) error {
// // 	if pattern == "" {
// // 		return nil
// // 	}
// // 	if pattern[0] == '*' {
// // 		if pattern == "*" {
// // 			return tx.Descend("", iterator)
// // 		}
// // 		return tx.Descend("", func(key, value string) bool {
// // 			if match.Match(key, pattern) {
// // 				if !iterator(key, value) {
// // 					return false
// // 				}
// // 			}
// // 			return true
// // 		})
// // 	}
// // 	min, max := match.Allowable(pattern)
// // 	return tx.DescendLessOrEqual("", max, func(key, value string) bool {
// // 		if key < min {
// // 			return false
// // 		}
// // 		if match.Match(key, pattern) {
// // 			if !iterator(key, value) {
// // 				return false
// // 			}
// // 		}
// // 		return true
// // 	})
// // }

// // Iterate iterates all elements has prefix
// func (tx *Tx) Iterate(prefix []byte, fn func(key []byte, value interface{}) error) error {
// 	if len(prefix) > 0 {
// 		end := make([]byte, len(prefix))
// 		copy(end, prefix)
// 		for i := len(end) - 1; i >= 0; i-- {
// 			end[i]++
// 			if end[i] != 0 {
// 				break
// 			}
// 		}
// 		if bytes.Compare(prefix, end) > 0 {
// 			return nil
// 		}
// 		var inErr error
// 		tx.AscendRange(prefix, end, func(key []byte, data []byte) bool {
// 			value, err := tx.db.unmarshaler(key, data)
// 			if err != nil {
// 				inErr = err
// 				return false
// 			}
// 			if err := fn(key, value); err != nil {
// 				inErr = err
// 				return false
// 			}
// 			return true
// 		})
// 		if inErr != nil {
// 			return errors.WithStack(inErr)
// 		}
// 	} else {
// 		var inErr error
// 		tx.Ascend(func(key []byte, data []byte) bool {
// 			value, err := tx.db.unmarshaler(key, data)
// 			if err != nil {
// 				inErr = err
// 				return false
// 			}
// 			if err := fn(key, value); err != nil {
// 				inErr = err
// 				return false
// 			}
// 			return true
// 		})
// 		if inErr != nil {
// 			return errors.WithStack(inErr)
// 		}
// 	}
// 	return nil
// }

// // AscendRange calls the iterator for every item in the database within
// // the range [greaterOrEqual, lessThan), until iterator returns false.
// // When an index is provided, the results will be ordered by the item values
// // as specified by the less() function of the defined index.
// // When an index is not provided, the results will be ordered by the item key.
// // An invalid index will return an error.
// func (tx *Tx) AscendRange(greaterOrEqual, lessThan []byte,
// 	iterator func(key []byte, value []byte) bool) error {
// 	return tx.scan(
// 		false, true, true, greaterOrEqual, lessThan, iterator,
// 	)
// }

// // Ascend calls the iterator for every item in the database within the range
// // [first, last], until iterator returns false.
// // When an index is provided, the results will be ordered by the item values
// // as specified by the less() function of the defined index.
// // When an index is not provided, the results will be ordered by the item key.
// // An invalid index will return an error.
// func (tx *Tx) Ascend(iterator func(key []byte, value []byte) bool) error {
// 	return tx.scan(false, false, false, nil, nil, iterator)
// }

// // Len returns the number of items in the database
// func (tx *Tx) Len() (int, error) {
// 	if tx.db == nil {
// 		return 0, ErrTxClosed
// 	}
// 	return tx.db.keys.Len(), nil
// }

// // Rect is helper function that returns a string representation
// // of a rect. IndexRect() is the reverse function and can be used
// // to generate a rect from a string.
// func Rect(min, max []float64) string {
// 	r := grect.Rect{Min: min, Max: max}
// 	return r.String()
// }

// // Point is a helper function that converts a series of float64s
// // to a rectangle for a spatial index.
// func Point(coords ...float64) string {
// 	return Rect(coords, coords)
// }

// // IndexRect is a helper function that converts string to a rect.
// // Rect() is the reverse function and can be used to generate a string
// // from a rect.
// func IndexRect(a string) (min, max []float64) {
// 	r := grect.Get(a)
// 	return r.Min, r.Max
// }

// // IndexString is a helper function that return true if 'a' is less than 'b'.
// // This is a case-insensitive comparison. Use the IndexBinary() for comparing
// // case-sensitive strings.
// func IndexString(a, b string) bool {
// 	for i := 0; i < len(a) && i < len(b); i++ {
// 		if a[i] >= 'A' && a[i] <= 'Z' {
// 			if b[i] >= 'A' && b[i] <= 'Z' {
// 				// both are uppercase, do nothing
// 				if a[i] < b[i] {
// 					return true
// 				} else if a[i] > b[i] {
// 					return false
// 				}
// 			} else {
// 				// a is uppercase, convert a to lowercase
// 				if a[i]+32 < b[i] {
// 					return true
// 				} else if a[i]+32 > b[i] {
// 					return false
// 				}
// 			}
// 		} else if b[i] >= 'A' && b[i] <= 'Z' {
// 			// b is uppercase, convert b to lowercase
// 			if a[i] < b[i]+32 {
// 				return true
// 			} else if a[i] > b[i]+32 {
// 				return false
// 			}
// 		} else {
// 			// neither are uppercase
// 			if a[i] < b[i] {
// 				return true
// 			} else if a[i] > b[i] {
// 				return false
// 			}
// 		}
// 	}
// 	return len(a) < len(b)
// }

// // IndexBinary is a helper function that returns true if 'a' is less than 'b'.
// // This compares the raw binary of the string.
// func IndexBinary(a, b string) bool {
// 	return a < b
// }

// // IndexInt is a helper function that returns true if 'a' is less than 'b'.
// func IndexInt(a, b string) bool {
// 	ia, _ := strconv.ParseInt(a, 10, 64)
// 	ib, _ := strconv.ParseInt(b, 10, 64)
// 	return ia < ib
// }

// // IndexUint is a helper function that returns true if 'a' is less than 'b'.
// // This compares uint64s that are added to the database using the
// // Uint() conversion function.
// func IndexUint(a, b string) bool {
// 	ia, _ := strconv.ParseUint(a, 10, 64)
// 	ib, _ := strconv.ParseUint(b, 10, 64)
// 	return ia < ib
// }

// // IndexFloat is a helper function that returns true if 'a' is less than 'b'.
// // This compares float64s that are added to the database using the
// // Float() conversion function.
// func IndexFloat(a, b string) bool {
// 	ia, _ := strconv.ParseFloat(a, 64)
// 	ib, _ := strconv.ParseFloat(b, 64)
// 	return ia < ib
// }

// // IndexJSON provides for the ability to create an index on any JSON field.
// // When the field is a string, the comparison will be case-insensitive.
// // It returns a helper function used by CreateIndex.
// func IndexJSON(path string) func(a, b string) bool {
// 	return func(a, b string) bool {
// 		return gjson.Get(a, path).Less(gjson.Get(b, path), false)
// 	}
// }

// // IndexJSONCaseSensitive provides for the ability to create an index on
// // any JSON field.
// // When the field is a string, the comparison will be case-sensitive.
// // It returns a helper function used by CreateIndex.
// func IndexJSONCaseSensitive(path string) func(a, b string) bool {
// 	return func(a, b string) bool {
// 		return gjson.Get(a, path).Less(gjson.Get(b, path), true)
// 	}
// }

// // Desc is a helper function that changes the order of an index.
// func Desc(less func(a, b string) bool) func(a, b string) bool {
// 	return func(a, b string) bool { return less(b, a) }
// }

// //// Wrappers around btree Ascend/Descend

// func bLT(tr *btree.BTree, a, b interface{}) bool { return tr.Less(a, b) }
// func bGT(tr *btree.BTree, a, b interface{}) bool { return tr.Less(b, a) }

// // func bLTE(tr *btree.BTree, a, b interface{}) bool { return !tr.Less(b, a) }
// // func bGTE(tr *btree.BTree, a, b interface{}) bool { return !tr.Less(a, b) }

// // Ascend

// func btreeAscend(tr *btree.BTree, iter func(item interface{}) bool) {
// 	tr.Ascend(nil, iter)
// }

// func btreeAscendLessThan(tr *btree.BTree, pivot interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Ascend(nil, func(item interface{}) bool {
// 		return bLT(tr, item, pivot) && iter(item)
// 	})
// }

// func btreeAscendGreaterOrEqual(tr *btree.BTree, pivot interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Ascend(pivot, iter)
// }

// func btreeAscendRange(tr *btree.BTree, greaterOrEqual, lessThan interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Ascend(greaterOrEqual, func(item interface{}) bool {
// 		return bLT(tr, item, lessThan) && iter(item)
// 	})
// }

// // Descend

// func btreeDescend(tr *btree.BTree, iter func(item interface{}) bool) {
// 	tr.Descend(nil, iter)
// }

// func btreeDescendGreaterThan(tr *btree.BTree, pivot interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Descend(nil, func(item interface{}) bool {
// 		return bGT(tr, item, pivot) && iter(item)
// 	})
// }

// func btreeDescendRange(tr *btree.BTree, lessOrEqual, greaterThan interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Descend(lessOrEqual, func(item interface{}) bool {
// 		return bGT(tr, item, greaterThan) && iter(item)
// 	})
// }

// func btreeDescendLessOrEqual(tr *btree.BTree, pivot interface{},
// 	iter func(item interface{}) bool,
// ) {
// 	tr.Descend(pivot, iter)
// }

// func btreeNew(less func(a, b interface{}) bool) *btree.BTree {
// 	// Using NewNonConcurrent because we're managing our own locks.
// 	return btree.NewNonConcurrent(less)
// }
