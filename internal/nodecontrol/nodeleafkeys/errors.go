package nodeleafkeys

import (
	"errors"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
)

var (
	ErrInvalidInput                    = nodeleaf.ErrInvalidInput
	ErrLimitExceeded                   = nodeleaf.ErrLimitExceeded
	ErrInvalidMaterial                 = nodeleaf.ErrInvalidMaterial
	ErrUnsupportedLeafProfile          = nodeleaf.ErrUnsupportedLeafProfile
	ErrBindingMismatch                 = nodeleaf.ErrBindingMismatch
	ErrUnsupportedKeyIdentifierProfile = errors.New("node leaf keys: unsupported key identifier profile")
)
