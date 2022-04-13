package keydb

import "errors"

// errors
var (
	// ErrTxNotWritable is returned when performing a write operation on a
	// read-only transaction.
	ErrTxNotWritable = errors.New("tx not writable")

	// ErrTxClosed is returned when committing or rolling back a transaction
	// that has already been committed or rolled back.
	ErrTxClosed = errors.New("tx closed")

	// ErrNotFound is returned when an item or index is not in the database.
	ErrNotFound = errors.New("not found")

	// ErrInvalidDatabase is returned when the database file is an invalid format.
	ErrInvalidDatabase = errors.New("invalid database")

	// ErrDatabaseClosed is returned when the database is closed.
	ErrDatabaseClosed = errors.New("database closed")

	// ErrInvalidOperation is returned when an operation cannot be completed.
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrInvalidSyncPolicy is returned for an invalid SyncPolicy value.
	ErrInvalidSyncPolicy = errors.New("invalid sync policy")

	// ErrShrinkInProcess is returned when a shrink operation is in-process.
	ErrShrinkInProcess = errors.New("shrink is in-process")

	// ErrTxIterating is returned when Set or Delete are called while iterating.
	ErrTxIterating = errors.New("tx is iterating")

	ErrNotExistUnmarshaler = errors.New("not exist unmarshaler")

	errValidEOF = errors.New("valid eof")
)
