package nodeissuerbinding

import "errors"

var (
	ErrInvalidInput    = errors.New("nodeissuerbinding: invalid input")
	ErrLimitExceeded   = errors.New("nodeissuerbinding: limit exceeded")
	ErrInvalidEvidence = errors.New("nodeissuerbinding: invalid evidence")
	ErrBindingMismatch = errors.New("nodeissuerbinding: binding mismatch")
)
