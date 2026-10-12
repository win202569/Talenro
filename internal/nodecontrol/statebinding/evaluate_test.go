package statebinding

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func evaluatedBinding(t *testing.T, in Input) *BindingModel {
	t.Helper()
	model, err := Evaluate(context.Background(), in)
	if err != nil || model == nil {
		t.Fatalf("missing binding: %v", err)
	}
	t.Cleanup(model.Clear)
	return model
}
func wantBindingError(t *testing.T, in Input, want error) {
	t.Helper()
	model, err := Evaluate(context.Background(), in)
	if model != nil {
		defer model.Clear()
	}
	if !errors.Is(err, want) || model != nil {
		t.Fatalf("want %v, model=%v err=%v", want, model != nil, err)
	}
}
func changedSignature(t *testing.T, raw []byte, field string) []byte {
	t.Helper()
	o := object(t, raw)
	if field == "signature" {
		b, err := base64.RawURLEncoding.DecodeString(o[field].(string))
		if err != nil {
			t.Fatal(err)
		}
		b[0] ^= 1
		o[field] = base64.RawURLEncoding.EncodeToString(b)
	} else {
		share := o[field].([]any)[0].(map[string]any)
		b, err := base64.RawURLEncoding.DecodeString(share["signature"].(string))
		if err != nil {
			t.Fatal(err)
		}
		b[0] ^= 1
		share["signature"] = base64.RawURLEncoding.EncodeToString(b)
	}
	return encode(t, o)
}

func TestBindingEvaluate(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) { evaluatedBinding(t, bindingFixture(t, n)) })
	}
	t.Run("previous_ne", func(t *testing.T) {
		in := bindingFixture(t, 1)
		neverPrevious(&in)
		checkBindingLayers(t, in)
		evaluatedBinding(t, in)
	})
	t.Run("dual_envelope", func(t *testing.T) {
		in := bindingFixture(t, 1)
		old := statecomposition.Input{Materials: in.Previous}
		r := in.Proposed.Target.Metadata
		r = replaceEnvelope(t, &old, r, extraShare(t, material(&old, r).Raw, "signatures"))
		in.Previous = old.Materials
		in.Main.Previous.Metadata.Mark.Material = artifact(r)
		checkBindingLayers(t, in)
		evaluatedBinding(t, in)
	})
	t.Run("separated_chains", func(t *testing.T) { in := separatedFixture(t); checkBindingLayers(t, in); evaluatedBinding(t, in) })
	t.Run("previous_root_keys", func(t *testing.T) {
		in := bindingFixture(t, 1)
		in.Main.Previous.Root.KeyIDs = []contracts.Digest{d(99)}
		checkBindingLayers(t, in)
		wantBindingError(t, in, ErrBindingMismatch)
	})
	t.Run("transition_envelope", func(t *testing.T) {
		in := separatedFixture(t)
		in.Main.RootTransition[0].Material.Envelope[0] ^= 1
		checkBindingLayers(t, in)
		wantBindingError(t, in, ErrBindingMismatch)
	})
	t.Run("missing_transition_start", func(t *testing.T) {
		in := missingTransitionStartFixture(t)
		proof := checkBindingLayers(t, in)
		if proof.FinalRoot.Payload == in.Main.Previous.Root.Mark.Material.Payload || proof.FinalRoot.Payload == in.Main.Proposed.Root.Mark.Material.Payload {
			t.Fatal("fixture lost historical authorizer")
		}
		if err := bindProjection(context.Background(), in); err != nil {
			t.Fatal("projection rejected before transition lookup", err)
		}
		wantBindingError(t, in, ErrMissingMaterial)
	})
	for _, tc := range []struct {
		name string
		edit func(*Input)
		want error
	}{
		{"l0_coverage", func(in *Input) { in.Main.Proposed.Checkpoint.Header.Coverage = statecontract.NeverEstablished }, ErrIncompleteCoverage},
		{"l0_rollback", func(in *Input) { in.Main.Proposed.Checkpoint.Sequence-- }, ErrRollback},
		{"l0_conflict", func(in *Input) { in.Main.Proposed.Root.Mark.Sequence++ }, ErrConflict},
		{"old_raw", func(in *Input) { in.Previous.Materials[0].Raw[0] = '!' }, ErrInvalidMaterial},
		{"old_missing_relation", func(in *Input) {
			in.Previous.Materials = in.Previous.Materials[1:]
			in.Previous.Expected = in.Previous.Expected[1:]
		}, ErrMissingMaterial},
		{"old_duplicate", func(in *Input) {
			in.Previous.Materials = append(in.Previous.Materials, in.Previous.Materials[0])
			in.Previous.Expected = append(in.Previous.Expected, in.Previous.Expected[0])
		}, ErrDuplicateMaterial},
		{"old_expected", func(in *Input) { in.Previous.Expected[0].Sequence++ }, ErrBindingMismatch},
		{"proposed_projection", func(in *Input) { in.Main.Proposed.Metadata.KeyIDs = []contracts.Digest{d(98)} }, ErrBindingMismatch},
		{"budget", func(in *Input) { in.Previous.Expected = make([]statematerials.Expectation, 129) }, ErrLimitExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) { in := bindingFixture(t, 1); tc.edit(&in); wantBindingError(t, in, tc.want) })
	}
	t.Run("target_signature", func(t *testing.T) {
		in := bindingFixture(t, 1)
		r := in.Proposed.Target.Desired
		in.Proposed.Target.Desired = replaceEnvelope(t, &in.Proposed, r, changedSignature(t, material(&in.Proposed, r).Raw, "signature"))
		in.Main.Proposed.Desired.Mark.Material = artifact(in.Proposed.Target.Desired)
		wantBindingError(t, in, ErrSignatureMismatch)
	})
	t.Run("new_ne", func(t *testing.T) {
		in := bindingFixture(t, 1)
		neverPrevious(&in)
		in.Main.Proposed = in.Main.Previous
		checkBindingLayers(t, in)
		wantBindingError(t, in, ErrInvalidInput)
	})
	var nilModel *BindingModel
	nilModel.Clear()
	zero := &BindingModel{}
	zero.Clear()
	zero.Clear()
	if model, err := Evaluate(nil, bindingFixture(t, 1)); err != ErrInvalidInput || model != nil {
		t.Fatal("nil ctx", err)
	}
}

