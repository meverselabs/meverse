package formulator

import "errors"

var (
	ErrNotFormulatorOwner               = errors.New("not formulator owner")
	ErrInvalidSigmaCreationCount        = errors.New("invalid sigma creation count")
	ErrInvalidSigmaCreationBlocks       = errors.New("invalid sigma creation blocks")
	ErrInvalidSigmaFormulatorType       = errors.New("invalid sigma formulator type")
	ErrInvalidOmegaCreationCount        = errors.New("invalid omega creation count")
	ErrInvalidOmegaCreationBlocks       = errors.New("invalid omega creation blocks")
	ErrInvalidOmegaFormulatorType       = errors.New("invalid omega formulator type")
	ErrInvalidStakeAmount               = errors.New("invalid stake amount")
	ErrInvalidStakeGenerator            = errors.New("invalid stake generator")
	ErrUnknownFormulatorType            = errors.New("unknown formulator type")
	ErrNotExistFormulator               = errors.New("not exist formulator")
	ErrNotGenesis                       = errors.New("not genesis")
	ErrApprovalToCurrentOwner           = errors.New("approval to current owner")
	ErrNotFormulatorApproved            = errors.New("formulator not approved")
	ErrTransferToZeroAddress            = errors.New("transfer to the zero address")
	ErrAlreadyRegisteredSalesFormulator = errors.New("already registered sales")
	ErrNotRegisteredSalesFormulator     = errors.New("not registerd sales")
)
