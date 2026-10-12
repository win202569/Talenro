package statecomposition

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

var (
	ErrInvalidInput      = errors.New("statecomposition: invalid input")
	ErrLimitExceeded     = errors.New("statecomposition: limit exceeded")
	ErrInvalidMaterial   = errors.New("statecomposition: invalid material")
	ErrMissingMaterial   = errors.New("statecomposition: missing material")
	ErrDuplicateMaterial = errors.New("statecomposition: duplicate material")
	ErrBindingMismatch   = errors.New("statecomposition: binding mismatch")
	ErrSignatureMismatch = errors.New("statecomposition: signature mismatch")
	ErrClearedModel      = errors.New("statecomposition: cleared model")
)

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
	switch {
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	case errors.Is(err, statematerials.ErrInvalidInput):
		return ErrInvalidInput
	case errors.Is(err, statematerials.ErrLimitExceeded):
		return ErrLimitExceeded
	case errors.Is(err, statematerials.ErrInvalidMaterial), errors.Is(err, stateformat.ErrInvalidFormat), errors.Is(err, stateverify.ErrInvalidInput):
		return ErrInvalidMaterial
	case errors.Is(err, statematerials.ErrMissingMaterial):
		return ErrMissingMaterial
	case errors.Is(err, statematerials.ErrDuplicateMaterial):
		return ErrDuplicateMaterial
	case errors.Is(err, statematerials.ErrClearedModel):
		return ErrClearedModel
	case errors.Is(err, statematerials.ErrReferenceMismatch), errors.Is(err, stateverify.ErrBindingMismatch):
		return ErrBindingMismatch
	case errors.Is(err, stateverify.ErrSignatureMismatch):
		return ErrSignatureMismatch
	case errors.Is(err, stateverify.ErrCanceled):
		return context.Canceled
	default:
		return ErrInvalidInput
	}
}
