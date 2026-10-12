package nodecerttime

import (
	"context"
	"strings"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
	"time"
)

func budgetInput() Input {
	var in Input
	in.Certificate.Leaf.DER = []byte{1}
	in.Certificate.Leaf.ExpectedSPKI = []byte{2}
	in.Certificate.CA.DER = []byte{3}
	in.Certificate.CA.DeclaredSPKI = []byte{4}
	in.Time.Policy.Binding.Proposed.Anchor.Raw = []byte{5}
	return in
}
func budgetResize[T any](v []T, n int) []T { return make([]T, n) }

func TestCertTimeFixedErrors(t *testing.T) {
	errors := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidEvidence, ErrBindingMismatch, ErrNotYetValid, ErrExpired}
	texts := []string{"nodecerttime: invalid input", "nodecerttime: limit exceeded", "nodecerttime: invalid evidence", "nodecerttime: binding mismatch", "nodecerttime: certificate not yet valid", "nodecerttime: certificate expired"}
	for i, e := range errors {
		if e.Error() != texts[i] {
			t.Fatal(i, e)
		}
		for j := 0; j < i; j++ {
			if e == errors[j] {
				t.Fatal("aliased error")
			}
		}
	}
}
func TestCertTimePreflightPriority(t *testing.T) {
	if preflight(nil, Input{}) != ErrInvalidInput {
		t.Fatal("nil context")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if preflight(ctx, Input{}) != context.Canceled {
		t.Fatal("cancel priority")
	}
	expired, done := context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer done()
	if preflight(expired, Input{}) != context.DeadlineExceeded {
		t.Fatal("deadline priority")
	}
	empties := []func(*Input){func(x *Input) { x.Certificate.Leaf.DER = nil }, func(x *Input) { x.Certificate.Leaf.ExpectedSPKI = nil }, func(x *Input) { x.Certificate.CA.DER = nil }, func(x *Input) { x.Certificate.CA.DeclaredSPKI = nil }, func(x *Input) { x.Time.Policy.Binding.Proposed.Anchor.Raw = nil }}
	for i, set := range empties {
		in := budgetInput()
		set(&in)
		if preflight(context.Background(), in) != ErrInvalidInput {
			t.Fatal("empty", i)
		}
		in.Time.Policy.Policy.History = budgetResize(in.Time.Policy.Policy.History, 4097)
		if preflight(context.Background(), in) != ErrLimitExceeded {
			t.Fatal("late cap before empty", i)
		}
	}
	in := budgetInput()
	in.Time.Policy.Binding.Previous.Materials = []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: 255}, Raw: []byte{1}}}
	if e := preflight(context.Background(), in); e != nil {
		t.Fatal("semantic repair", e)
	}
}

