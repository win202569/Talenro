package statebinding

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func TestBindingSummary(t *testing.T) {
	in := richBindingFixture(t)
	lower := checkBindingLayers(t, in)
	model := evaluatedBinding(t, in)
	s := model.Summary()
	if s.Status != ConditionalProjectionBound || !reflect.DeepEqual(s.Proposed, lower) || s.Contract.Status != statecontract.ContractShapeValid || s.Previous.Status != statematerials.SubsetConsistent {
		t.Fatal("proof changed")
	}
	want := [8]Obligation{ExternalAnchorAndPlatformProvenance, PreviousSignaturesAndUnselectedHistory, UnverifiedHistoryAndSetDigests, DomainAndCurrentTime, FullStateMaterialClosure, FullStateComposition, PhysicalCommit, RuntimeAuthorization}
	if s.Obligations != want {
		t.Fatal("missing responsibilities")
	}
	contract, err := statecontract.EvaluateModel(context.Background(), in.Main)
	if err != nil {
		t.Fatal(err)
	}
	defer contract.Clear()
	old, err := statematerials.Evaluate(context.Background(), in.Previous)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Clear()
	if s.Contract != contract.Summary() || s.Previous != old.Summary() {
		t.Fatal("lower obligations altered")
	}
	for _, m := range []*BindingModel{nil, {}, model} {
		if m == model {
			model.Clear()
		}
		if !reflect.DeepEqual(m.Summary(), Summary{}) || Recheck(context.Background(), m, in) != ErrClearedModel {
			t.Fatal("cleared/zero result")
		}
	}
}

func TestBindingModelOwnership(t *testing.T) {
	for _, shared := range []bool{false, true} {
		t.Run(fmt.Sprint(shared), func(t *testing.T) {
			in := richBindingFixture(t)
			if shared {
				in = bindingFixture(t, 1)
				in.Proposed.Materials = in.Previous
			}
			checkBindingLayers(t, in)
			model := evaluatedBinding(t, in)
			ownerBefore := string(jsonInput(t, model.input))
			summaryBefore := model.Summary()
			summaryExpected := summaryBeforeCopy(t, summaryBefore)
			if !shared {
				for _, side := range []string{"Previous", "Proposed"} {
					for _, c := range mainCollections() {
						if inputField(reflect.ValueOf(in.Main), side+"."+c.path).Len() == 0 {
							t.Fatal("fixture missing L0 collection", side, c.path)
						}
					}
				}
				if len(in.Main.RootTransition) != 1 {
					t.Fatal("fixture missing transition")
				}
			}
			aliases := mutableSlices(reflect.ValueOf(model.input))
			summaryAliases := []reflect.Value{reflect.ValueOf(model.proposedSummary.Target.Links), reflect.ValueOf(model.proposedSummary.Coverage)}
			callerAliases := mutableSlices(reflect.ValueOf(in))
			for i := len(callerAliases) - 1; i >= 0; i-- {
				callerAliases[i].Clear()
			}
			if string(jsonInput(t, model.input)) != ownerBefore || !reflect.DeepEqual(model.Summary(), summaryBefore) {
				t.Fatal("borrowed caller data")
			}
			detached := model.Summary()
			if len(detached.Proposed.Target.Links) > 0 {
				detached.Proposed.Target.Links[0].Rotation.Payload[0] ^= 1
			}
			detached.Proposed.Coverage[0].Ref.Payload[0] ^= 1
			if !reflect.DeepEqual(model.Summary(), summaryBefore) {
				t.Fatal("summary exposes owner")
			}
			callerAfter := string(jsonInput(t, in))
			model.Clear()
			model.Clear()
			for _, v := range append(aliases, summaryAliases...) {
				for i := 0; i < v.Len(); i++ {
					if !v.Index(i).IsZero() {
						t.Fatal("owned alias not cleared", v.Type())
					}
				}
			}
			if string(jsonInput(t, in)) != callerAfter || !reflect.DeepEqual(summaryBefore, summaryExpected) {
				t.Fatal("caller/detached summary cleared")
			}
			// Retained detached data remains nonzero and independent after owner Clear.
			if summaryBefore.Proposed.Target.Desired.Payload == (contracts.Digest{}) || summaryBefore.Proposed.Coverage[0].Ref.Payload == (contracts.Digest{}) {
				t.Fatal("detached result cleared")
			}
		})
	}
}
func summaryBeforeCopy(t *testing.T, s Summary) Summary {
	t.Helper()
	out := s
	out.Proposed.Target.Links = append([]statecomposition.RotationLink(nil), s.Proposed.Target.Links...)
	out.Proposed.Coverage = append([]statecomposition.Coverage(nil), s.Proposed.Coverage...)
	return out
}

