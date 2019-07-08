package txpool

import "errors"

// TransactionPool errors
var (
	ErrEmptyQueue            = errors.New("empty queue")
	ErrNotAccountTransaction = errors.New("not account transaction")
	ErrExistTransaction      = errors.New("exist transaction")
)
