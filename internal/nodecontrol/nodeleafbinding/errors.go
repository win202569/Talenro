package nodeleafbinding

import "errors"

var (
	ErrInvalidInput    = errors.New("nodeleafbinding: invalid input")
	ErrLimitExceeded   = errors.New("nodeleafbinding: limit exceeded")
	ErrInvalidEvidence = errors.New("nodeleafbinding: invalid evidence")
	ErrBindingMismatch = errors.New("nodeleafbinding: binding mismatch")
)
