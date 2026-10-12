package nodeca

import "errors"

var (
	ErrInvalidInput         = errors.New("nodeca: invalid input")
	ErrLimitExceeded        = errors.New("nodeca: limit exceeded")
	ErrInvalidMaterial      = errors.New("nodeca: invalid material")
	ErrUnsupportedCAProfile = errors.New("nodeca: unsupported CA profile")
	ErrBindingMismatch      = errors.New("nodeca: binding mismatch")
)