func TestBindingRecheck(t *testing.T) {
	in := bindingFixture(t, 2)
	old := evaluatedBinding(t, in)
	if err := Recheck(context.Background(), old, in); err != nil {
		t.Fatal(err)
	}
	reorder := copyBindingInput(t, in)
	for _, table := range []*statematerials.Input{&reorder.Previous, &reorder.Proposed.Materials} {
		for i, j := 0, len(table.Materials)-1; i < j; i, j = i+1, j-1 {
			table.Materials[i], table.Materials[j] = table.Materials[j], table.Materials[i]
		}
		for i, j := 0, len(table.Expected)-1; i < j; i, j = i+1, j-1 {
			table.Expected[i], table.Expected[j] = table.Expected[j], table.Expected[i]
		}
		for i := range table.Expected {
			e := &table.Expected[i]
			if len(e.Root.KeyIDs) == 0 {
				e.Root.KeyIDs = []contracts.Digest{}
			}
			if len(e.Metadata.KeyIDs) == 0 {
				e.Metadata.KeyIDs = []contracts.Digest{}
			}
			if len(e.Metadata.RevokedIDs) == 0 {
				e.Metadata.RevokedIDs = []contracts.Digest{}
			}
			if len(e.Desired.SlotIDs) == 0 {
				e.Desired.SlotIDs = []string{}
			}
		}
	}
	reorder.Main.RootTransition = []statecontract.RotationRef{}
	reorder.Main.Previous.Recovery.Local = []statecontract.LocalFaultBinding{}
	if err := Recheck(context.Background(), old, reorder); err != nil {
		t.Fatal("normalized identity", err)
	}
	reorder.Proposed.Target.Links[0], reorder.Proposed.Target.Links[1] = reorder.Proposed.Target.Links[1], reorder.Proposed.Target.Links[0]
	if err := Recheck(context.Background(), old, reorder); err == nil {
		t.Fatal("ordered chain accepted")
	}
	// Every complete L0 leaf (including unprojected ones) is part of identity.
	for _, side := range []string{"Previous", "Proposed"} {
		for _, path := range scalarPaths(reflect.ValueOf(in.Main).FieldByName(side), nil) {
			t.Run(side+"/"+fmt.Sprint(path), func(t *testing.T) {
				changed := copyBindingInput(t, in)
				v := reflect.ValueOf(&changed.Main).Elem().FieldByName(side)
				for _, i := range path {
					v = v.Field(i)
				}
				changeValue(v)
				if err := Recheck(context.Background(), old, changed); err == nil {
					t.Fatal("main field omitted", path)
				}
			})
		}
	}
	if err := Recheck(nil, old, in); err != ErrInvalidInput {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Recheck(canceled, nil, in); err != context.Canceled {
		t.Fatal("cancel priority", err)
	}
}
func scalarPaths(v reflect.Value, path []int) [][]int {
	if v.Kind() != reflect.Struct {
		return [][]int{append([]int(nil), path...)}
	}
	var out [][]int
	for i := 0; i < v.NumField(); i++ {
		out = append(out, scalarPaths(v.Field(i), append(append([]int(nil), path...), i))...)
	}
	return out
}
func changeValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(v.String() + "-changed")
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Array:
		v.Index(0).SetUint(v.Index(0).Uint() ^ 1)
	case reflect.Slice:
		if v.Len() == 0 {
			v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		} else {
			v.Index(0).Set(reflect.Zero(v.Type().Elem()))
		}
	case reflect.Int64:
		v.SetInt(v.Int() + 1)
	default:
		v.SetUint(v.Uint() + 1)
	}
}