func TestBindingErrorMapping(t *testing.T) {
	rows := []struct{ in, want error }{
		{nil, nil}, {errors.New("raw secret / external dynamic error"), ErrInvalidInput},
		{context.Canceled, context.Canceled}, {context.DeadlineExceeded, context.DeadlineExceeded},
		{statecontract.ErrInvalidInput, ErrInvalidInput}, {statecontract.ErrIncompleteCoverage, ErrIncompleteCoverage}, {statecontract.ErrBindingMismatch, ErrBindingMismatch}, {statecontract.ErrRollback, ErrRollback}, {statecontract.ErrConflict, ErrConflict}, {statecontract.ErrStaleModel, ErrClearedModel},
		{&statecontract.ContractError{Domain: statecontract.DomainRoot, Cause: statecontract.ErrRollback}, ErrRollback},
		{statematerials.ErrInvalidInput, ErrInvalidInput}, {statematerials.ErrLimitExceeded, ErrLimitExceeded}, {statematerials.ErrInvalidMaterial, ErrInvalidMaterial}, {statematerials.ErrMissingMaterial, ErrMissingMaterial}, {statematerials.ErrDuplicateMaterial, ErrDuplicateMaterial}, {statematerials.ErrReferenceMismatch, ErrBindingMismatch}, {statematerials.ErrClearedModel, ErrClearedModel},
		{statecomposition.ErrInvalidInput, ErrInvalidInput}, {statecomposition.ErrLimitExceeded, ErrLimitExceeded}, {statecomposition.ErrInvalidMaterial, ErrInvalidMaterial}, {statecomposition.ErrMissingMaterial, ErrMissingMaterial}, {statecomposition.ErrDuplicateMaterial, ErrDuplicateMaterial}, {statecomposition.ErrBindingMismatch, ErrBindingMismatch}, {statecomposition.ErrSignatureMismatch, ErrSignatureMismatch}, {statecomposition.ErrClearedModel, ErrClearedModel},
	}
	for _, e := range []error{ErrInvalidInput, ErrLimitExceeded, ErrIncompleteCoverage, ErrInvalidMaterial, ErrMissingMaterial, ErrDuplicateMaterial, ErrBindingMismatch, ErrSignatureMismatch, ErrRollback, ErrConflict, ErrClearedModel} {
		rows = append(rows, struct{ in, want error }{e, e})
	}
	for i, row := range rows {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			for _, wrapped := range []bool{false, true} {
				err := row.in
				if wrapped && err != nil {
					err = fmt.Errorf("untrusted wrapper: %w", err)
				}
				got := mapError(context.Background(), err)
				if got != row.want || !errors.Is(got, row.want) {
					t.Fatalf("got %v want %v", got, row.want)
				}
			}
		})
	}
	if err := mapError(nil, nil); err != ErrInvalidInput {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := mapError(canceled, statecontract.ErrRollback); err != context.Canceled {
		t.Fatal(err)
	}
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if err := mapError(expired, statecontract.ErrConflict); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}