// Detects missing checks in either complete table, Main, Old or History position.
func TestCertTimeBudgetEdges(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		set   func(*Input, int)
	}{
		{"leaf-der", 65536, func(x *Input, n int) { x.Certificate.Leaf.DER = make([]byte, n) }},
		{"leaf-spki", 512, func(x *Input, n int) { x.Certificate.Leaf.ExpectedSPKI = make([]byte, n) }},
		{"ca-der", 65536, func(x *Input, n int) { x.Certificate.CA.DER = make([]byte, n) }},
		{"ca-spki", 8192, func(x *Input, n int) { x.Certificate.CA.DeclaredSPKI = make([]byte, n) }},
		{"domain", 253, func(x *Input, n int) { x.Certificate.Leaf.TrustDomain = strings.Repeat("x", n) }},
		{"anchor", 262144, func(x *Input, n int) { x.Time.Policy.Binding.Proposed.Anchor.Raw = make([]byte, n) }},
		{"transitions", 32, func(x *Input, n int) {
			x.Time.Policy.Binding.Main.RootTransition = budgetResize(x.Time.Policy.Binding.Main.RootTransition, n)
		}},
		{"links", 32, func(x *Input, n int) {
			x.Time.Policy.Binding.Proposed.Target.Links = budgetResize(x.Time.Policy.Binding.Proposed.Target.Links, n)
		}},
		{"batch-rotations", 32, func(x *Input, n int) { x.Time.Policy.Policy.Batch.Rotations = make([][]byte, n) }},
		{"batch-anchor", 262144, func(x *Input, n int) { x.Time.Policy.Policy.Batch.AnchorEnvelope = make([]byte, n) }},
		{"batch-active", 262144, func(x *Input, n int) { x.Time.Policy.Policy.Batch.ActiveRootEnvelope = make([]byte, n) }},
		{"batch-metadata", 262144, func(x *Input, n int) { x.Time.Policy.Policy.Batch.MetadataEnvelope = make([]byte, n) }},
		{"batch-desired", 262144, func(x *Input, n int) { x.Time.Policy.Policy.Batch.DesiredEnvelope = make([]byte, n) }},
		{"batch-rotation-raw", 262144, func(x *Input, n int) { x.Time.Policy.Policy.Batch.Rotations = [][]byte{nil, make([]byte, n)} }},
		{"attestation", 262144, func(x *Input, n int) { x.Time.Time.AttestationEnvelope = make([]byte, n) }},
		{"time-metadata", 262144, func(x *Input, n int) { x.Time.Time.MetadataEnvelope = make([]byte, n) }},
	}
	for table := 0; table < 2; table++ {
		get := func(x *Input) *statematerials.Input {
			if table == 0 {
				return &x.Time.Policy.Binding.Previous
			}
			return &x.Time.Policy.Binding.Proposed.Materials
		}
		label := []string{"previous-", "proposed-"}[table]
		cases = append(cases, struct {
			name  string
			limit int
			set   func(*Input, int)
		}{label + "materials", 128, func(x *Input, n int) { get(x).Materials = make([]statematerials.Material, n) }}, struct {
			name  string
			limit int
			set   func(*Input, int)
		}{label + "expected", 128, func(x *Input, n int) { get(x).Expected = make([]statematerials.Expectation, n) }})
		for _, expected := range []bool{false, true} {
			cases = append(cases, struct {
				name  string
				limit int
				set   func(*Input, int)
			}{label + "rotations", 32, func(x *Input, n int) {
				a := get(x)
				if expected {
					a.Expected = make([]statematerials.Expectation, n)
					for i := range a.Expected {
						a.Expected[i].Ref.Kind = statematerials.RootRotation
					}
				} else {
					a.Materials = make([]statematerials.Material, n)
					for i := range a.Materials {
						a.Materials[i].Ref.Kind = statematerials.RootRotation
					}
				}
			}})
		}
		for _, kind := range []statematerials.MaterialKind{statematerials.RootSet, statematerials.RootRotation, statematerials.Metadata, statematerials.Desired, 255} {
			limit := 262144
			if kind == statematerials.RootSet {
				limit = 65536
			}
			cases = append(cases, struct {
				name  string
				limit int
				set   func(*Input, int)
			}{label + "raw", limit, func(x *Input, n int) {
				get(x).Materials = []statematerials.Material{{}, {Ref: statematerials.MaterialRef{Kind: kind}, Raw: make([]byte, n)}}
			}})
		}
		for pos := 0; pos < 2; pos++ {
			for _, c := range []struct {
				name  string
				limit int
				set   func(*statematerials.Expectation, int)
			}{
				{"root-keys", 5, func(e *statematerials.Expectation, n int) { e.Root.KeyIDs = budgetResize(e.Root.KeyIDs, n) }},
				{"metadata-keys", 64, func(e *statematerials.Expectation, n int) { e.Metadata.KeyIDs = budgetResize(e.Metadata.KeyIDs, n) }},
				{"revoked", 4096, func(e *statematerials.Expectation, n int) {
					e.Metadata.RevokedIDs = budgetResize(e.Metadata.RevokedIDs, n)
				}},
				{"slots", 8, func(e *statematerials.Expectation, n int) { e.Desired.SlotIDs = make([]string, n) }},
			} {
				cases = append(cases, struct {
					name  string
					limit int
					set   func(*Input, int)
				}{label + c.name, c.limit, func(x *Input, n int) {
					get(x).Expected = make([]statematerials.Expectation, 2)
					c.set(&get(x).Expected[pos], n)
				}})
			}
		}
	}
	for which := 0; which < 2; which++ {
		for _, c := range mainBudgetCases() {
			cases = append(cases, struct {
				name  string
				limit int
				set   func(*Input, int)
			}{c.name, c.limit, func(x *Input, n int) { c.set(x, which, n) }})
		}
	}
	cases = append(cases, oldBudgetCases()...)
	for i, c := range cases {
		t.Run(c.name+"-"+string(rune('A'+i)), func(t *testing.T) {
			for _, n := range []int{c.limit, c.limit + 1} {
				x := budgetInput()
				c.set(&x, n)
				want := error(nil)
				if n > c.limit {
					want = ErrLimitExceeded
				}
				if got := preflight(context.Background(), x); got != want {
					t.Fatalf("n=%d got=%v want=%v", n, got, want)
				}
			}
		})
	}
}

