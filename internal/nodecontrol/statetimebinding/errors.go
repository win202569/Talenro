package statetimebinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
)

var (
	ErrInvalidInput            = errors.New("statetimebinding: invalid input")
	ErrLimitExceeded           = errors.New("statetimebinding: limit exceeded")
	ErrIncompleteCoverage      = errors.New("statetimebinding: incomplete coverage")
	ErrInvalidMaterial         = errors.New("statetimebinding: invalid material")
	ErrMissingMaterial         = errors.New("statetimebinding: missing material")
	ErrDuplicateMaterial       = errors.New("statetimebinding: duplicate material")
	ErrBindingMismatch         = errors.New("statetimebinding: binding mismatch")
	ErrSignatureMismatch       = errors.New("statetimebinding: signature mismatch")
	ErrRollback                = errors.New("statetimebinding: rollback")
	ErrConflict                = errors.New("statetimebinding: conflict")
	ErrClearedModel            = errors.New("statetimebinding: cleared model")
	ErrUnsupportedPolicyShape  = errors.New("statetimebinding: unsupported policy shape")
	ErrPolicyRejected          = errors.New("statetimebinding: policy rejected")
	ErrUnsupportedBindingShape = errors.New("statetimebinding: unsupported binding shape")
	ErrUnauthorizedKey         = errors.New("statetimebinding: unauthorized key")
	ErrAssuranceMissing        = errors.New("statetimebinding: assurance missing")
	ErrInvalidTime             = errors.New("statetimebinding: invalid time")
	ErrExpired                 = errors.New("statetimebinding: expired")
	ErrCheckpointRollback      = errors.New("statetimebinding: checkpoint rollback")
	ErrStaleEvidence           = errors.New("statetimebinding: stale evidence")
)

type PolicyFailure struct {
	Code   statepolicy.Code
	Reason statepolicy.Reason
}

func (e *PolicyFailure) Error() string { return ErrPolicyRejected.Error() }
func (e *PolicyFailure) Unwrap() error { return ErrPolicyRejected }
func mapError(ctx context.Context, err error) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if err == nil {
		return nil
	}
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, e) {
			return e
		}
	}
	var failure *statepolicybinding.PolicyFailure
	if errors.As(err, &failure) {
		if failure == nil || failure.Code < statepolicy.Code(1) || failure.Code > statepolicy.Code(5) || failure.Reason < statepolicy.Reason(1) || failure.Reason > statepolicy.Reason(14) {
			return ErrInvalidInput
		}
		return &PolicyFailure{Code: failure.Code, Reason: failure.Reason}
	}
	var own *PolicyFailure
	if errors.As(err, &own) {
		if own == nil || own.Code < statepolicy.Code(1) || own.Code > statepolicy.Code(5) || own.Reason < statepolicy.Reason(1) || own.Reason > statepolicy.Reason(14) {
			return ErrInvalidInput
		}
		return &PolicyFailure{Code: own.Code, Reason: own.Reason}
	}
	for _, e := range []error{ErrInvalidInput, ErrLimitExceeded, ErrIncompleteCoverage, ErrInvalidMaterial, ErrMissingMaterial, ErrDuplicateMaterial, ErrBindingMismatch, ErrSignatureMismatch, ErrRollback, ErrConflict, ErrClearedModel, ErrUnsupportedPolicyShape, ErrPolicyRejected, ErrUnsupportedBindingShape, ErrUnauthorizedKey, ErrAssuranceMissing, ErrInvalidTime, ErrExpired, ErrCheckpointRollback, ErrStaleEvidence} {
		if errors.Is(err, e) {
			return e
		}
	}
	for _, pair := range [][2]error{
		{statepolicybinding.ErrInvalidInput, ErrInvalidInput},
		{statepolicybinding.ErrLimitExceeded, ErrLimitExceeded},
		{statepolicybinding.ErrIncompleteCoverage, ErrIncompleteCoverage},
		{statepolicybinding.ErrInvalidMaterial, ErrInvalidMaterial},
		{statepolicybinding.ErrMissingMaterial, ErrMissingMaterial},
		{statepolicybinding.ErrDuplicateMaterial, ErrDuplicateMaterial},
		{statepolicybinding.ErrBindingMismatch, ErrBindingMismatch},
		{statepolicybinding.ErrSignatureMismatch, ErrSignatureMismatch},
		{statepolicybinding.ErrRollback, ErrRollback},
		{statepolicybinding.ErrConflict, ErrConflict},
		{statepolicybinding.ErrClearedModel, ErrClearedModel},
		{statepolicybinding.ErrUnsupportedPolicyShape, ErrUnsupportedPolicyShape},
		{statepolicybinding.ErrPolicyRejected, ErrPolicyRejected},
		{statetime.ErrInvalidInput, ErrInvalidInput},
		{statetime.ErrBindingMismatch, ErrBindingMismatch},
		{statetime.ErrSignatureMismatch, ErrSignatureMismatch},
		{statetime.ErrUnauthorizedKey, ErrUnauthorizedKey},
		{statetime.ErrAssuranceMissing, ErrAssuranceMissing},
		{statetime.ErrInvalidTime, ErrInvalidTime},
		{statetime.ErrExpired, ErrExpired},
		{statetime.ErrCheckpointRollback, ErrCheckpointRollback},
		{statetime.ErrStaleEvidence, ErrStaleEvidence},
	} {
		if errors.Is(err, pair[0]) {
			return pair[1]
		}
	}
	return ErrInvalidInput
}
