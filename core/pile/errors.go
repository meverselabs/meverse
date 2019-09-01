package pile

import "errors"

// errors
var (
	ErrInvalidChunkBeginHeight = errors.New("invalid chunk begin height")
	ErrInvalidChunkEndHeight   = errors.New("invalid chunk end height")
	ErrInvalidAppendHeight     = errors.New("invalid append height")
	ErrInvalidHeight           = errors.New("invalid height")
	ErrMissingPile             = errors.New("invalid missing pile")
	ErrInvalidFileSize         = errors.New("invalid file size")
	ErrInvalidGenesisHash      = errors.New("invalid genesis hash")
	ErrAlreadyInitialized      = errors.New("already initialized")
)