func mainBudgetCases() []struct {
	name  string
	limit int
	set   func(*Input, int, int)
} {
	return []struct {
		name  string
		limit int
		set   func(*Input, int, int)
	}{
		{"main-root", 5, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Root.KeyIDs = budgetResize(m.Root.KeyIDs, n)
		}},
		{"main-meta", 64, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Metadata.KeyIDs = budgetResize(m.Metadata.KeyIDs, n)
		}},
		{"main-revoked", 4096, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Metadata.RevokedIDs = budgetResize(m.Metadata.RevokedIDs, n)
		}},
		{"main-bootstrap", 128, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.ServerTrust.Bootstrap.Deauthorized = budgetResize(m.ServerTrust.Bootstrap.Deauthorized, n)
		}},
		{"main-agent", 128, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.ServerTrust.Agent.Deauthorized = budgetResize(m.ServerTrust.Agent.Deauthorized, n)
		}},
		{"main-slots", 8, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Desired.SlotIDs = make([]string, n)
		}},
		{"main-lkg", 8, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.LKG.Entries = budgetResize(m.LKG.Entries, n)
		}},
		{"main-local", 64, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Recovery.Local = budgetResize(m.Recovery.Local, n)
		}},
		{"main-supervisor", 16, func(x *Input, w, n int) {
			m := &x.Time.Policy.Binding.Main.Previous
			if w == 1 {
				m = &x.Time.Policy.Binding.Main.Proposed
			}
			m.Recovery.Supervisor = budgetResize(m.Recovery.Supervisor, n)
		}},
	}
}

func oldBudgetCases() []struct {
	name  string
	limit int
	set   func(*Input, int)
} {
	cases := []struct {
		name  string
		limit int
		set   func(*Input, int)
	}{
		{"old-roots", 33, func(x *Input, n int) {
			x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
			o := x.Time.Policy.Policy.Old
			o.Roots = budgetResize(o.Roots, n)
		}},
		{"old-keys", 4096, func(x *Input, n int) {
			x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
			o := x.Time.Policy.Policy.Old
			o.Keys = budgetResize(o.Keys, n)
		}},
		{"old-revoked", 4096, func(x *Input, n int) {
			x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
			o := x.Time.Policy.Policy.Old
			o.RevokedIDs = budgetResize(o.RevokedIDs, n)
		}},
		{"history", 4096, func(x *Input, n int) { x.Time.Policy.Policy.History = budgetResize(x.Time.Policy.Policy.History, n) }},
	}
	for pos := 0; pos < 2; pos++ {
		cases = append(cases,
			struct {
				name  string
				limit int
				set   func(*Input, int)
			}{"old-root-publickeys", 5, func(x *Input, n int) {
				x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
				o := x.Time.Policy.Policy.Old
				o.Roots = budgetResize(o.Roots, 2)
				o.Roots[pos].Root.Keys = budgetResize(o.Roots[pos].Root.Keys, n)
			}},
			struct {
				name  string
				limit int
				set   func(*Input, int)
			}{"old-key-status", 8, func(x *Input, n int) {
				x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
				o := x.Time.Policy.Policy.Old
				o.Keys = budgetResize(o.Keys, 2)
				o.Keys[pos].Material.Status = strings.Repeat("x", n)
			}},
			struct {
				name  string
				limit int
				set   func(*Input, int)
			}{"old-key-reason", 32, func(x *Input, n int) {
				x.Time.Policy.Policy.Old = &statepolicy.StateModel{}
				o := x.Time.Policy.Policy.Old
				o.Keys = budgetResize(o.Keys, 2)
				o.Keys[pos].Material.RevocationReason = strings.Repeat("x", n)
			}},
			struct {
				name  string
				limit int
				set   func(*Input, int)
			}{"history-kind", 64, func(x *Input, n int) {
				x.Time.Policy.Policy.History = budgetResize(x.Time.Policy.Policy.History, 2)
				x.Time.Policy.Policy.History[pos].IncidentKind = strings.Repeat("x", n)
			}},
		)
	}
	return cases
}

