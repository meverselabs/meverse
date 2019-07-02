package chain

import "errors"

// errors
var (
	ErrExistProcessName     = errors.New("exist process name")
	ErrExistProcessID       = errors.New("exist process id")
	ErrNotExistProcess      = errors.New("not exist process")
	ErrExistServiceName     = errors.New("exist service name")
	ErrExistServiceID       = errors.New("exist service id")
	ErrNotExistService      = errors.New("not exist service")
	ErrInvalidVersion       = errors.New("invalid version")
	ErrInvalidHeight        = errors.New("invalid height")
	ErrInvalidPrevHash      = errors.New("invalid prev hash")
	ErrInvalidContextHash   = errors.New("invalid context hash")
	ErrInvalidLevelRootHash = errors.New("invalid level root hash")
	ErrInvalidTimestamp     = errors.New("invalid timestamp")
	ErrExceedHashCount      = errors.New("exceed hash count")
	ErrInvalidHashCount     = errors.New("invalid hash count")
	ErrInvalidGenesisHash   = errors.New("invalid genesis hash")
	ErrInvalidTxInKey       = errors.New("invalid txin key")
	ErrChainClosed          = errors.New("chain closed")
	ErrStoreClosed          = errors.New("store closed")
	ErrNotExistKey          = errors.New("not exist key")
	ErrAlreadyGenesised     = errors.New("already genesised")
	ErrDirtyContext         = errors.New("dirty context")
	ErrReservedID           = errors.New("reserved id")
	ErrAddBeforeChainInit   = errors.New("add before chain init")
)
