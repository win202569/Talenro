package statetimebinding

import (
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
	"testing"
)

func evaluated(t *testing.T, in Input) *Model {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if e != nil || m == nil {
		t.Fatal(e)
	}
	t.Cleanup(m.Clear)
	return m
}
func TestTimeBindingEvaluate(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		for _, cp := range []uint64{100, 101} {
			in := timeBindingFixture(t, n, cp)
			checkLowerModels(t, in)
			m := evaluated(t, in)
			if m.Summary().Status != ConditionalTimeEvidenceBound {
				t.Fatal("status")
			}
		}
	}
	in := timeBindingFixture(t, 1, 100)
	desiredUpdate(t, &in)
	checkLowerModels(t, in)
	m := evaluated(t, in)
	if m.Summary().Policy.SelectedPolicyNoChange {
		t.Fatal("desired update discarded")
	}
}
func requireEvaluation(t *testing.T, in Input, want error) {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	defer m.Clear()
	if m != nil || !errors.Is(e, want) {
		t.Fatalf("got %v want %v", e, want)
	}
}
func TestTimeBindingMissingObservationPrecedence(t *testing.T) {
	for _, observation := range []struct {
		name   string
		zero   func(*Input)
		status func(*Input, statetime.ObservationStatus)
	}{
		{"metadata", func(x *Input) { x.Time.Metadata = statetime.CurrentMetadataObservation{} }, func(x *Input, s statetime.ObservationStatus) { x.Time.Metadata.Status = s }},
		{"time", func(x *Input) { x.Time.Time = statetime.TimeObservation{} }, func(x *Input, s statetime.ObservationStatus) { x.Time.Time.Status = s }},
		{"request", func(x *Input) { x.Time.Request = statetime.RequestObservation{} }, func(x *Input, s statetime.ObservationStatus) { x.Time.Request.Status = s }},
		{"checkpoint", func(x *Input) { x.Time.Checkpoint = statetime.CheckpointObservation{} }, func(x *Input, s statetime.ObservationStatus) { x.Time.Checkpoint.Status = s }},
	} {
		t.Run(observation.name, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				change func(*Input)
			}{
				{"zero", observation.zero},
				{"unknown", func(x *Input) { observation.status(x, statetime.Unknown); x.Time.Metadata.Version++ }},
				{"unavailable", func(x *Input) { observation.status(x, statetime.Unavailable); x.Time.Metadata.Version++ }},
				{"rollback", func(x *Input) { observation.status(x, statetime.Rollback); x.Time.Metadata.Version++ }},
				{"out-of-range", func(x *Input) { observation.status(x, 99); x.Time.Metadata.Version++ }},
			} {
				t.Run(tc.name, func(t *testing.T) {
					in := timeBindingFixture(t, 1, 100)
					tc.change(&in)
					before := cloneFixture(t, in)
					low, e := statepolicybinding.Evaluate(context.Background(), in.Policy)
					defer low.Clear()
					if low == nil || e != nil {
						t.Fatal("selected-policy control", e)
					}
					requireEvaluation(t, in, ErrAssuranceMissing)
					if !reflect.DeepEqual(in, before) {
						t.Fatal("caller mutated")
					}
				})
			}
		})
	}
	// Complete selected-policy evaluation still precedes time availability.
	in := attachTime(t, historyBindingFixture(t, false, false), 100)
	in.Policy.Policy.History = nil
	in.Time.Metadata = statetime.CurrentMetadataObservation{}
	low, e := statepolicybinding.Evaluate(context.Background(), in.Policy)
	defer low.Clear()
	var lower *statepolicybinding.PolicyFailure
	if low != nil || !errors.As(e, &lower) || lower.Code != statepolicy.AssuranceMissing || lower.Reason != statepolicy.MissingHistory {
		t.Fatal("policy rejection control", e)
	}
	m, e := Evaluate(context.Background(), in)
	defer m.Clear()
	var high *PolicyFailure
	if m != nil || !errors.As(e, &high) || high.Code != lower.Code || high.Reason != lower.Reason || !errors.Is(e, ErrPolicyRejected) {
		t.Fatal("selected-policy precedence lost", e)
	}
}
func TestTimeBindingLowerRejections(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Input)
		want   error
	}{
		{"missing", func(x *Input) { x.Time.Request.Status = statetime.Unknown }, ErrAssuranceMissing},
		{"format", func(x *Input) { x.Time.AttestationEnvelope = []byte("invalid") }, ErrInvalidInput},
		{"signature", func(x *Input) {
			o := object(t, x.Time.AttestationEnvelope)
			sig := o["signature"].(string)
			if sig[0] == 'A' {
				sig = "B" + sig[1:]
			} else {
				sig = "A" + sig[1:]
			}
			o["signature"] = sig
			x.Time.AttestationEnvelope = canon(t, o)
		}, ErrSignatureMismatch},
		{"unauthorized", func(x *Input) {
			o := object(t, x.Time.AttestationEnvelope)
			o["key_id"] = fmtDigest(d(99))
			x.Time.AttestationEnvelope = canon(t, o)
		}, ErrUnauthorizedKey},
		{"nonce", func(x *Input) { x.Time.Request.ExpectedNonce[0] = 1 }, ErrBindingMismatch},
		{"rollback", func(x *Input) { signAttestation(t, x, 99, [32]byte{}) }, ErrCheckpointRollback},
		{"expired", func(x *Input) {
			x.Time.Time.Now = x.Time.Time.Now.Add(5 * 60 * 1000000000)
			x.Policy.Policy.Time.Now = x.Time.Time.Now
		}, ErrExpired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := timeBindingFixture(t, 1, 100)
			tc.change(&in)
			low, e := statetime.EvaluateModel(context.Background(), in.Time)
			defer low.Clear()
			if low != nil || mapError(context.Background(), e) != tc.want {
				t.Fatal("lower control", e)
			}
			requireEvaluation(t, in, tc.want)
		})
	}
	for _, tc := range []struct {
		incident bool
		change   func(*Input)
		code     statepolicy.Code
		reason   statepolicy.Reason
	}{
		{false, func(x *Input) { x.Policy.Policy.History = nil }, statepolicy.AssuranceMissing, statepolicy.MissingHistory},
		{false, func(x *Input) { x.Policy.Policy.History[0].LastCompositeDeadline = x.Policy.Policy.Time.Now }, statepolicy.Rejected, statepolicy.HistoryConflict},
		{true, func(x *Input) {}, statepolicy.SecurityFault, statepolicy.EmergencyRevocation},
	} {
		in := attachTime(t, historyBindingFixture(t, tc.incident, false), 100)
		tc.change(&in)
		low, e := statepolicybinding.Evaluate(context.Background(), in.Policy)
		defer low.Clear()
		var lower *statepolicybinding.PolicyFailure
		if low != nil || !errors.As(e, &lower) || lower.Code != tc.code || lower.Reason != tc.reason {
			t.Fatal("policy rejection control", e)
		}
		m, e := Evaluate(context.Background(), in)
		defer m.Clear()
		var high *PolicyFailure
		if m != nil || !errors.As(e, &high) || high.Code != tc.code || high.Reason != tc.reason || !errors.Is(e, ErrPolicyRejected) {
			t.Fatal("policy failure lost", e)
		}
	}
}
func TestTimeBindingSummary(t *testing.T) {
	in := timeBindingFixture(t, 1, 101)
	s := evaluated(t, in).Summary()
	p := checkLowerModels(t, in)
	tm := timeEvidence(t, in)
	if !reflect.DeepEqual(s.Policy, p) || s.Time != tm || !s.Policy.SelectedPolicyNoChange || !s.Time.AdvanceRequired {
		t.Fatal("nested evidence")
	}
	want := ObservationBinding{in.Policy.Policy.Old.StateID, in.Time.Metadata.SourceID, in.Time.Metadata.ObservationID, in.Time.Request.ObservationID, in.Time.Checkpoint.SourceID, in.Time.Checkpoint.ObservationID}
	if s.Observations != want {
		t.Fatal("identities")
	}
	if s.Obligations != ([5]Obligation{ObservationAndBaseProvenance, TrustedTimeAndTLS, AtomicRequestConsumption, DurableCheckpointCommit, FullMainStateAndRuntimeAuthorization}) {
		t.Fatal("obligations")
	}
	until := p.ValidUntil
	if tm.ValidUntil.Before(until) {
		until = tm.ValidUntil
	}
	if s.ValidUntil != until {
		t.Fatal("minimum")
	}
	for _, method := range []string{"State", "Ready", "Commit", "Authorize", "Consumed", "WholeStateNoChange"} {
		if _, ok := reflect.TypeOf(&Model{}).MethodByName(method); ok {
			t.Fatal("capability", method)
		}
	}
}
func TestTimeBindingRequestSemantics(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	a := evaluated(t, in)
	b := evaluated(t, in)
	if !reflect.DeepEqual(a.Summary(), b.Summary()) || a.Summary().Time.Nonce != ([32]byte{}) {
		t.Fatal("zero nonce/repeat")
	}
	changed := cloneFixture(t, in)
	changed.Time.Request.RequestID = d(96)
	changed.Time.Request.ObservationID = d(97)
	evaluated(t, changed)
}

func TestTimeBindingInheritedUnsupportedShapes(t *testing.T) {
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy = historicalPolicyBindingFixture(t) },
		func(x *Input) { x.Policy.Policy.Old = nil; x.Policy.Policy.Deployment.FreshInstall = true },
		func(x *Input) {
			x.Policy.Binding.Main.Proposed.Scope.TimeFloor.Present = true
			x.Policy.Binding.Main.Proposed.Scope.TimeFloor.UnixSeconds = x.Policy.Policy.Old.TimeFloor.Unix()
		},
	} {
		in := timeBindingFixture(t, 1, 100)
		change(&in)
		low, e := statepolicybinding.Evaluate(context.Background(), in.Policy)
		defer low.Clear()
		if low != nil || !errors.Is(e, statepolicybinding.ErrUnsupportedPolicyShape) {
			t.Fatal("lower unsupported shape control", e)
		}
		requireEvaluation(t, in, ErrUnsupportedPolicyShape)
	}
}
