package staterecovery

import "errors"

var (
	ErrInvalidInput      = errors.New("recovery model: invalid input")
	ErrAssuranceMissing  = errors.New("recovery model: assurance missing")
	ErrSecurityFault     = errors.New("recovery model: security fault")
	ErrCounterExhausted  = errors.New("recovery model: counter exhausted")
	ErrStaleModel        = errors.New("recovery model: stale model")
	ErrInvalidTransition = errors.New("recovery model: invalid transition")
)

type SecurityReason uint8

const (
	ReasonUnknown SecurityReason = iota
	CorruptSlot
	UnsupportedVersion
	ScopeMismatch
	DuplicateEpoch
	Fork
	CounterMismatch
	UnexpectedFuture
)

type SecurityError struct{ Reason SecurityReason }

func (e *SecurityError) Error() string {
	if e == nil {
		return "recovery model: security fault"
	}
	switch e.Reason {
	case CorruptSlot:
		return "recovery model: corrupt slot"
	case UnsupportedVersion:
		return "recovery model: unsupported version"
	case ScopeMismatch:
		return "recovery model: scope mismatch"
	case DuplicateEpoch:
		return "recovery model: duplicate epoch"
	case Fork:
		return "recovery model: fork"
	case CounterMismatch:
		return "recovery model: counter mismatch"
	case UnexpectedFuture:
		return "recovery model: unexpected future"
	default:
		return "recovery model: security fault"
	}
}
func (e *SecurityError) Unwrap() error     { return ErrSecurityFault }
func security(reason SecurityReason) error { return &SecurityError{Reason: reason} }
