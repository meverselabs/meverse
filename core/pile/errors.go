package pile

import "errors"

// errors
var (
	ErrInvalidChunkBeginHeight     = errors.New("invalid chunk begin height")
	ErrInvalidChunkEndHeight       = errors.New("invalid chunk end height")
	ErrInvalidAppendHeight         = errors.New("invalid append height")
	ErrInvalidHeight               = errors.New("invalid height")
	ErrInvalidFileSize             = errors.New("invalid file size")
	ErrInvalidGenesisHash          = errors.New("invalid genesis hash")
	ErrInvalidDataIndex            = errors.New("invalid data index")
	ErrMissingPile                 = errors.New("invalid missing pile")
	ErrAlreadyInitialized          = errors.New("already initialized")
	ErrExeedMaximumDataArrayLength = errors.New("exceed maximum data array length")
	ErrHeightCrashed               = errors.New("height crashed")
)
