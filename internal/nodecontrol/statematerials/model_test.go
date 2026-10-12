package statematerials

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func populatedFixture(t *testing.T) Input {
	in := subsetFixture(t)
	rewrapMaterial(t, &in, 3, func(w map[string]any) { w["revoked_key_ids"] = []string{fmt.Sprintf("%064x", 1)} })
	in.Expected[3].Metadata.RevokedIDs = []contracts.Digest{digestText(t, fmt.Sprintf("%064x", 1))}
	in.Expected[4].Desired.MetadataPayload = in.Expected[3].Ref.Payload
	process := obj(t, []byte(`{"adapter":"fixture","capacity_limits":{"connection_limit":"1","cpu_limit_basis_points":"1","cpu_quota_millicores":"100","egress_limit_bps":"1000000","file_descriptor_limit":"64","handshake_limit_per_second":"1","memory_limit_bytes":"67108864","packet_loss_limit_basis_points":"1","queue_limit":"1","task_limit":"32"},"capacity_profile_id":"capacity.v1","capacity_profile_version":"1","lifecycle":"running","probe_profile":"probe.v1","release_id":"release.v1","required":false,"required_metrics":["bytes"],"restart_profile":"restart.v1","slot_id":"slot-a","test_profile":"test.v1"}`))
	rewrapMaterial(t, &in, 4, func(w map[string]any) { w["processes"] = []any{process} })
	in.Expected[4].Desired.SlotIDs = []string{"slot-a"}
	return in
}
func TestSubsetSummary(t *testing.T) {
	want := [6]Obligation{OtherMaterialsAndHistory, SignaturesAndAnchor, DomainSemantics, CompositionRecheck, PhysicalCommit, RuntimeAuthorization}
	for _, all := range []bool{false, true} {
		in := subsetFixture(t)
		if !all {
			in.Materials = in.Materials[:1]
			in.Expected = in.Expected[:1]
		}
		m, e := Evaluate(testContext(), in)
		if e != nil {
			t.Fatal(e)
		}
		s := m.Summary()
		if s.Status != SubsetConsistent || s.Scope != in.Scope || s.MaterialCount != len(in.Materials) || s.Obligations != want {
			t.Fatal(s)
		}
		for j, c := range s.Coverage {
			if c.Kind != MaterialKind(j+1) {
				t.Fatal(c)
			}
			if !all && j > 0 && c.State != NotProvided {
				t.Fatal("false coverage")
			}
		}
		m.Clear()
		if m.Summary() != (Summary{}) {
			t.Fatal("cleared summary")
		}
	}
	var nilModel *MaterialSubsetModel
	if nilModel.Summary() != (Summary{}) {
		t.Fatal("nil summary")
	}
}

type slicePath []int

