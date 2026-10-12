package stateverify

import "errors"

var (
	ErrInvalidInput      = errors.New("stateverify: invalid input")
	ErrBindingMismatch   = errors.New("stateverify: binding mismatch")
	ErrSignatureMismatch = errors.New("stateverify: signature mismatch")
	ErrCanceled          = errors.New("stateverify: canceled")
)
