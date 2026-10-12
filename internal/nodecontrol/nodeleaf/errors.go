package nodeleaf

import "errors"

var (
	ErrInvalidInput           = errors.New("node leaf: invalid input")
	ErrLimitExceeded          = errors.New("node leaf: limit exceeded")
	ErrInvalidMaterial        = errors.New("node leaf: invalid material")
	ErrUnsupportedLeafProfile = errors.New("node leaf: unsupported leaf profile")
	ErrBindingMismatch        = errors.New("node leaf: binding mismatch")
)
