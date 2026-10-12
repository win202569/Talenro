package statematerials

import (
	"context"
	"reflect"
)

func Evaluate(ctx context.Context, in Input) (result *MaterialSubsetModel, err error) {
	if ctx == nil {
		return nil, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if e := validateInput(in); e != nil {
		return nil, e
	}
	owned := cloneInput(in)
	defer func() {
		if result == nil {
			clearInput(&owned)
		}
	}()
	parsed, e := parseBatch(ctx, owned)
	if e != nil {
		return nil, e
	}
	defer parsed.Clear()
	if e := validateRelations(ctx, owned, parsed); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return &MaterialSubsetModel{valid: true, input: owned}, nil
}
func Recheck(ctx context.Context, m *MaterialSubsetModel, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if m == nil || !m.valid {
		return ErrClearedModel
	}
	fresh, e := Evaluate(ctx, in)
	if e != nil {
		return e
	}
	defer fresh.Clear()
	same := reflect.DeepEqual(m.input, fresh.input)
	if e := ctx.Err(); e != nil {
		return e
	}
	if !same {
		return ErrReferenceMismatch
	}
	return nil
}