func TestBindingProofStrength(t *testing.T) {
	in := separatedFixture(t)
	old := statecomposition.Input{Materials: in.Previous}
	oldRef := in.Proposed.Target.Metadata
	oldRef = replaceEnvelope(t, &old, oldRef, changedSignature(t, material(&old, oldRef).Raw, "signatures"))
	in.Previous = old.Materials
	in.Main.Previous.Metadata.Mark.Material = artifact(oldRef)
	currentMeta := in.Main.Proposed.Metadata.Mark.Material
	metaRef := statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: currentMeta.Payload, Envelope: currentMeta.Envelope}
	metaRef = replaceEnvelope(t, &in.Proposed, metaRef, changedSignature(t, material(&in.Proposed, metaRef).Raw, "signatures"))
	in.Main.Proposed.Metadata.Mark.Material = artifact(metaRef)
	rotation := in.Main.RootTransition[0].Material
	rr := statematerials.MaterialRef{Kind: statematerials.RootRotation, Payload: rotation.Payload, Envelope: rotation.Envelope}
	rr = replaceEnvelope(t, &in.Proposed, rr, changedSignature(t, material(&in.Proposed, rr).Raw, "current_signatures"))
	in.Main.RootTransition[0].Material = artifact(rr)
	checkBindingLayers(t, in)
	model := evaluatedBinding(t, in)
	if model.contractSummary.Status != statecontract.ContractShapeValid || model.previousSummary.Status != statematerials.SubsetConsistent || model.proposedSummary.Status != statecomposition.ConditionalTargetVerified {
		t.Fatal("lower summaries changed")
	}
	foundMeta, foundRotation, foundDesired := false, false, false
	for _, c := range model.proposedSummary.Coverage {
		switch c.Ref {
		case metaRef:
			foundMeta = true
			if c.Status != statecomposition.MaterialOnly {
				t.Fatal("current metadata promoted")
			}
		case rr:
			foundRotation = true
			if c.Status != statecomposition.MaterialOnly {
				t.Fatal("current transition promoted")
			}
		case in.Proposed.Target.Desired:
			foundDesired = true
			if c.Status != statecomposition.SignatureChecked {
				t.Fatal("selected desired proof lost")
			}
		}
	}
	if !foundMeta || !foundRotation || !foundDesired {
		t.Fatal("missing coverage")
	}
	if in.Main.Previous.Scope.TrustAnchorID == in.Proposed.Anchor.Payload {
		t.Fatal("opaque fixture accidentally equal")
	}
	in.Main.Previous.Metadata.History.Payload = d(90)
	in.Main.Proposed.Metadata.History.Payload = d(90)
	in.Main.Previous.Metadata.RevokedSetDigest = d(91)
	in.Main.Proposed.Metadata.RevokedSetDigest = d(91)
	evaluatedBinding(t, in) // These declarations are saved, not given invented wire equivalences.
	broken := copyBindingInput(t, in)
	material(&broken.Proposed, metaRef).Raw[0] = '!'
	wantBindingError(t, broken, ErrInvalidMaterial)
}
