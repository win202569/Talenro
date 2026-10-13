package nodeleafbinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"testing"
	"time"
)

func leafBindingResult(t testing.TB, in Input, want error) *Model {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if e != want || (e != nil && m != nil) || (e == nil && m == nil) {
		if m != nil {
			m.Clear()
		}
		t.Fatal("result", e, want)
	}
	return m
}
func TestLeafBindingProjection(t *testing.T) {
	in := leafBindingFixture(t, "base")
	evidence := assertLeafBindingLower(t, in)
	// Frozen by independent ignored reference-check.go, not recomputed by tests.
	caSPKI := [32]byte{0x5c, 0xd2, 0x52, 0xfb, 0xc, 0xe8, 0x93, 0x24, 0x36, 0xfa, 0xf8, 0xcc, 0xd1, 0x4, 0x9, 0x81, 0xb8, 0x9e, 0xe4, 0xad, 0x6b, 0x9f, 0xe9, 0xe2, 0xa2, 0xb7, 0xe7, 0x1a, 0xac, 0xb2, 0x7c, 0xd3}
	rawPoint := [32]byte{0xa9, 0xf3, 0x0, 0xeb, 0x59, 0x60, 0xe8, 0x91, 0x33, 0xaf, 0x73, 0x62, 0x1, 0x1a, 0x1e, 0x26, 0xf0, 0xe2, 0xea, 0x2e, 0x36, 0xdc, 0x40, 0x2a, 0x4, 0xaf, 0x6c, 0x19, 0x2b, 0x89, 0x1a, 0x8c}
	for _, c := range []struct {
		name      string
		mutate    func(*statecontract.MainStateContentModel)
		reachable bool
	}{
		{"node", func(s *statecontract.MainStateContentModel) { s.Scope.NodeID = "22345678-1234-4234-8234-123456789abc" }, false},
		{"header-domain", func(s *statecontract.MainStateContentModel) { s.Identity.Header.Domain = statecontract.DomainMetadata }, false},
		{"header-coverage", func(s *statecontract.MainStateContentModel) {
			s.Identity.Header.Coverage = statecontract.NeverEstablished
		}, false},
		{"leaf-kind", func(s *statecontract.MainStateContentModel) { s.Identity.Leaf.Kind = statecontract.IdentityHistory }, false},
		{"leaf-payload", func(s *statecontract.MainStateContentModel) { s.Identity.Leaf.Payload[0] ^= 1 }, true},
		{"leaf-envelope", func(s *statecontract.MainStateContentModel) { s.Identity.Leaf.Envelope[0] = 1 }, false},
		{"public-key", func(s *statecontract.MainStateContentModel) { s.Identity.PublicKey[0] ^= 1 }, true},
		{"ca-spki", func(s *statecontract.MainStateContentModel) { s.Identity.PublicKey = caSPKI }, true},
		{"ca-der", func(s *statecontract.MainStateContentModel) {
			s.Identity.PublicKey = evidence.Certificate.CA.CADERDigest
		}, true},
		{"raw-ec-point", func(s *statecontract.MainStateContentModel) { s.Identity.PublicKey = rawPoint }, true},
		{"issued", func(s *statecontract.MainStateContentModel) { s.Identity.Window.Issued++ }, true},
		{"until", func(s *statecontract.MainStateContentModel) { s.Identity.Window.Until-- }, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			trial := leafBindingFixture(t, "base")
			c.mutate(&trial.Time.Policy.Binding.Main.Proposed)
			projection, e := bindProjection(context.Background(), trial, evidence)
			if e != ErrBindingMismatch || projection != (Projection{}) {
				t.Fatal("necessary projection comparison", e)
			}
			c.mutate(&trial.Time.Policy.Binding.Main.Previous)
			if c.reachable {
				assertLeafBindingLower(t, trial)
				leafBindingResult(t, trial, ErrBindingMismatch)
			} else {
				lower, e := nodeLowerError(trial)
				if lower {
					t.Fatal("expected lower precondition failure")
				}
				if e == nil {
					t.Fatal("missing lower error")
				}
			}
		})
	}
	m := leafBindingResult(t, in, nil)
	defer m.Clear()
	s := m.Summary()
	if s.Proof != SelectedLeafProjectionConditionalTimeBound || s.Proposed != leafBindingExpected(t, "base").Projection || !reflect.DeepEqual(s.Evidence, evidence) {
		t.Fatal("exact projection/full lower")
	}
}
func TestLeafBindingSameKeyDifferentDER(t *testing.T) {
	a, b := leafBindingFixture(t, "base"), leafBindingFixture(t, "leaf-first")
	sa, sb := assertLeafBindingLower(t, a), assertLeafBindingLower(t, b)
	if sa.Certificate.Leaf.Leaf.NodeID != sb.Certificate.Leaf.Leaf.NodeID {
		t.Fatal("same-node alternative signed leaf control required")
	}
	if sa.Certificate.Leaf.Leaf.SPKIDERDigest != sb.Certificate.Leaf.Leaf.SPKIDERDigest || sa.Certificate.Leaf.Leaf.LeafDERDigest == sb.Certificate.Leaf.Leaf.LeafDERDigest {
		t.Fatal("same SPKI different signed DER control")
	}
	b.Time.Policy.Binding.Main.Previous.Identity.Leaf = a.Time.Policy.Binding.Main.Previous.Identity.Leaf
	b.Time.Policy.Binding.Main.Proposed.Identity.Leaf = a.Time.Policy.Binding.Main.Proposed.Identity.Leaf
	assertLeafBindingLower(t, b)
	leafBindingResult(t, b, ErrBindingMismatch)
}
func TestLeafBindingWindowEdges(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	for _, delta := range []int64{-1, 0, 1} {
		for _, endpoint := range []int{0, 1} {
			w := statecontract.TimeWindow{Issued: start.Unix(), Until: end.Unix()}
			if endpoint == 0 {
				w.Issued += delta
			} else {
				w.Until += delta
			}
			if windowMatches(w, start, end) != (delta == 0) {
				t.Fatal("exact seconds", endpoint, delta)
			}
		}
	}
	for _, w := range []statecontract.TimeWindow{{Issued: -62135596801, Until: end.Unix()}, {Issued: start.Unix(), Until: 253402300800}, {Issued: end.Unix(), Until: start.Unix()}, {Issued: start.Unix(), Until: start.Unix()}} {
		if windowMatches(w, start, end) {
			t.Fatal("invalid range/order")
		}
	}
	w := statecontract.TimeWindow{Issued: start.Unix(), Until: end.Unix()}
	if windowMatches(w, start.Add(time.Nanosecond), end) || windowMatches(w, start, end.Add(time.Nanosecond)) {
		t.Fatal("nanoseconds truncated")
	}
}
func TestLeafBindingYearOneWindow(t *testing.T) {
	start := time.Time{}
	end := time.Date(1, 1, 2, 0, 0, 0, 0, time.UTC)
	if !windowMatches(statecontract.TimeWindow{Issued: -62135596800, Until: end.Unix()}, start, end) {
		t.Fatal("valid year-one endpoint is not a missing observation")
	}
	// Frozen year-one leaf is expired at its real 2020 observation; this is not
	// a claim of a year-one provider or successful runtime authorization.
	leafBindingResult(t, leafBindingFixture(t, "leaf-year-1"), certTimeExpiredError())
	m := leafBindingResult(t, leafBindingFixture(t, "ca-year-1"), nil)
	m.Clear()
}
