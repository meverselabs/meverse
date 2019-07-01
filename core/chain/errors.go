package chain

import "errors"

// errors
var (
	ErrExistProcessName             = errors.New("exist process name")
	ErrExistProcessID               = errors.New("exist process id")
	ErrNotExistProcess              = errors.New("not exist process")
	ErrInvalidVersion               = errors.New("invalid version")
	ErrInvalidHeight                = errors.New("invalid height")
	ErrInvalidPrevHash              = errors.New("invalid prev hash")
	ErrInvalidContextHash           = errors.New("invalid context hash")
	ErrInvalidLevelRootHash         = errors.New("invalid level root hash")
	ErrInvalidTimestamp             = errors.New("invalid timestamp")
	ErrExceedHashCount              = errors.New("exceed hash count")
	ErrInvalidHashCount             = errors.New("invalid hash count")
	ErrInvalidGenesisHash           = errors.New("invalid genesis hash")
	ErrInvalidTxInKey               = errors.New("invalid txin key")
	ErrChainClosed                  = errors.New("chain closed")
	ErrStoreClosed                  = errors.New("store closed")
	ErrNotExistKey                  = errors.New("not exist key")
	ErrAlreadyGenesised             = errors.New("already genesised")
	ErrDirtyContext                 = errors.New("dirty context")
	ErrZeroIDIsReservedForConsensus = errors.New("zero id is reserved for consensus")
)
