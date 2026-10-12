package statematerials

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gowebpki/jcs"
	"os"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func rawFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("../stateformat/testdata/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func obj(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		t.Fatal("fixture JSON")
	}
	return m
}
func canon(t *testing.T, m any) []byte {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	c, e := jcs.Transform(b)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func digestText(t *testing.T, s string) contracts.Digest {
	t.Helper()
	var d contracts.Digest
	b, e := hex.DecodeString(s)
	if e != nil || len(b) != 32 {
		t.Fatal(s)
	}
	copy(d[:], b)
	return d
}
func envelopeRef(t *testing.T, k MaterialKind, b []byte) MaterialRef {
	w := obj(t, b)
	p, e := base64.RawURLEncoding.DecodeString(w["payload_b64url"].(string))
	if e != nil {
		t.Fatal(e)
	}
	return MaterialRef{k, sha256.Sum256(p), sha256.Sum256(b)}
}
func subsetFixture(t *testing.T) Input {
	key := digestText(t, "01d0fabd251fcbbe2b93b4b927b26ad2a1a99077152e45ded1e678afa45dbec5")
	r1 := rawFixture(t, "root-initial.payload")
	r2 := canon(t, obj(t, rawFixture(t, "root-rotation.payload"))["new_root_set"])
	rotation := rawFixture(t, "root-rotation.envelope")
	metadata := rawFixture(t, "metadata.envelope")
	desired := rawFixture(t, "desired.envelope")
	refs := []MaterialRef{{RootSet, sha256.Sum256(r1), contracts.Digest{}}, {RootSet, sha256.Sum256(r2), contracts.Digest{}}, envelopeRef(t, RootRotation, rotation), envelopeRef(t, Metadata, metadata), envelopeRef(t, Desired, desired)}
	in := Input{Scope: Scope{"12345678-1234-4234-8234-123456789abc", 19}}
	raws := [][]byte{r1, r2, rotation, metadata, desired}
	for i := range refs {
		in.Materials = append(in.Materials, Material{refs[i], raws[i]})
	}
	week := Window{1791504000, 1792108800}
	day := Window{1791504000, 1791590400}
	in.Expected = []Expectation{
		{Ref: refs[0], Epoch: 19, Sequence: 3, Root: RootExpectation{1, 1, []contracts.Digest{key}}},
		{Ref: refs[1], Epoch: 19, Sequence: 4, Root: RootExpectation{2, 1, []contracts.Digest{key}}},
		{Ref: refs[2], Epoch: 19, Sequence: 4, Rotation: RotationExpectation{1, 2, refs[0].Payload, refs[1].Payload, week}},
		{Ref: refs[3], Epoch: 19, Sequence: 5, Metadata: MetadataExpectation{6, 2, refs[1].Payload, []contracts.Digest{key}, nil, week}},
		{Ref: refs[4], Epoch: 19, Sequence: 9007199254740993, Desired: DesiredExpectation{in.Scope.NodeID, 7, 8, 9, digestText(t, "abababababababababababababababababababababababababababababababab"), refs[1].Payload, refs[3].Payload, key, day, nil}},
	}
	return in
}
func expectError(t *testing.T, in Input, want error) {
	t.Helper()
	m, e := Evaluate(testContext(), in)
	if m != nil || e != want {
		if m != nil {
			m.Clear()
		}
		t.Fatalf("want %v nil model, got %v %v", want, m, e)
	}
}
func rewrapMaterial(t *testing.T, in *Input, index int, mut func(map[string]any)) {
	t.Helper()
	w := obj(t, in.Materials[index].Raw)
	p, _ := base64.RawURLEncoding.DecodeString(w["payload_b64url"].(string))
	o := obj(t, p)
	mut(o)
	p = canon(t, o)
	w["payload_b64url"] = base64.RawURLEncoding.EncodeToString(p)
	w["payload_sha256"] = fmt.Sprintf("%x", sha256.Sum256(p))
	b := canon(t, w)
	r := envelopeRef(t, in.Materials[index].Ref.Kind, b)
	in.Materials[index] = Material{r, b}
	in.Expected[index].Ref = r
}
