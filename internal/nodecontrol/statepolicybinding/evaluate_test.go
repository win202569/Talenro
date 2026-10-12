package statepolicybinding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
)

func fmtDigest(d [32]byte) string   { return hex.EncodeToString(d[:]) }
func shaDigest(raw []byte) [32]byte { return sha256.Sum256(raw) }
func uuidFixture(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, e := uuid.Parse(s)
	if e != nil {
		t.Fatal(e)
	}
	return id
}
func materialRef(r statecontract.ArtifactRef) statematerials.MaterialRef {
	kind := statematerials.RootSet
	switch r.Kind {
	case statecontract.RootRotation:
		kind = statematerials.RootRotation
	case statecontract.Metadata:
		kind = statematerials.Metadata
	case statecontract.DesiredSnapshot:
		kind = statematerials.Desired
	}
	return statematerials.MaterialRef{Kind: kind, Payload: r.Payload, Envelope: r.Envelope}
}
func TestPolicyBindingEvaluate(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		for _, old := range []int{0, n} {
			in := policyBindingFixture(t, n, old)
			checkBindingInput(t, in)
			checkPolicyInput(t, in)
			m, e := Evaluate(context.Background(), in)
			if e != nil {
				t.Fatal(n, old, e)
			}
			if m.summary.Status != ConditionalSelectedPolicyBound || m.summary.SelectedPolicyNoChange != (n == old) {
				t.Fatal("conditional summary")
			}
			m.Clear()
		}
	}
	wrong := policyBindingFixture(t, 1, 0)
	wrong.Policy = policyBindingFixture(t, 0, 0).Policy
	checkBindingInput(t, wrong)
	checkPolicyInput(t, wrong)
	requireBindingEvaluation(t, wrong, ErrBindingMismatch)
	for _, current := range []bool{false, true} {
		in := policyBindingFixture(t, 1, 0)
		if current {
			comp := &in.Binding.Proposed
			ref := replaceRaw(&comp.Materials, comp.Target.Metadata, corruptSignature(t, in.Policy.Batch.MetadataEnvelope))
			comp.Target.Metadata = ref
			in.Binding.Main.Proposed.Metadata.Mark.Material = artifact(ref)
			in.Policy.Batch.MetadataEnvelope = material(comp, ref).Raw
			requireBindingEvaluation(t, in, ErrSignatureMismatch)
		} else {
			ref := in.Binding.Main.Previous.Metadata.Mark.Material
			raw := []byte(nil)
			for _, m := range in.Binding.Previous.Materials {
				if m.Ref.Payload == ref.Payload && m.Ref.Envelope == ref.Envelope {
					raw = m.Raw
				}
			}
			r := replaceRaw(&in.Binding.Previous, materialRef(ref), corruptSignature(t, raw))
			in.Binding.Main.Previous.Metadata.Mark.Material = artifact(r)
			checkBindingInput(t, in)
			checkPolicyInput(t, in)
			requireBindingEvaluation(t, in, nil)
		}
	}
}
func assertPolicyFailure(t *testing.T, in Input, code statepolicy.Code, reason statepolicy.Reason) {
	t.Helper()
	checkBindingInput(t, in)
	c, e := statepolicy.EvaluateModel(context.Background(), in.Policy)
	defer c.Clear()
	var lower *statepolicy.PolicyError
	if c != nil || !errors.As(e, &lower) || lower == nil || lower.Code != code || lower.Reason != reason {
		t.Fatalf("independent policy got %v want %v/%v", e, code, reason)
	}
	m, e := Evaluate(context.Background(), in)
	var pf *PolicyFailure
	if m != nil || !errors.Is(e, ErrPolicyRejected) || !errors.As(e, &pf) || pf == nil || pf.Code != code || pf.Reason != reason {
		t.Fatalf("upper got %v want %v/%v", e, code, reason)
	}
	if e.Error() != ErrPolicyRejected.Error() {
		t.Fatal("dynamic diagnostic")
	}
}
func TestPolicyBindingPolicyFailure(t *testing.T) {
	good := historyBindingFixture(t, false, false)
	checkBindingInput(t, good)
	checkPolicyInput(t, good)
	requireBindingEvaluation(t, good, nil)
	missing := copyPolicyBindingInput(t, good)
	missing.Policy.History = nil
	assertPolicyFailure(t, missing, statepolicy.AssuranceMissing, statepolicy.MissingHistory)
	early := copyPolicyBindingInput(t, good)
	early.Policy.History[0].LastCompositeDeadline = early.Policy.Time.Now
	assertPolicyFailure(t, early, statepolicy.Rejected, statepolicy.HistoryConflict)
	emergency := historyBindingFixture(t, true, false)
	assertPolicyFailure(t, emergency, statepolicy.SecurityFault, statepolicy.EmergencyRevocation)
	removed := historyBindingFixture(t, false, true)
	assertPolicyFailure(t, removed, statepolicy.Rejected, statepolicy.HistoryConflict)
}
func TestPolicyBindingErrorMapping(t *testing.T) {
	ctx := context.Background()
	pairs := [][2]error{{statebinding.ErrInvalidInput, ErrInvalidInput}, {statebinding.ErrLimitExceeded, ErrLimitExceeded}, {statebinding.ErrIncompleteCoverage, ErrIncompleteCoverage}, {statebinding.ErrInvalidMaterial, ErrInvalidMaterial}, {statebinding.ErrMissingMaterial, ErrMissingMaterial}, {statebinding.ErrDuplicateMaterial, ErrDuplicateMaterial}, {statebinding.ErrBindingMismatch, ErrBindingMismatch}, {statebinding.ErrSignatureMismatch, ErrSignatureMismatch}, {statebinding.ErrRollback, ErrRollback}, {statebinding.ErrConflict, ErrConflict}, {statebinding.ErrClearedModel, ErrClearedModel}}
	for _, p := range pairs {
		if e := mapError(ctx, fmt.Errorf("secret: %w", p[0])); e != p[1] {
			t.Fatal(e)
		}
	}
	for _, e := range []error{ErrInvalidInput, ErrLimitExceeded, ErrIncompleteCoverage, ErrInvalidMaterial, ErrMissingMaterial, ErrDuplicateMaterial, ErrBindingMismatch, ErrSignatureMismatch, ErrRollback, ErrConflict, ErrClearedModel, ErrUnsupportedPolicyShape, ErrPolicyRejected, context.Canceled, context.DeadlineExceeded} {
		if got := mapError(ctx, fmt.Errorf("secret: %w", e)); got != e {
			t.Fatal(got)
		}
	}
	for code := statepolicy.Code(1); code <= 4; code++ {
		for reason := statepolicy.Reason(1); reason <= 13; reason++ {
			got := mapError(ctx, fmt.Errorf("secret: %w", &statepolicy.PolicyError{Code: code, Reason: reason}))
			var pf *PolicyFailure
			if !errors.As(got, &pf) || pf.Code != code || pf.Reason != reason || pf.Error() != ErrPolicyRejected.Error() {
				t.Fatal(got)
			}
		}
	}
	for _, e := range []error{errors.New("secret"), (*statepolicy.PolicyError)(nil), fmt.Errorf("wrapped: %w", (*statepolicy.PolicyError)(nil)), &statepolicy.PolicyError{Code: 99, Reason: 1}, &statepolicy.PolicyError{Code: 1, Reason: 99}} {
		if got := mapError(ctx, e); got != ErrInvalidInput {
			t.Fatal(got)
		}
	}
	if mapError(ctx, nil) != nil || mapError(nil, nil) != ErrInvalidInput {
		t.Fatal("nil mapping")
	}
	if got := mapError(ctx, &statepolicy.PolicyError{Code: statepolicy.Canceled, Reason: statepolicy.Cancellation}); got != context.Canceled {
		t.Fatal(got)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if got := mapError(canceled, (*statepolicy.PolicyError)(nil)); got != context.Canceled {
		t.Fatal(got)
	}
	deadline, cancel := context.WithDeadline(ctx, fixtureTime(t, "2000-01-01T00:00:00Z"))
	defer cancel()
	if got := mapError(deadline, errors.New("secret")); got != context.DeadlineExceeded {
		t.Fatal(got)
	}
}
