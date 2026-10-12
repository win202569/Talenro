package statecontract

import (
	"context"
	"reflect"
)

func EvaluateModel(ctx context.Context, in ModelInput) (result *ContractModel, err error) {
	if ctx == nil {
		return nil, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if len(in.RootTransition) > 32 {
		return nil, ErrInvalidInput
	}
	if e := validateShape(in.Previous); e != nil {
		return nil, e
	}
	if e := validateShape(in.Proposed); e != nil {
		return nil, e
	}
	if e := visibleIdentities(in); e != nil {
		return nil, e
	}
	if e := validateBindings(in.Previous); e != nil {
		return nil, e
	}
	if e := validateBindings(in.Proposed); e != nil {
		return nil, e
	}
	if e := validateTransition(in); e != nil {
		return nil, e
	}
	m := &ContractModel{valid: true, input: cloneInput(in)}
	defer func() {
		if result == nil {
			m.Clear()
		}
	}()
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return m, nil
}
func RecheckModel(ctx context.Context, m *ContractModel, in ModelInput) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if m == nil || !m.valid {
		return ErrStaleModel
	}
	fresh, e := EvaluateModel(ctx, in)
	if e != nil {
		return e
	}
	defer fresh.Clear()
	same := reflect.DeepEqual(m.input, fresh.input)
	if e := ctx.Err(); e != nil {
		return e
	}
	if !same {
		return ErrStaleModel
	}
	return nil
}
