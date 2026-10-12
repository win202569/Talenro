package statepolicy

import "context"

type Code uint8

const (
	InvalidInput Code = iota + 1
	Rejected
	AssuranceMissing
	SecurityFault
	Canceled
)

type Reason uint8

const (
	Malformed Reason = iota + 1
	SignatureOrBinding
	UnsupportedRange
	MissingObservation
	IdentityMismatch
	StateRollback
	ContentConflict
	InvalidTime
	Expired
	MissingHistory
	HistoryConflict
	EmergencyRevocation
	StaleCandidate
	Cancellation
)

type PolicyError struct {
	Code   Code
	Reason Reason
}

func (e *PolicyError) Error() string { return "statepolicy: conditional model rejected" }
func fail(c Code, r Reason) error    { return &PolicyError{c, r} }
func contextError(ctx context.Context) error {
	if ctx == nil {
		return fail(InvalidInput, Malformed)
	}
	if ctx.Err() != nil {
		return fail(Canceled, Cancellation)
	}
	return nil
}
