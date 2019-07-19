package admin

import "errors"

// errors
var (
	ErrInvalidAdminAddress                    = errors.New("invalid admin address")
	ErrUnauthorizedTransaction                = errors.New("unauthorized transaction")
	ErrAdminAddressShouldBeSetupInApplication = errors.New("admin address should be setup in application")
)
