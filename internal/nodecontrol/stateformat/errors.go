package stateformat

import "errors"

// ErrInvalidFormat is deliberately independent of untrusted input values.
var ErrInvalidFormat = errors.New("invalid node state format")
