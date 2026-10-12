package statecomposition

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
	"testing"
)

func evaluated(t *testing.T, in Input) *CompositionModel {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	requireOutcome(t, m, e, nil)
	return m
}
func TestCompositionSummary(t *testing.T) {
	for _, n := range []int{0, 1, 2, 32} {
		in := fixture(t, n)
		if n < 32 {
			compatibleMetadata(t, &in, 7, false)
			disconnected(t, &in)
		}
		m := evaluated(t, in)
		s := m.Summary()
		want := [7]Obligation{ExternalAnchorProvenance, UnselectedSignaturesAndHistory, DomainAndCurrentTime, FullStateMaterialClosure, FullStateComposition, PhysicalCommit, RuntimeAuthorization}
		if s.Status != ConditionalTargetVerified || s.Obligations != want || len(s.Coverage) != len(in.Materials.Materials) || s.Scope != in.Materials.Scope || s.Anchor != (stateverify.ArtifactIdentity{PayloadDigest: in.Anchor.Payload, EnvelopeDigest: in.Anchor.Envelope}) || s.FinalRoot != tailRef(in) || !reflect.DeepEqual(s.Target, in.Target) {
			t.Fatal("summary coordinates/obligations", n)
		}
		subset, e := statematerials.Evaluate(context.Background(), in.Materials)
		if e != nil {
			t.Fatal(e)
		}
		if s.Subset != subset.Summary() {
			t.Fatal("subset upgraded")
		}
		subset.Clear()
		statuses := map[statematerials.MaterialRef]ProofStatus{in.Target.AnchorRoot: AnchorAssumed, in.Target.Metadata: SignatureChecked, in.Target.Desired: SignatureChecked}
		for _, l := range in.Target.Links {
			statuses[l.Rotation] = SignatureChecked
			statuses[l.NewRoot] = ChainDerived
		}
		seen := map[statematerials.MaterialRef]bool{}
		for i, c := range s.Coverage {
			status, ok := statuses[c.Ref]
			if !ok {
				status = MaterialOnly
			}
			if c.Status != status || seen[c.Ref] || i > 0 && !refLess(s.Coverage[i-1].Ref, c.Ref) {
				t.Fatal("coverage", n, i, c.Status, status)
			}
			seen[c.Ref] = true
		}
		for _, m := range in.Materials.Materials {
			if !seen[m.Ref] {
				t.Fatal("missing coverage")
			}
		}
	}
}
func addSlot(t *testing.T, in *Input) {
	p := object(t, payload(t, material(in, in.Target.Desired).Raw))
	p["processes"] = []any{object(t, []byte(`{"adapter":"fixture","capacity_limits":{"connection_limit":"1","cpu_limit_basis_points":"1","cpu_quota_millicores":"100","egress_limit_bps":"1000000","file_descriptor_limit":"64","handshake_limit_per_second":"1","memory_limit_bytes":"67108864","packet_loss_limit_basis_points":"1","queue_limit":"1","task_limit":"32"},"capacity_profile_id":"capacity.v1","capacity_profile_version":"1","lifecycle":"running","probe_profile":"probe.v1","release_id":"release.v1","required":false,"required_metrics":["bytes"],"restart_profile":"restart.v1","slot_id":"slot-a","test_profile":"test.v1"}`))}
	in.Target.Desired = replaceEnvelope(t, in, in.Target.Desired, mint(t, "desired", p, private(2)))
	expected(in, in.Target.Desired).Desired.SlotIDs = []string{"slot-a"}
}
func TestCompositionModelOwnership(t *testing.T) {
	in := fixture(t, 1)
	extra := compatibleMetadata(t, &in, 7, true)
	addSlot(t, &in)
	m := evaluated(t, in)
	owned, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer clearInput(&owned)
	s := m.Summary()
	saved := m.Summary()
	in.Anchor.Raw[0] ^= 1
	for i := range in.Materials.Materials {
		in.Materials.Materials[i].Raw[0] ^= 1
	}
	expected(&in, in.Target.AnchorRoot).Root.KeyIDs[0][0] ^= 1
	expected(&in, in.Target.Metadata).Metadata.KeyIDs[0][0] ^= 1
	expected(&in, extra).Metadata.RevokedIDs[0][0] ^= 1
	expected(&in, in.Target.Desired).Desired.SlotIDs[0] = "other"
	in.Target.Links[0].Rotation.Payload[0] ^= 1
	s.Target.Links[0].Rotation.Payload[0] ^= 1
	s.Coverage[0].Status = InvalidProof
	if !reflect.DeepEqual(m.input, owned) || !reflect.DeepEqual(m.Summary(), saved) {
		t.Fatal("model or summary aliases")
	}
	anchor := m.input.Anchor.Raw
	raws := [][]byte{}
	for _, mat := range m.input.Materials.Materials {
		raws = append(raws, mat.Raw)
	}
	rootKeys := expected(&m.input, m.input.Target.AnchorRoot).Root.KeyIDs
	metaKeys := expected(&m.input, m.input.Target.Metadata).Metadata.KeyIDs
	revoked := expected(&m.input, extra).Metadata.RevokedIDs
	slots := expected(&m.input, m.input.Target.Desired).Desired.SlotIDs
	links := m.input.Target.Links
	caller := encode(t, in)
	m.Clear()
	m.Clear()
	var nilModel *CompositionModel
	nilModel.Clear()
	if !reflect.DeepEqual(m.Summary(), Summary{}) || !reflect.DeepEqual(nilModel.Summary(), Summary{}) || !reflect.DeepEqual((&CompositionModel{}).Summary(), Summary{}) || !errors.Is(Recheck(context.Background(), m, in), ErrClearedModel) || !errors.Is(Recheck(context.Background(), nilModel, in), ErrClearedModel) {
		t.Fatal("invalid owner")
	}
	if !bytes.Equal(anchor, make([]byte, len(anchor))) || rootKeys[0] != (contracts.Digest{}) || metaKeys[0] != (contracts.Digest{}) || revoked[0] != (contracts.Digest{}) || slots[0] != "" || links[0] != (RotationLink{}) {
		t.Fatal("reachable owned aliases not zero")
	}
	for _, raw := range raws {
		if !bytes.Equal(raw, make([]byte, len(raw))) {
			t.Fatal("raw not zero")
		}
	}
	if !bytes.Equal(caller, encode(t, in)) || !reflect.DeepEqual(saved.Target, owned.Target) {
		t.Fatal("caller/detached summary cleared")
	}
}
func TestCompositionRecheck(t *testing.T) {
	in := fixture(t, 2)
	old := evaluated(t, in)
	if e := Recheck(context.Background(), old, in); e != nil {
		t.Fatal(e)
	}
	for i, j := 0, len(in.Materials.Materials)-1; i < j; i, j = i+1, j-1 {
		in.Materials.Materials[i], in.Materials.Materials[j] = in.Materials.Materials[j], in.Materials.Materials[i]
		in.Materials.Expected[i], in.Materials.Expected[j] = in.Materials.Expected[j], in.Materials.Expected[i]
	}
	for i := range in.Materials.Expected {
		e := &in.Materials.Expected[i]
		if len(e.Metadata.RevokedIDs) == 0 {
			e.Metadata.RevokedIDs = []contracts.Digest{}
		}
		if len(e.Desired.SlotIDs) == 0 {
			e.Desired.SlotIDs = []string{}
		}
	}
	if e := Recheck(context.Background(), old, in); e != nil {
		t.Fatal("normalized identity", e)
	}
	in.Target.Links[0], in.Target.Links[1] = in.Target.Links[1], in.Target.Links[0]
	if e := Recheck(context.Background(), old, in); e != ErrBindingMismatch {
		t.Fatal("ordered links", e)
	}
	for _, mode := range []string{"selected same P different E", "compatible metadata declaration", "unselected replacement", "coherent scope"} {
		t.Run(mode, func(t *testing.T) {
			original := fixture(t, 1)
			extra := compatibleMetadata(t, &original, 8, false)
			old := evaluated(t, original)
			replacement := fixture(t, 1)
			compatibleMetadata(t, &replacement, 8, false)
			switch mode {
			case "selected same P different E":
				r := replacement.Target.Metadata
				replacement.Target.Metadata = replaceEnvelope(t, &replacement, r, extraShare(t, material(&replacement, r).Raw, "signatures"))
			case "compatible metadata declaration":
				r := compatibleMetadata(t, &replacement, 7, false)
				replacement.Target.Metadata = r
				expected(&replacement, replacement.Target.Desired).Desired.MetadataPayload = r.Payload
			case "unselected replacement":
				replaceEnvelope(t, &replacement, extra, extraShare(t, material(&replacement, extra).Raw, "signatures"))
			case "coherent scope":
				replacement.Materials.Scope.NodeID = "22345678-1234-4234-8234-123456789abc"
				replacement.Anchor.Scope = replacement.Materials.Scope
				p := object(t, payload(t, material(&replacement, replacement.Target.Desired).Raw))
				p["node_id"] = replacement.Materials.Scope.NodeID
				replacement.Target.Desired = replaceEnvelope(t, &replacement, replacement.Target.Desired, mint(t, "desired", p, private(2)))
				expected(&replacement, replacement.Target.Desired).Desired.NodeID = replacement.Materials.Scope.NodeID
			}
			evaluated(t, replacement)
			if e := Recheck(context.Background(), old, replacement); !errors.Is(e, ErrBindingMismatch) {
				t.Fatal("complete identity", e)
			}
		})
	}
	base := fixture(t, 1)
	old = evaluated(t, base)
	for i, e := range base.Materials.Expected {
		for _, path := range expectationPaths(reflect.ValueOf(e), nil) {
			in := fixture(t, 1)
			v := reflect.ValueOf(&in.Materials.Expected[i]).Elem()
			for _, idx := range path {
				v = v.Field(idx)
			}
			mutateExpectation(v)
			if e := Recheck(context.Background(), old, in); e == nil {
				t.Fatalf("expected field accepted: %d %v", i, path)
			}
		}
	}
}
func expectationPaths(v reflect.Value, path []int) (out [][]int) {
	if v.Kind() != reflect.Struct {
		return [][]int{append([]int(nil), path...)}
	}
	for i := 0; i < v.NumField(); i++ {
		out = append(out, expectationPaths(v.Field(i), append(append([]int(nil), path...), i))...)
	}
	return
}
func mutateExpectation(v reflect.Value) {
	switch v.Kind() {
	case reflect.Uint8, reflect.Uint64:
		v.SetUint(v.Uint() + 1)
	case reflect.Int64:
		v.SetInt(v.Int() + 1)
	case reflect.String:
		v.SetString(v.String() + "x")
	case reflect.Array:
		mutateExpectation(v.Index(0))
	case reflect.Slice:
		if v.Len() == 0 {
			v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		}
		mutateExpectation(v.Index(0))
	default:
		panic("unexpected expectation field")
	}
}

