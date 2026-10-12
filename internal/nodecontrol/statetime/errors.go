package statetime

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid time evidence input")
	ErrBindingMismatch    = errors.New("time evidence binding mismatch")
	ErrSignatureMismatch  = errors.New("time evidence signature mismatch")
	ErrUnauthorizedKey    = errors.New("unauthorized time evidence key")
	ErrAssuranceMissing   = errors.New("time evidence assurance missing")
	ErrInvalidTime        = errors.New("invalid time observation")
	ErrExpired            = errors.New("time evidence outside validity window")
	ErrCheckpointRollback = errors.New("time evidence checkpoint rollback")
	ErrStaleEvidence      = errors.New("stale time evidence model")
)
