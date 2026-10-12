package statematerials

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid material subset input")
	ErrLimitExceeded     = errors.New("material subset limit exceeded")
	ErrInvalidMaterial   = errors.New("invalid raw material")
	ErrMissingMaterial   = errors.New("missing material or expectation")
	ErrDuplicateMaterial = errors.New("duplicate material or expectation")
	ErrReferenceMismatch = errors.New("material reference mismatch")
	ErrClearedModel      = errors.New("missing or cleared material subset model")
)