func TestCertTimeBudgetTotals(t *testing.T) {
	for _, limit := range []int{139776, 16777216, 33816576, 9437184, 43253760, 524288, 43778048, 43917824, int(^uint(0) >> 1)} {
		if total, ok := addRaw(limit, 0, limit); !ok || total != limit {
			t.Fatal("equal", limit)
		}
		if total, ok := addRaw(limit-1, 1, limit); !ok || total != limit {
			t.Fatal("last byte", limit)
		}
		if _, ok := addRaw(limit-1, 2, limit); ok {
			t.Fatal("overflow", limit)
		}
	}
	for _, v := range [][3]int{{-1, 1, 10}, {0, -1, 10}, {0, 1, -1}, {11, 0, 10}} {
		if _, ok := addRaw(v[0], v[1], v[2]); ok {
			t.Fatal("invalid accumulator")
		}
	}
	x := budgetInput()
	raw := make([]byte, 262144)
	x.Certificate.Leaf.DER = raw[:65536]
	x.Certificate.Leaf.ExpectedSPKI = raw[:512]
	x.Certificate.CA.DER = raw[:65536]
	x.Certificate.CA.DeclaredSPKI = raw[:8192]
	for _, table := range []*statematerials.Input{&x.Time.Policy.Binding.Previous, &x.Time.Policy.Binding.Proposed.Materials} {
		table.Materials = make([]statematerials.Material, 64)
		for i := range table.Materials {
			table.Materials[i].Raw = raw
		}
	}
	x.Time.Policy.Binding.Proposed.Anchor.Raw = raw
	b := &x.Time.Policy.Policy.Batch
	b.AnchorEnvelope = raw
	b.ActiveRootEnvelope = raw
	b.MetadataEnvelope = raw
	b.DesiredEnvelope = raw
	b.Rotations = make([][]byte, 32)
	for i := range b.Rotations {
		b.Rotations[i] = raw
	}
	x.Time.Time.MetadataEnvelope = raw
	x.Time.Time.AttestationEnvelope = raw
	if n, e := certificateBudget(x.Certificate); e != nil || n != 139776 {
		t.Fatal("certificate total", n, e)
	}
	if n, e := bindingBudget(x.Time.Policy.Binding); e != nil || n != 33816576 {
		t.Fatal("binding appearances", n, e)
	}
	if n, e := timeBudget(x.Time); e != nil || n != 43778048 {
		t.Fatal("time appearances", n, e)
	}
	if e := preflight(context.Background(), x); e != nil {
		t.Fatal("complete cap", e)
	}
	x.Time.Policy.Binding.Previous.Materials = append(x.Time.Policy.Binding.Previous.Materials, statematerials.Material{Raw: []byte{1}})
	if e := preflight(context.Background(), x); e != ErrLimitExceeded {
		t.Fatal("table aggregate", e)
	}
}
