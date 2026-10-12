package nodecerttime

import "errors"

var (
	ErrInvalidInput    = errors.New("nodecerttime: invalid input")
	ErrLimitExceeded   = errors.New("nodecerttime: limit exceeded")
	ErrInvalidEvidence = errors.New("nodecerttime: invalid evidence")
	ErrBindingMismatch = errors.New("nodecerttime: binding mismatch")
	ErrNotYetValid     = errors.New("nodecerttime: certificate not yet valid")
	ErrExpired         = errors.New("nodecerttime: certificate expired")
)
