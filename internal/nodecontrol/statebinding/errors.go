package statebinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

var (
	ErrInvalidInput       = errors.New("statebinding: invalid input")
	ErrLimitExceeded      = errors.New("statebinding: limit exceeded")
	ErrIncompleteCoverage = errors.New("statebinding: incomplete coverage")
	ErrInvalidMaterial    = errors.New("statebinding: invalid material")
	ErrMissingMaterial    = errors.New("statebinding: missing material")
	ErrDuplicateMaterial  = errors.New("statebinding: duplicate material")
	ErrBindingMismatch    = errors.New("statebinding: binding mismatch")
	ErrSignatureMismatch  = errors.New("statebinding: signature mismatch")
	ErrRollback           = errors.New("statebinding: rollback")
	ErrConflict           = errors.New("statebinding: conflict")
	ErrClearedModel       = errors.New("statebinding: cleared model")
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
	case errors.Is(err, ErrInvalidInput), errors.Is(err, statecontract.ErrInvalidInput), errors.Is(err, statematerials.ErrInvalidInput), errors.Is(err, statecomposition.ErrInvalidInput):
		return ErrInvalidInput
	case errors.Is(err, ErrLimitExceeded), errors.Is(err, statematerials.ErrLimitExceeded), errors.Is(err, statecomposition.ErrLimitExceeded):
		return ErrLimitExceeded
	case errors.Is(err, ErrIncompleteCoverage), errors.Is(err, statecontract.ErrIncompleteCoverage):
		return ErrIncompleteCoverage
	case errors.Is(err, ErrInvalidMaterial), errors.Is(err, statematerials.ErrInvalidMaterial), errors.Is(err, statecomposition.ErrInvalidMaterial):
		return ErrInvalidMaterial
	case errors.Is(err, ErrMissingMaterial), errors.Is(err, statematerials.ErrMissingMaterial), errors.Is(err, statecomposition.ErrMissingMaterial):
		return ErrMissingMaterial
	case errors.Is(err, ErrDuplicateMaterial), errors.Is(err, statematerials.ErrDuplicateMaterial), errors.Is(err, statecomposition.ErrDuplicateMaterial):
		return ErrDuplicateMaterial
	case errors.Is(err, ErrBindingMismatch), errors.Is(err, statecontract.ErrBindingMismatch), errors.Is(err, statematerials.ErrReferenceMismatch), errors.Is(err, statecomposition.ErrBindingMismatch):
		return ErrBindingMismatch
	case errors.Is(err, ErrSignatureMismatch), errors.Is(err, statecomposition.ErrSignatureMismatch):
		return ErrSignatureMismatch
	case errors.Is(err, ErrRollback), errors.Is(err, statecontract.ErrRollback):
		return ErrRollback
	case errors.Is(err, ErrConflict), errors.Is(err, statecontract.ErrConflict):
		return ErrConflict
	case errors.Is(err, ErrClearedModel), errors.Is(err, statecontract.ErrStaleModel), errors.Is(err, statematerials.ErrClearedModel), errors.Is(err, statecomposition.ErrClearedModel):
		return ErrClearedModel
	default:
		return ErrInvalidInput
	}
}
