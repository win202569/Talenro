package nodeissuer

import "errors"

var (
	ErrInvalidInput                = errors.New("node issuer: invalid input")
	ErrLimitExceeded               = errors.New("node issuer: limit exceeded")
	ErrInvalidMaterial             = errors.New("node issuer: invalid material")
	ErrUnsupportedSignatureProfile = errors.New("node issuer: unsupported leaf signature profile")
	ErrIssuerRelationMismatch      = errors.New("node issuer: issuer relation mismatch")
	ErrInvalidSignature            = errors.New("node issuer: invalid leaf signature")
)
