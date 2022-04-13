// Package keydb came from buntdb that implements a low-level in-memory key/value store in pure Go.
// It persists to disk, is ACID compliant, and uses locking for multiple
// readers and a single writer. Bunt is ideal for projects that need
// a dependable database, and favor speed over data size.
package keydb

// SyncPolicy represents how often data is synced to disk.
type SyncPolicy int

// Config represents database configuration options. These
// options are used to change various behaviors of the database.
type Config struct {
	// SyncPolicy adjusts how often the data is synced to disk.
	// This value can be Never, EverySecond, or Always.
	// The default is EverySecond.
	SyncPolicy SyncPolicy

	// AutoShrinkPercentage is used by the background process to trigger
	// a shrink of the aof file when the size of the file is larger than the
	// percentage of the result of the previous shrunk file.
	// For example, if this value is 100, and the last shrink process
	// resulted in a 100mb file, then the new aof file must be 200mb before
	// a shrink is triggered.
	AutoShrinkPercentage int

	// AutoShrinkMinSize defines the minimum size of the aof file before
	// an automatic shrink can occur.
	AutoShrinkMinSize int

	// AutoShrinkDisabled turns off automatic background shrinking
	AutoShrinkDisabled bool
}