type countedContext struct {
	context.Context
	calls, at int
	failure   error
	panicAt   int
}

func (c *countedContext) Err() error {
	c.calls++
	if c.calls == c.panicAt {
		panic("composition-context-panic")
	}
	if c.at > 0 && c.calls >= c.at {
		return c.failure
	}
	return nil
}
func TestCompositionCancellation(t *testing.T) {
	in := fixture(t, 1)
	original := fixture(t, 1)
	counter := &countedContext{Context: context.Background()}
	m, e := Evaluate(counter, in)
	requireOutcome(t, m, e, nil)
	total := counter.calls
	t.Logf("Evaluate Context.Err checkpoints: %d", total)
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for i := 1; i <= total; i++ {
			ctx := &countedContext{Context: context.Background(), at: i, failure: failure}
			m, e := Evaluate(ctx, in)
			requireOutcome(t, m, e, failure)
			if !reflect.DeepEqual(in, original) {
				t.Fatal("caller changed at checkpoint", i)
			}
		}
	}
	old := evaluated(t, in)
	counter = &countedContext{Context: context.Background()}
	if e := Recheck(counter, old, in); e != nil {
		t.Fatal(e)
	}
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		ctx := &countedContext{Context: context.Background(), at: counter.calls, failure: failure}
		if e := Recheck(ctx, old, in); !errors.Is(e, failure) {
			t.Fatal("late recheck cancellation", e)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	oversized := Input{Anchor: AnchorInput{Raw: make([]byte, 262145)}}
	m, e = Evaluate(ctx, oversized)
	requireOutcome(t, m, e, context.Canceled)
}
func TestCompositionPanicCleanup(t *testing.T) {
	in := fixture(t, 1)
	before := fixture(t, 1)
	counter := &countedContext{Context: context.Background()}
	m, e := Evaluate(counter, in)
	requireOutcome(t, m, e, nil)
	for _, at := range []int{4, 16, counter.calls} {
		func() {
			defer func() {
				if got := recover(); got != "composition-context-panic" {
					t.Fatal("panic not propagated", fmt.Sprint(got))
				}
			}()
			ctx := &countedContext{Context: context.Background(), panicAt: at}
			Evaluate(ctx, in)
			t.Fatal("missing panic")
		}()
		if !reflect.DeepEqual(in, before) {
			t.Fatal("panic cleared caller")
		}
	}
	// Reachable alias clearing is asserted above and in owned_test. AST cleanup
	// registration supplies structural evidence for unreachable failure owners.
}
