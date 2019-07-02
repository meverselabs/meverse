package pof

import "errors"

// consensus errors
var (
	ErrInsufficientCandidateCount    = errors.New("insufficient candidate count")
	ErrExceedCandidateCount          = errors.New("exceed candidate count")
	ErrInvalidMaxBlocksPerFormulator = errors.New("invalid max blocks per formulator")
	ErrInvalidObserverKey            = errors.New("invalid observer key")
	ErrInvalidTopAddress             = errors.New("invalid top address")
	ErrInvalidTopSignature           = errors.New("invalid top signature")
	ErrInvalidSignatureCount         = errors.New("invalid signature count")
	ErrInvalidPhase                  = errors.New("invalid phase")
	ErrExistAddress                  = errors.New("exist address")
)