func nonemptySlices(v reflect.Value, path slicePath, out *[]slicePath) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			nonemptySlices(v.Field(i), append(append(slicePath{}, path...), i), out)
		}
	case reflect.Slice:
		if v.Len() == 0 {
			return
		}
		*out = append(*out, append(slicePath{}, path...))
		if v.Index(0).Kind() == reflect.Struct {
			for i := 0; i < v.Len(); i++ {
				nonemptySlices(v.Index(i), append(append(slicePath{}, path...), i), out)
			}
		}
	}
}
func atPath(v reflect.Value, path slicePath) reflect.Value {
	for _, i := range path {
		if v.Kind() == reflect.Struct {
			v = v.Field(i)
		} else {
			v = v.Index(i)
		}
	}
	return v
}
func independentOrder(in *Input) {
	sort.Slice(in.Materials, func(i, j int) bool {
		return fmt.Sprintf("%d:%x", in.Materials[i].Ref.Kind, in.Materials[i].Ref.Payload) < fmt.Sprintf("%d:%x", in.Materials[j].Ref.Kind, in.Materials[j].Ref.Payload)
	})
	sort.Slice(in.Expected, func(i, j int) bool {
		return fmt.Sprintf("%d:%x", in.Expected[i].Ref.Kind, in.Expected[i].Ref.Payload) < fmt.Sprintf("%d:%x", in.Expected[j].Ref.Kind, in.Expected[j].Ref.Payload)
	})
}
func TestSubsetOwnership(t *testing.T) {
	fixture := populatedFixture(t)
	var paths []slicePath
	nonemptySlices(reflect.ValueOf(fixture), nil, &paths)
	if len(paths) != 12 {
		t.Fatalf("collection coverage %d", len(paths))
	}
	for _, path := range paths {
		t.Run(fmt.Sprint(path), func(t *testing.T) {
			in, expected := populatedFixture(t), populatedFixture(t)
			m, e := Evaluate(testContext(), in)
			if e != nil {
				t.Fatal(e)
			}
			s := atPath(reflect.ValueOf(&in).Elem(), path)
			first := s.Index(0)
			switch first.Kind() {
			case reflect.Uint8:
				first.SetUint(first.Uint() ^ 1)
			case reflect.String:
				first.SetString("changed")
			case reflect.Array:
				first.Index(0).SetUint(first.Index(0).Uint() ^ 1)
			case reflect.Struct:
				if first.Type() == reflect.TypeOf(Material{}) {
					first.FieldByName("Ref").FieldByName("Payload").Index(0).SetUint(2)
				} else {
					first.FieldByName("Epoch").SetUint(20)
				}
			}
			independentOrder(&expected)
			if !reflect.DeepEqual(m.input, expected) {
				t.Fatal("input aliases owner")
			}
			if e := Recheck(testContext(), m, expected); e != nil {
				t.Fatal(e)
			}
			// Independent mutated snapshot confirms Clear never touches caller storage.
			mutated := populatedFixture(t)
			other := atPath(reflect.ValueOf(&mutated).Elem(), path)
			other.Index(0).Set(first)
			var ownedPaths []slicePath
			nonemptySlices(reflect.ValueOf(m.input), nil, &ownedPaths)
			var held []reflect.Value
			for _, p := range ownedPaths {
				held = append(held, atPath(reflect.ValueOf(m.input), p))
			}
			m.Clear()
			m.Clear()
			if !reflect.DeepEqual(in, mutated) {
				t.Fatal("clear modified caller")
			}
			for _, s := range held {
				for j := 0; j < s.Len(); j++ {
					if !s.Index(j).IsZero() {
						t.Fatal("owned collection not cleared")
					}
				}
			}
		})
	}
}
func TestSubsetRecheck(t *testing.T) {
	in := subsetFixture(t)
	m, e := Evaluate(testContext(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	reordered := subsetFixture(t)
	reordered.Materials[0], reordered.Materials[4] = reordered.Materials[4], reordered.Materials[0]
	reordered.Expected[0], reordered.Expected[4] = reordered.Expected[4], reordered.Expected[0]
	if e := Recheck(testContext(), m, reordered); e != nil {
		t.Fatal("order", e)
	}
	fresh := subsetFixture(t)
	fresh.Expected[3].Metadata.RevokedIDs = []contracts.Digest{}
	fresh.Expected[4].Desired.SlotIDs = []string{}
	if e := Recheck(testContext(), m, fresh); e != nil {
		t.Fatal("empty normalization", e)
	}
	fresh = subsetFixture(t)
	outer := obj(t, fresh.Materials[4].Raw)
	outer["signature"] = "AwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	fresh.Materials[4].Raw = canon(t, outer)
	fresh.Materials[4].Ref = envelopeRef(t, Desired, fresh.Materials[4].Raw)
	fresh.Expected[4].Ref = fresh.Materials[4].Ref
	replacement, e := Evaluate(testContext(), fresh)
	if e != nil {
		t.Fatal("valid replacement format", e)
	}
	replacement.Clear()
	if e := Recheck(testContext(), m, fresh); e != ErrReferenceMismatch {
		t.Fatal("envelope substituted", e)
	}
	for i := range in.Materials {
		bad := subsetFixture(t)
		bad.Materials[i].Raw[0] ^= 1
		if e := Recheck(testContext(), m, bad); e == nil {
			t.Fatal("bytes ignored")
		}
	}
	// Two individually well-formed metadata associations cannot rewrite an existing model.
	fresh = subsetFixture(t)
	variant := subsetFixture(t)
	rewrapMaterial(t, &variant, 3, func(w map[string]any) { w["metadata_version"] = "7" })
	variant.Expected[3].Metadata.Version = 7
	fresh.Materials = append(fresh.Materials, variant.Materials[3])
	fresh.Expected = append(fresh.Expected, variant.Expected[3])
	withBoth, e := Evaluate(testContext(), fresh)
	if e != nil {
		t.Fatal(e)
	}
	defer withBoth.Clear()
	fresh.Expected[4].Desired.MetadataPayload = variant.Expected[3].Ref.Payload
	validAssociation, e := Evaluate(testContext(), fresh)
	if e != nil {
		t.Fatal(e)
	}
	validAssociation.Clear()
	if e := Recheck(testContext(), withBoth, fresh); e != ErrReferenceMismatch {
		t.Fatal("association rewritten", e)
	}
	if e := Recheck(testContext(), nil, in); e != ErrClearedModel {
		t.Fatal(e)
	}
	m.Clear()
	if e := Recheck(testContext(), m, in); e != ErrClearedModel {
		t.Fatal(e)
	}
}

type countedContext struct {
	calls, stop int
	failure     error
}

func (c *countedContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *countedContext) Done() <-chan struct{}       { return nil }
func (c *countedContext) Value(any) any               { return nil }
func (c *countedContext) Err() error {
	c.calls++
	if c.stop > 0 && c.calls >= c.stop {
		if c.failure != nil {
			return c.failure
		}
		return context.Canceled
	}
	return nil
}
func TestSubsetCancellation(t *testing.T) {
	in := subsetFixture(t)
	c := &countedContext{}
	m, e := Evaluate(c, in)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	n := c.calls
	for stop := 1; stop <= n; stop++ {
		c := &countedContext{stop: stop}
		r, e := Evaluate(c, in)
		if r != nil || e != context.Canceled {
			t.Fatalf("cancel at %d: %v %v", stop, r, e)
		}
	}
	c = &countedContext{stop: 1, failure: context.DeadlineExceeded}
	if r, e := Evaluate(c, in); r != nil || e != context.DeadlineExceeded {
		t.Fatal(e)
	}
	c = &countedContext{}
	if e := Recheck(c, m, in); e != nil {
		t.Fatal(e)
	}
	for stop := 1; stop <= c.calls; stop++ {
		ctx := &countedContext{stop: stop}
		if e := Recheck(ctx, m, in); e != context.Canceled {
			t.Fatalf("recheck cancel %d: %v", stop, e)
		}
	}
	if e := Recheck(nil, m, in); e != ErrInvalidInput {
		t.Fatal(e)
	}
}
