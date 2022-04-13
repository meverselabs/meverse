package txpool

import "errors"

// TransactionPool errors
var (
	ErrEmptyQueue                = errors.New("empty queue")
	ErrNotAccountTransaction     = errors.New("not account transaction")
	ErrExistTransaction          = errors.New("exist transaction")
	ErrTransactionPoolOverflowed = errors.New("transaction pool overflowed")
	ErrPastSeq                   = errors.New("past seq")
	ErrTooFarSeq                 = errors.New("too far seq")
)
