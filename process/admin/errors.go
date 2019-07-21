package admin

import "errors"

// errors
var (
	ErrInvalidAdminAddress     = errors.New("invalid admin address")
	ErrUnauthorizedTransaction = errors.New("unauthorized transaction")
	ErrNotExistAdminAddress    = errors.New("not exist admin address")
)
