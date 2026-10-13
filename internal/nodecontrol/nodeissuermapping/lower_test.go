package nodeissuermapping

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	"testing"
	"time"
)

func TestIssuerMappingLowerErrors(t *testing.T) {
	var packets map[string]Input
	issuerMappingDecode(t, issuerMappingFrozenInputs, &packets)
	negative := 0
	typed := 0
	for name, in := range packets {
		m, want := nodeissuerbinding.Evaluate(context.Background(), in.Material)
		if m != nil {
			m.Clear()
		}
		if want == nil {
			continue
		}
		negative++
		o := &owner{}
		got := evaluateLower(context.Background(), o, in.Material)
		o.clear()
		if !issuerMappingErrorPreserved(got, want) {
			t.Fatal(name, got, want)
		}
		if name == "policy-failure" {
			if !issuerMappingPolicyFailure(want) || !issuerMappingPolicyFailure(got) {
				t.Fatal("missing structured policy control", got, want)
			}
			typed++
		}
	}
	if negative == 0 || typed != 1 {
		t.Fatal("missing negative/typed controls")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if evaluateLower(ctx, &owner{}, nodeissuerbinding.Input{}) != context.Canceled || evaluateLower(nil, &owner{}, nodeissuerbinding.Input{}) != ErrInvalidInput || evaluateLower(context.Background(), nil, nodeissuerbinding.Input{}) != ErrInvalidInput {
		t.Fatal("context/owner")
	}
}

// Preserve concrete types/fields through every unwrap level. Terminal sentinels
// must retain identity; independently allocated structured wrappers may differ.
func issuerMappingErrorPreserved(got, want error) bool {
	for a, b := got, want; a != nil || b != nil; a, b = errors.Unwrap(a), errors.Unwrap(b) {
		if reflect.TypeOf(a) != reflect.TypeOf(b) || !reflect.DeepEqual(a, b) {
			return false
		}
		nextA, nextB := errors.Unwrap(a), errors.Unwrap(b)
		if nextA == nil || nextB == nil {
			return nextA == nil && nextB == nil && reflect.TypeOf(a).Comparable() && a == b
		}
	}
	return true
}
func TestIssuerMappingErrorPreservationControls(t *testing.T) {
	sentinel := errors.New("fixed lower error")
	clone := errors.New("fixed lower error")
	t.Run("direct", func(t *testing.T) {
		if issuerMappingErrorPreserved(clone, sentinel) {
			t.Fatal("different sentinel identity accepted")
		}
	})
	t.Run("wrapped", func(t *testing.T) {
		if issuerMappingErrorPreserved(fmt.Errorf("context: %w", clone), fmt.Errorf("context: %w", sentinel)) {
			t.Fatal("different wrapped sentinel identity accepted")
		}
	})
	if !issuerMappingErrorPreserved(sentinel, sentinel) || !issuerMappingErrorPreserved(fmt.Errorf("context: %w", sentinel), fmt.Errorf("context: %w", sentinel)) {
		t.Fatal("same terminal sentinel rejected")
	}
	if issuerMappingErrorPreserved(nil, sentinel) || issuerMappingErrorPreserved(sentinel, nil) || !issuerMappingErrorPreserved(nil, nil) {
		t.Fatal("nil mismatch")
	}
}
func issuerMappingPolicyFailure(err error) bool {
	if err == nil {
		return false
	}
	typ := reflect.TypeOf(err)
	value := reflect.ValueOf(err)
	if typ.Kind() != reflect.Pointer || value.IsNil() || typ.Elem().Kind() != reflect.Struct {
		return false
	}
	if typ.Elem().PkgPath() != "talenro.local/platform/internal/nodecontrol/statetimebinding" || typ.Elem().Name() != "PolicyFailure" {
		return false
	}
	return value.Elem().FieldByName("Code").IsValid() && value.Elem().FieldByName("Reason").IsValid() && errors.Unwrap(err) != nil
}
func TestIssuerMappingStructuredFailureControls(t *testing.T) {
	t.Run("ordinary pointer is not structured", func(t *testing.T) {
		if issuerMappingPolicyFailure(errors.New("statetimebinding: policy rejected")) {
			t.Fatal("ordinary sentinel counted as structured policy failure")
		}
	})
	in := issuerMappingFixture(t, "policy-failure")
	model, want := nodeissuerbinding.Evaluate(context.Background(), in.Material)
	if model != nil {
		model.Clear()
		t.Fatal("policy control returned model")
	}
	if !issuerMappingPolicyFailure(want) || errors.Unwrap(want) == nil {
		t.Fatal("missing real structured policy failure", want)
	}
	clone := reflect.New(reflect.TypeOf(want).Elem())
	clone.Elem().Set(reflect.ValueOf(want).Elem())
	if !issuerMappingErrorPreserved(clone.Interface().(error), want) {
		t.Fatal("structured clone with preserved fields and terminal identity")
	}
	for _, name := range []string{"Code", "Reason"} {
		changed := reflect.New(reflect.TypeOf(want).Elem())
		changed.Elem().Set(reflect.ValueOf(want).Elem())
		field := changed.Elem().FieldByName(name)
		if !field.IsValid() || field.IsZero() {
			t.Fatal("inactive structured field control", name)
		}
		field.SetZero()
		if issuerMappingErrorPreserved(changed.Interface().(error), want) {
			t.Fatal("changed structured field accepted", name)
		}
	}
}
func TestIssuerMappingLowerOwnership(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	before := issuerMappingJSONHash(t, in)
	want := assertIssuerMappingLower(t, in)
	o := &owner{}
	defer o.clear()
	if e := evaluateLower(context.Background(), o, in.Material); e != nil || !reflect.DeepEqual(o.summary.Evidence, want) || issuerMappingJSONHash(t, in) != before {
		t.Fatal("capture", e)
	}
	views := ownedTestViews(reflect.ValueOf(o.summary.Evidence))
	o.clear()
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(want, assertIssuerMappingLower(t, in)) || issuerMappingJSONHash(t, in) != before {
		t.Fatal("lower clear caller")
	}
}
func TestIssuerMappingLowerConsistency(t *testing.T) {
	good := assertIssuerMappingLower(t, issuerMappingFixture(t, "serial-one-high"))
	if e := validateEvidence(good); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*nodeissuerbinding.Summary){
		func(s *nodeissuerbinding.Summary) { s.Proof = 0 },
		func(s *nodeissuerbinding.Summary) { s.Profile = "wrong" },
		func(s *nodeissuerbinding.Summary) { s.ObservedAt = time.Time{} },
		func(s *nodeissuerbinding.Summary) { s.ValidUntil = s.ObservedAt },
		func(s *nodeissuerbinding.Summary) { s.ObservedAt = s.ObservedAt.In(time.FixedZone("offset", 3600)) },
		func(s *nodeissuerbinding.Summary) { s.ObservedAt = time.Now() },
		func(s *nodeissuerbinding.Summary) { s.ValidUntil = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *nodeissuerbinding.Summary) { s.Evidence.ObservedAt = s.Evidence.ObservedAt.Add(time.Nanosecond) },
		func(s *nodeissuerbinding.Summary) { s.Evidence.ValidUntil = s.Evidence.ValidUntil.Add(time.Nanosecond) },
		func(s *nodeissuerbinding.Summary) {
			s.Evidence.Evidence.ObservedAt = s.Evidence.Evidence.ObservedAt.Add(time.Nanosecond)
		},
		func(s *nodeissuerbinding.Summary) {
			s.Evidence.Evidence.ValidUntil = s.Evidence.Evidence.ValidUntil.Add(time.Nanosecond)
		},
		func(s *nodeissuerbinding.Summary) { s.Proposed.Leaf.Leaf[0] ^= 1 },
		func(s *nodeissuerbinding.Summary) {
			s.Evidence.Evidence.Certificate.CA.CADERDigest = contracts.Digest{}
		},
		func(s *nodeissuerbinding.Summary) { s.Evidence.Evidence.Certificate.Leaf.Leaf.SerialLength = 0 },
		func(s *nodeissuerbinding.Summary) { s.Evidence.Evidence.Certificate.Leaf.Leaf.SerialLength = 21 },
		func(s *nodeissuerbinding.Summary) { s.Evidence.Evidence.Certificate.Leaf.Leaf.Serial[0] = 0 },
		func(s *nodeissuerbinding.Summary) { s.Evidence.Evidence.Certificate.Leaf.Leaf.Serial[19] = 1 },
	} {
		s := ownEvidence(good)
		mutate(&s)
		if e := validateEvidence(s); e != ErrInvalidEvidence {
			t.Fatal("invalid defensive summary", e)
		}
		clearEvidence(&s)
	}
}
