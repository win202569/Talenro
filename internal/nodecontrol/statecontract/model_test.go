package statecontract

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func TestContractNoAuthority(t *testing.T) {
	m, e := EvaluateModel(context.Background(), pair())
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	s := m.Summary()
	if s.Status != ContractShapeValid || s.Outstanding != ([5]Obligation{MaterialClosure, DomainSemantics, CompositionRecheck, PhysicalCommit, RuntimeAuthorization}) {
		t.Fatal("authority strength")
	}
	s.Outstanding[0] = ObligationUnknown
	s.Coverage[0] = Unknown
	if m.Summary().Outstanding[0] != MaterialClosure || m.Summary().Coverage[0] != Present {
		t.Fatal("summary alias")
	}
	var nilModel *ContractModel
	if nilModel.Summary() != (ContractSummary{}) {
		t.Fatal("nil summary")
	}
	nilModel.Clear()
	m.Clear()
	m.Clear()
	if m.Summary() != (ContractSummary{}) {
		t.Fatal("clear")
	}
	m, e = EvaluateModel(context.Background(), ModelInput{})
	if e == nil || m != nil {
		t.Fatal("invalid succeeded")
	}
}
func mutateSlices(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Struct:
			mutateSlices(f)
		case reflect.Slice:
			if f.Len() > 0 {
				f.Index(0).Set(reflect.Zero(f.Type().Elem()))
			}
		}
	}
}
func TestContractOwnership(t *testing.T) {
	in := pair()
	in.Previous.Recovery = recoveryFixture()
	in.Proposed.Recovery = recoveryFixture()
	in.Previous.Recovery.Local = []LocalFaultBinding{{faultID(1), faultID(2)}}
	in.Proposed.Recovery.Local = []LocalFaultBinding{{faultID(1), faultID(2)}}
	in.Previous.Recovery.Supervisor = []SupervisorFaultBinding{{faultID(1), faultID(2), d(1)}}
	in.Proposed.Recovery.Supervisor = []SupervisorFaultBinding{{faultID(1), faultID(2), d(1)}}
	m, e := EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	before := m.Summary()
	mutateSlices(reflect.ValueOf(&in).Elem())
	if m.Summary() != before {
		t.Fatal("borrowed summary")
	}
	// This compares private owned values solely to detect accidental slice borrowing.
	if m.input.Previous.Root.KeyIDs[0] != d(1) || m.input.Proposed.Recovery.Local[0].FaultID != faultID(1) || m.input.Proposed.LKG.Entries[0].SlotID != "slot" {
		t.Fatal("borrowed slices")
	}
}
func TestContractRecheck(t *testing.T) {
	m, e := EvaluateModel(context.Background(), pair())
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	if e := RecheckModel(context.Background(), m, pair()); e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(*ModelInput){func(x *ModelInput) { x.Proposed.Metadata.Mark.Material.Envelope = d(99) }, func(x *ModelInput) { x.Proposed.Desired.Window.Until++ }, func(x *ModelInput) { x.Proposed.Releases.BuildID = d(99) }, func(x *ModelInput) { x.Proposed.Identity.History.Payload = d(99) }} {
		in := pair()
		edit(&in)
		if !errors.Is(RecheckModel(context.Background(), m, in), ErrStaleModel) {
			t.Fatal("dependency replacement accepted")
		}
	}
	in := pair()
	in.Proposed.Metadata.RevokedIDs = []contracts.Digest{}
	if e := RecheckModel(context.Background(), m, in); e != nil {
		t.Fatal("empty differs from nil", e)
	}
	m.Clear()
	if !errors.Is(RecheckModel(context.Background(), m, pair()), ErrStaleModel) {
		t.Fatal("cleared accepted")
	}
}

type lastCancel struct {
	context.Context
	calls int
}

func (c *lastCancel) Err() error {
	c.calls++
	if c.calls >= 2 {
		return context.Canceled
	}
	return nil
}
func TestContractCancellationAndErrors(t *testing.T) {
	if m, e := EvaluateModel(nil, pair()); m != nil || !errors.Is(e, ErrInvalidInput) {
		t.Fatal("nil context")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if m, e := EvaluateModel(ctx, ModelInput{}); m != nil || !errors.Is(e, context.Canceled) {
		t.Fatal("cancellation precedence")
	}
	if m, e := EvaluateModel(&lastCancel{Context: context.Background()}, pair()); m != nil || !errors.Is(e, context.Canceled) {
		t.Fatal("late cancellation")
	}
	in := pair()
	in.Proposed.Scope.NodeID = "secret"
	m, e := EvaluateModel(context.Background(), in)
	if m != nil || !errors.Is(e, ErrInvalidInput) || strings.Contains(e.Error(), "secret") {
		t.Fatal("leaking input or wrong error")
	}
	wrapped := &ContractError{Domain: DomainRoot, Cause: ErrConflict}
	if !errors.Is(wrapped, ErrConflict) {
		t.Fatal("unwrap")
	}
}
