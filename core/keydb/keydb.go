// Package keydb came from buntdb that implements a low-level in-memory key/value store in pure Go.
// It persists to disk, is ACID compliant, and uses locking for multiple
// readers and a single writer. Bunt is ideal for projects that need
// a dependable database, and favor speed over data size.
package keydb

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/pkg/errors"
	"github.com/tidwall/btree"
)

// Unmarshaler needs to unmarshal structures
type Unmarshaler func(key []byte, data []byte) (interface{}, error)

// DB represents a collection of key-value pairs that persist on disk.
// Transactions are used for all forms of data access to the DB.
type DB struct {
	mu          sync.RWMutex // the gatekeeper for all fields
	file        *os.File     // the underlying file
	buf         []byte       // a buffer to write to
	keys        *btree.BTree // a tree of all item ordered by key
	flushes     int          // a count of the number of disk flushes
	closed      bool         // set when the database has been closed
	config      Config       // the database configuration
	shrinking   bool         // when an aof shrink is in-process.
	lastaofsz   int          // the size of the last shrink aof size
	unmarshaler Unmarshaler
}

// Open opens a database at the provided path.
// If the file does not exist then it will be created automatically.
func Open(path string, fn Unmarshaler) (*DB, error) {
	os.MkdirAll(filepath.Dir(path), os.ModeDir)

	if fn == nil {
		return nil, ErrNotExistUnmarshaler
	}

	db := &DB{
		keys:        btree.New(btreeDegrees, nil),
		unmarshaler: fn,
		config: Config{
			SyncPolicy:           EverySecond,
			AutoShrinkPercentage: 100,
			AutoShrinkMinSize:    32 * 1024 * 1024,
		},
	}
	// initialize trees and indexes
	// initialize default configuration
	// hardcoding 0666 as the default mode.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	db.file = file

	// load the database from disk
	if err := db.load(); err != nil {
		// close on error, ignore close error
		_ = db.file.Close()
		return nil, err
	}
	// start the background manager.
	go db.backgroundManager()
	return db, nil
}

// Close releases all database resources.
// All transactions must be closed before closing the database.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrDatabaseClosed
	}
	db.closed = true
	db.file.Sync() // do a sync but ignore the error
	if err := db.file.Close(); err != nil {
		return err
	}
	// Let's release all references to nil. This will help both with debugging
	// late usage panics and it provides a hint to the garbage collector
	db.keys = nil
	db.file = nil
	return nil
}

// ReadConfig returns the database configuration.
func (db *DB) ReadConfig(config *Config) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return ErrDatabaseClosed
	}
	*config = db.config
	return nil
}

// SetConfig updates the database configuration.
func (db *DB) SetConfig(config Config) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrDatabaseClosed
	}
	switch config.SyncPolicy {
	default:
		return ErrInvalidSyncPolicy
	case Never, EverySecond, Always:
	}
	db.config = config
	return nil
}

// insertIntoDatabase performs inserts an item in to the database and updates
// all indexes. If a previous item with the same key already exists, that item
// will be replaced with the new one, and return the previous item.
func (db *DB) insertIntoDatabase(item *dbItem) *dbItem {
	var pdbi *dbItem
	prev := db.keys.ReplaceOrInsert(item)
	if prev != nil {
		pdbi = prev.(*dbItem)
	}
	return pdbi
}

// deleteFromDatabase removes and item from the database and indexes. The input
// item must only have the key field specified thus "&dbItem{key: key}" is all
// that is needed to fully remove the item with the matching key. If an item
// with the matching key was found in the database, it will be removed and
// returned to the caller. A nil return value means that the item was not
// found in the database
func (db *DB) deleteFromDatabase(item *dbItem) *dbItem {
	var pdbi *dbItem
	prev := db.keys.Delete(item)
	if prev != nil {
		pdbi = prev.(*dbItem)
	}
	return pdbi
}

// backgroundManager runs continuously in the background and performs various
func (db *DB) backgroundManager() {
	defer recover()

	flushes := 0
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for range t.C {
		var shrink bool
		// Open a standard view. This will take a full lock of the
		// database thus allowing for access to anything we need.
		if err := db.Update(func(tx *Tx) error {
			if !db.config.AutoShrinkDisabled {
				pos, err := db.file.Seek(0, 1)
				if err != nil {
					return err
				}
				aofsz := int(pos)
				if aofsz > db.config.AutoShrinkMinSize {
					prc := float64(db.config.AutoShrinkPercentage) / 100.0
					shrink = aofsz > db.lastaofsz+int(float64(db.lastaofsz)*prc)
				}
			}
			return nil
		}); errors.Cause(err) == ErrDatabaseClosed {
			break
		}

		// execute a disk sync, if needed
		func() {
			db.mu.Lock()
			defer db.mu.Unlock()
			if db.config.SyncPolicy == EverySecond &&
				flushes != db.flushes {
				_ = db.file.Sync()
				flushes = db.flushes
			}
		}()
		if shrink {
			if err := db.Shrink(); err != nil {
				if errors.Cause(err) == ErrDatabaseClosed {
					break
				}
			}
		}
	}
}

