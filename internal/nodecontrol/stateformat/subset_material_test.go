package stateformat

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"
)

// Catch root/rotation identity conflation and accidental initial-root widening.
func TestSubsetRootFacts(t *testing.T) {
	for _, version := range []string{"1", "2", "9223372036854775807"} {
		w := object(t, fixture(t, "root-initial.payload"))
		w["root_version"] = version
		raw := encoded(t, w)
		r, e := ParseRootSet(raw)
		if e != nil {
			t.Fatal(e)
		}
		if r.PayloadDigest() != sha256.Sum256(raw) || !bytes.Equal(r.Payload(), raw) || r.Bindings().Epoch != 19 {
			t.Fatal("root facts")
		}
		r.Clear()
	}
	a, e := ParseEnvelope(RootRotation, fixture(t, "root-rotation.envelope"))
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	r, e := a.EmbeddedRootSet()
	if e != nil {
		t.Fatal(e)
	}
	defer r.Clear()
	raw := encoded(t, object(t, fixture(t, "root-rotation.payload"))["new_root_set"])
	if !bytes.Equal(raw, r.Payload()) || r.PayloadDigest() != sha256.Sum256(raw) || r.PayloadDigest() == a.PayloadDigest() || r.Bindings().Version != 2 || r.Bindings().Sequence != 4 {
		t.Fatal("embedded identity")
	}
	w := object(t, fixture(t, "root-initial.payload"))
	w["root_version"] = "2"
	outer := object(t, rewrap(t, "root-initial", encoded(t, w)))
	outer["root_version"] = "2"
	if a, e := ParseEnvelope(InitialRoot, encoded(t, outer)); a != nil || e != ErrInvalidFormat {
		t.Fatal("initial widened")
	}
	for _, raw := range [][]byte{nil, []byte("{}"), fixture(t, "root-rotation.payload")} {
		if r, e := ParseRootSet(raw); r != nil || e != ErrInvalidFormat {
			t.Fatal("invalid root accepted")
		}
	}
}

func TestSubsetFactOwnership(t *testing.T) {
	raw := fixture(t, "root-initial.payload")
	expected := bytes.Clone(raw)
	r, e := ParseRootSet(raw)
	if e != nil {
		t.Fatal(e)
	}
	raw[0] ^= 1
	out := r.Payload()
	out[0] ^= 1
	m, e := r.RootMaterial()
	if e != nil {
		t.Fatal(e)
	}
	m.Keys[0].PublicKey[0] ^= 1
	again, _ := r.RootMaterial()
	if again.Keys[0].PublicKey != ([32]byte{1}) || !bytes.Equal(r.Payload(), expected) {
		t.Fatal("root alias")
	}
	held := r.Payload()
	r.Clear()
	r.Clear()
	if !bytes.Equal(held, expected) {
		t.Fatal("clear changed detached bytes")
	}
	for _, empty := range []*ParsedRootSet{nil, {}, r} {
		empty.Clear()
		if _, e := empty.RootMaterial(); e != ErrInvalidFormat {
			t.Fatal("cleared root")
		}
	}
	a, _ := ParseEnvelope(RootRotation, fixture(t, "root-rotation.envelope"))
	x, _ := a.EmbeddedRootSet()
	x.Clear()
	y, _ := a.EmbeddedRootSet()
	a.Clear()
	if y.Bindings().Version != 2 {
		t.Fatal("embedded owner alias")
	}
	y.Clear()
	m.Clear()
	again.Clear()
	for _, empty := range []*ParsedArtifact{nil, {}, a} {
		if _, e := empty.EmbeddedRootSet(); e != ErrInvalidFormat {
			t.Fatal("wrong embedded kind")
		}
	}
}

func TestSubsetSlots(t *testing.T) {
	for _, n := range []int{0, 8, 9} {
		w := desiredObject(t)
		ps := make([]any, n)
		want := make([]string, n)
		for i := range ps {
			p := processFixture(t)
			want[i] = fmt.Sprintf("slot-%02d", i)
			p["slot_id"] = want[i]
			ps[i] = p
		}
		w["processes"] = ps
		a, e := ParseEnvelope(Desired, rewrap(t, "desired", encoded(t, w)))
		if n == 9 {
			if a != nil || e != ErrInvalidFormat {
				t.Fatal("slots overlimit")
			}
			continue
		}
		if e != nil {
			t.Fatal(e)
		}
		got, e := a.DesiredSlotIDs()
		if e != nil || len(got) != n {
			t.Fatal("slots", e)
		}
		if n > 0 {
			if !reflect.DeepEqual(got, want) {
				t.Fatal(got)
			}
			got[0] = "changed"
			again, _ := a.DesiredSlotIDs()
			if again[0] != want[0] {
				t.Fatal("slots alias")
			}
		}
		a.Clear()
		if _, e := a.DesiredSlotIDs(); e != ErrInvalidFormat {
			t.Fatal("cleared slots")
		}
	}
	for _, slots := range [][]string{{"a", "a"}, {"z", "a"}} {
		w := desiredObject(t)
		ps := []any{processFixture(t), processFixture(t)}
		for i := range ps {
			ps[i].(map[string]any)["slot_id"] = slots[i]
		}
		w["processes"] = ps
		if a, e := ParseEnvelope(Desired, rewrap(t, "desired", encoded(t, w))); a != nil || e != ErrInvalidFormat {
			t.Fatal("noncanonical slots")
		}
	}
	for _, kind := range []Kind{InitialRoot, RootRotation, Metadata} {
		name := map[Kind]string{InitialRoot: "root-initial", RootRotation: "root-rotation", Metadata: "metadata"}[kind]
		a, _ := ParseEnvelope(kind, fixture(t, name+".envelope"))
		if _, e := a.DesiredSlotIDs(); e != ErrInvalidFormat {
			t.Fatal("wrong slots kind")
		}
		a.Clear()
	}
}