func TestBindingUnmappedRecheck(t *testing.T) {
	for _, mode := range []string{"history", "revoked_digest", "trust_anchor", "deployment", "old_envelope", "new_envelope", "old_extra", "new_unselected", "coherent_authorizer_rewrite"} {
		t.Run(mode, func(t *testing.T) {
			original := currentMetadataFixture(t, false)
			oldTable := statecomposition.Input{Materials: original.Previous, Target: original.Proposed.Target}
			extraOld := compatibleMetadata(t, &oldTable, 9, false)
			original.Previous = oldTable.Materials
			extraNew := compatibleMetadata(t, &original.Proposed, 9, false)
			old := evaluatedBinding(t, original)
			replacement := copyBindingInput(t, original)
			switch mode {
			case "history":
				replacement.Main.Previous.Metadata.History.Payload = d(90)
				replacement.Main.Proposed.Metadata.History.Payload = d(90)
			case "revoked_digest":
				replacement.Main.Previous.Metadata.RevokedSetDigest = d(91)
				replacement.Main.Proposed.Metadata.RevokedSetDigest = d(91)
			case "trust_anchor":
				replacement.Main.Previous.Scope.TrustAnchorID = d(92)
				replacement.Main.Proposed.Scope.TrustAnchorID = d(92)
			case "deployment":
				replacement.Main.Previous.Scope.DeploymentID = d(93)
				replacement.Main.Proposed.Scope.DeploymentID = d(93)
			case "old_envelope":
				temp := statecomposition.Input{Materials: replacement.Previous}
				r := replacement.Main.Previous.Metadata.Mark.Material
				ref := statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: r.Payload, Envelope: r.Envelope}
				nr := replaceEnvelope(t, &temp, ref, extraShare(t, material(&temp, ref).Raw, "signatures"))
				replacement.Previous = temp.Materials
				replacement.Main.Previous.Metadata.Mark.Material = artifact(nr)
			case "new_envelope":
				r := replacement.Proposed.Target.Metadata
				replacement.Proposed.Target.Metadata = replaceEnvelope(t, &replacement.Proposed, r, extraShare(t, material(&replacement.Proposed, r).Raw, "signatures"))
			case "old_extra":
				temp := statecomposition.Input{Materials: replacement.Previous}
				replaceEnvelope(t, &temp, extraOld, extraShare(t, material(&temp, extraOld).Raw, "signatures"))
				replacement.Previous = temp.Materials
			case "new_unselected":
				replaceEnvelope(t, &replacement.Proposed, extraNew, extraShare(t, material(&replacement.Proposed, extraNew).Raw, "signatures"))
			case "coherent_authorizer_rewrite":
				r := compatibleMetadata(t, &replacement.Proposed, 7, false)
				replacement.Proposed.Target.Metadata = r
				expected(&replacement.Proposed, replacement.Proposed.Target.Desired).Desired.MetadataPayload = r.Payload
				temp := statecomposition.Input{Materials: replacement.Previous, Target: original.Proposed.Target}
				compatibleMetadata(t, &temp, 7, false)
				expected(&temp, temp.Target.Desired).Desired.MetadataPayload = r.Payload
				replacement.Previous = temp.Materials
				for _, x := range []*statecontract.MainStateContentModel{&replacement.Main.Previous, &replacement.Main.Proposed} {
					x.Desired.MetadataPayload = r.Payload
					for i := range x.LKG.Entries {
						x.LKG.Entries[i].MetadataPayload = r.Payload
					}
				}
				if replacement.Proposed.Target.Desired != original.Proposed.Target.Desired {
					t.Fatal("desired wire changed")
				}
			}
			checkBindingLayers(t, replacement)
			evaluatedBinding(t, replacement)
			if err := Recheck(context.Background(), old, replacement); !errors.Is(err, ErrBindingMismatch) {
				t.Fatal("full identity", err)
			}
		})
	}
}

type bindingCheckContext struct {
	context.Context
	calls, failAt, panicAt int
	cause                  error
}

func (c *bindingCheckContext) Err() error {
	c.calls++
	if c.panicAt > 0 && c.calls == c.panicAt {
		panic("binding context probe")
	}
	if c.failAt > 0 && c.calls >= c.failAt {
		return c.cause
	}
	return nil
}
func bindingCheckCount(t *testing.T, in Input) int {
	t.Helper()
	ctx := &bindingCheckContext{Context: context.Background()}
	m, err := Evaluate(ctx, in)
	if err != nil || m == nil {
		t.Fatal(err)
	}
	m.Clear()
	return ctx.calls
}
func TestBindingCancellation(t *testing.T) {
	in := bindingFixture(t, 0)
	before := string(jsonInput(t, in))
	count := bindingCheckCount(t, in)
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		for n := 1; n <= count; n++ {
			t.Run(fmt.Sprintf("%v/%d", cause, n), func(t *testing.T) {
				ctx := &bindingCheckContext{Context: context.Background(), failAt: n, cause: cause}
				m, err := Evaluate(ctx, in)
				if m != nil {
					defer m.Clear()
				}
				if m != nil || !errors.Is(err, cause) {
					t.Fatalf("check %d: model=%v err=%v", n, m != nil, err)
				}
				if string(jsonInput(t, in)) != before {
					t.Fatal("caller modified on cancel")
				}
			})
		}
	}
}
func TestBindingLateCancellation(t *testing.T) {
	in := bindingFixture(t, 1)
	count := bindingCheckCount(t, in)
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		ctx := &bindingCheckContext{Context: context.Background(), failAt: count, cause: cause}
		m, err := Evaluate(ctx, in)
		if m != nil {
			defer m.Clear()
		}
		if m != nil || err != cause {
			t.Fatal("last check delivered owner", err)
		}
	}
	old := evaluatedBinding(t, in)
	ctx := &bindingCheckContext{Context: context.Background()}
	if err := Recheck(ctx, old, in); err != nil {
		t.Fatal(err)
	}
	total := ctx.calls
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		ctx = &bindingCheckContext{Context: context.Background(), failAt: total, cause: cause}
		if err := Recheck(ctx, old, in); err != cause {
			t.Fatal("late Recheck", err)
		}
	}
}
func TestBindingPanicCleanup(t *testing.T) {
	in := bindingFixture(t, 0)
	before := string(jsonInput(t, in))
	count := bindingCheckCount(t, in)
	// Caller bytes are observable; unreachable lower owners are not claimed here.
	for n := 2; n <= count; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			ctx := &bindingCheckContext{Context: context.Background(), panicAt: n}
			panicked := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if r != "binding context probe" {
							t.Fatalf("wrong panic %v", r)
						}
						panicked = true
					}
				}()
				m, err := Evaluate(ctx, in)
				if m != nil {
					m.Clear()
				}
				t.Fatalf("panic swallowed: %v", err)
			}()
			if !panicked || string(jsonInput(t, in)) != before {
				t.Fatal("panic/caller contract")
			}
		})
	}
}