// Shrink will make the database file smaller by removing redundant
// log entries. This operation does not block the database.
func (db *DB) Shrink() error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrDatabaseClosed
	}
	if db.shrinking {
		// The database is already in the process of shrinking.
		db.mu.Unlock()
		return ErrShrinkInProcess
	}
	db.shrinking = true
	defer func() {
		db.mu.Lock()
		db.shrinking = false
		db.mu.Unlock()
	}()
	fname := db.file.Name()
	tmpname := fname + ".tmp"
	// the endpos is used to return to the end of the file when we are
	// finished writing all of the current items.
	endpos, err := db.file.Seek(0, 2)
	if err != nil {
		return err
	}
	db.mu.Unlock()
	time.Sleep(time.Second / 4) // wait just a bit before starting
	f, err := os.Create(tmpname)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.RemoveAll(tmpname)
	}()

	// we are going to read items in as chunks as to not hold up the database
	// for too long.
	var buf []byte
	var pivot []byte
	done := false
	for !done {
		err := func() error {
			db.mu.RLock()
			defer db.mu.RUnlock()
			if db.closed {
				return ErrDatabaseClosed
			}
			done = true
			var n int
			db.keys.AscendGreaterOrEqual(&dbItem{key: pivot},
				func(item btree.Item) bool {
					dbi := item.(*dbItem)
					// 1000 items or 64MB buffer
					if n > 1000 || len(buf) > 64*1024*1024 {
						pivot = dbi.key
						done = false
						return false
					}
					buf = dbi.writeSetTo(buf)
					n++
					return true
				},
			)
			if len(buf) > 0 {
				buf = append(buf, tagTxEnd)
				if _, err := f.Write(buf); err != nil {
					return err
				}
				buf = buf[:0]
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	// We reached this far so all of the items have been written to a new tmp
	// There's some more work to do by appending the new line from the aof
	// to the tmp file and finally swap the files out.
	return func() error {
		// We're wrapping this in a function to get the benefit of a defered
		// lock/unlock.
		db.mu.Lock()
		defer db.mu.Unlock()
		if db.closed {
			return ErrDatabaseClosed
		}
		// We are going to open a new version of the aof file so that we do
		// not change the seek position of the previous. This may cause a
		// problem in the future if we choose to use syscall file locking.
		aof, err := os.Open(fname)
		if err != nil {
			return err
		}
		defer func() { _ = aof.Close() }()
		if _, err := aof.Seek(endpos, 0); err != nil {
			return err
		}
		// Just copy all of the new commands that have occurred since we
		// started the shrink process.
		if _, err := io.Copy(f, aof); err != nil {
			return err
		}
		// Close all files
		if err := aof.Close(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		if err := db.file.Close(); err != nil {
			return err
		}
		// Any failures below here is really bad. So just panic.
		if err := os.Rename(tmpname, fname); err != nil {
			panic(err)
		}
		db.file, err = os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		pos, err := db.file.Seek(0, 2)
		if err != nil {
			return err
		}
		db.lastaofsz = int(pos)
		return nil
	}()
}

type commitItem struct {
	Tag   byte
	Key   []byte
	Value interface{}
	Data  []byte
}

func (db *DB) readFill(r *bufio.Reader, bs []byte) (int, error) {
	read := 0
	for read < len(bs) {
		if n, err := r.Read(bs[read:]); err != nil {
			return read + n, err
		} else {
			read += n
		}
	}
	return read, nil
}

// readLoad reads from the reader and loads commands into the database.
// modTime is the modified time of the reader, should be no greater than
// the current time.Now().
func (db *DB) readLoad(file *os.File, modTime time.Time) error {
	commiteds := make([]*commitItem, 0)
	var fileOffset int64
	var lastTxPos int64
	count := 0
	r := bufio.NewReader(file)
	for {
		// read a tag
		tag, err := r.ReadByte()
		fileOffset++
		count++
		if err != nil {
			if errors.Cause(err) != io.EOF {
				if err := file.Truncate(lastTxPos); err != nil {
					return err
				}
			}
			if errors.Cause(err) == io.EOF {
				break
			}
			return err
		}
		switch tag {
		case tagItemWriteSet:
			var key []byte
			var value interface{}
			var data []byte
			if true {
				bs := make([]byte, 4)
				if n, err := db.readFill(r, bs); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
				keylen := bin.Uint32(bs)
				key = make([]byte, keylen)
				if n, err := db.readFill(r, key); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
			}
			if true {
				bs := make([]byte, 4)
				if n, err := db.readFill(r, bs); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
				valuelen := bin.Uint32(bs)
				data = make([]byte, valuelen)
				if n, err := db.readFill(r, data); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
				v, err := db.unmarshaler(key, data)
				if err != nil {
					return err
				}
				value = v
			}
			commiteds = append(commiteds, &commitItem{
				Tag:   tag,
				Key:   key,
				Value: value,
				Data:  data,
			})
		case tagItemWriteDel:
			var key []byte
			if true {
				bs := make([]byte, 4)
				if n, err := db.readFill(r, bs); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
				keylen := bin.Uint32(bs)
				key = make([]byte, keylen)
				if n, err := db.readFill(r, key); err != nil {
					if err := file.Truncate(lastTxPos); err != nil {
						return err
					}
					return nil
				} else {
					fileOffset += int64(n)
				}
			}
			commiteds = append(commiteds, &commitItem{
				Tag: tag,
				Key: key,
			})
		case tagFlushDB:
			commiteds = append(commiteds, &commitItem{
				Tag: tag,
			})
		case tagTxEnd:
			for _, v := range commiteds {
				switch v.Tag {
				case tagItemWriteSet:
					db.insertIntoDatabase(&dbItem{
						key:   v.Key,
						value: v.Value,
						data:  v.Data,
					})
				case tagItemWriteDel:
					db.deleteFromDatabase(&dbItem{
						key: v.Key,
					})
				case tagFlushDB:
					db.keys = btree.New(btreeDegrees, nil)
				default:
					return ErrInvalidDatabase
				}
			}
			commiteds = commiteds[:0]
			lastTxPos = fileOffset
		default:
			return ErrInvalidDatabase
		}
	}
	return nil
}

// load reads entries from the append only database file and fills the database.
// The file format uses the Redis append only file format, which is and a series
// of RESP commands. For more information on RESP please read
// http://redis.io/topics/protocol. The only supported RESP commands are DEL and
// SET.
func (db *DB) load() error {
	fi, err := db.file.Stat()
	if err != nil {
		return err
	}
	if err := db.readLoad(db.file, fi.ModTime()); err != nil {
		return err
	}
	pos, err := db.file.Seek(0, 2)
	if err != nil {
		return err
	}
	db.lastaofsz = int(pos)
	return nil
}

// managed calls a block of code that is fully contained in a transaction.
// This method is intended to be wrapped by Update and View
func (db *DB) managed(writable bool, fn func(tx *Tx) error) (err error) {
	var tx *Tx
	tx, err = db.Begin(writable)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			// The caller returned an error. We must rollback.
			_ = tx.Rollback()
			return
		}
		if writable {
			// Everything went well. Lets Commit()
			err = tx.Commit()
		} else {
			// read-only transaction can only roll back.
			err = tx.Rollback()
		}
	}()
	tx.funcd = true
	defer func() {
		tx.funcd = false
	}()
	err = fn(tx)
	return
}

// View executes a function within a managed read-only transaction.
// When a non-nil error is returned from the function that error will be return
// to the caller of View().
//
// Executing a manual commit or rollback from inside the function will result
// in a panic.
func (db *DB) View(fn func(tx *Tx) error) error {
	return db.managed(false, fn)
}

// Update executes a function within a managed read/write transaction.
// The transaction has been committed when no error is returned.
// In the event that an error is returned, the transaction will be rolled back.
// When a non-nil error is returned from the function, the transaction will be
// rolled back and the that error will be return to the caller of Update().
//
// Executing a manual commit or rollback from inside the function will result
// in a panic.
func (db *DB) Update(fn func(tx *Tx) error) error {
	return db.managed(true, fn)
}

// get return an item or nil if not found.
func (db *DB) get(key []byte) *dbItem {
	item := db.keys.Get(&dbItem{key: key})
	if item != nil {
		return item.(*dbItem)
	}
	return nil
}
