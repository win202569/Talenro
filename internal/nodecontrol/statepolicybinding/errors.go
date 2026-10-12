package statepolicybinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
)

var (
	ErrInvalidInput           = errors.New("statepolicybinding: invalid input")
	ErrLimitExceeded          = errors.New("statepolicybinding: limit exceeded")
	ErrIncompleteCoverage     = errors.New("statepolicybinding: incomplete coverage")
	ErrInvalidMaterial        = errors.New("statepolicybinding: invalid material")
	ErrMissingMaterial        = errors.New("statepolicybinding: missing material")
	ErrDuplicateMaterial      = errors.New("statepolicybinding: duplicate material")
	ErrBindingMismatch        = errors.New("statepolicybinding: binding mismatch")
	ErrSignatureMismatch      = errors.New("statepolicybinding: signature mismatch")
	ErrRollback               = errors.New("statepolicybinding: rollback")
	ErrConflict               = errors.New("statepolicybinding: conflict")
	ErrClearedModel           = errors.New("statepolicybinding: cleared model")
	ErrUnsupportedPolicyShape = errors.New("statepolicybinding: unsupported policy shape")
	ErrPolicyRejected         = errors.New("statepolicybinding: policy rejected")
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
	for _, sentinel := range []error{context.Canceled, context.DeadlineExceeded, ErrInvalidInput, ErrLimitExceeded, ErrIncompleteCoverage, ErrInvalidMaterial, ErrMissingMaterial, ErrDuplicateMaterial, ErrBindingMismatch, ErrSignatureMismatch, ErrRollback, ErrConflict, ErrClearedModel, ErrUnsupportedPolicyShape, ErrPolicyRejected} {
		if errors.Is(err, sentinel) {
			return sentinel
		}
	}
	for _, pair := range [][2]error{{statebinding.ErrInvalidInput, ErrInvalidInput}, {statebinding.ErrLimitExceeded, ErrLimitExceeded}, {statebinding.ErrIncompleteCoverage, ErrIncompleteCoverage}, {statebinding.ErrInvalidMaterial, ErrInvalidMaterial}, {statebinding.ErrMissingMaterial, ErrMissingMaterial}, {statebinding.ErrDuplicateMaterial, ErrDuplicateMaterial}, {statebinding.ErrBindingMismatch, ErrBindingMismatch}, {statebinding.ErrSignatureMismatch, ErrSignatureMismatch}, {statebinding.ErrRollback, ErrRollback}, {statebinding.ErrConflict, ErrConflict}, {statebinding.ErrClearedModel, ErrClearedModel}} {
		if errors.Is(err, pair[0]) {
			return pair[1]
		}
	}
	var policy *statepolicy.PolicyError
	if errors.As(err, &policy) && policy != nil {
		if policy.Code < statepolicy.Code(1) || policy.Code > statepolicy.Code(5) || policy.Reason < statepolicy.Reason(1) || policy.Reason > statepolicy.Reason(14) {
			return ErrInvalidInput
		}
		if policy.Code == statepolicy.Canceled || policy.Reason == statepolicy.Cancellation {
			return context.Canceled
		}
		return &PolicyFailure{Code: policy.Code, Reason: policy.Reason}
	}
	return ErrInvalidInput
}
