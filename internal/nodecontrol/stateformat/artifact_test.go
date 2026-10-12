package stateformat

import (
	"bytes"
	"crypto/sha256"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func TestBuildDesiredEnvelope(t *testing.T) {
	p := fixture(t, "desired.payload")
	d := sha256.Sum256(p)
	key := contracts.Digest{1}
	sig := [64]byte{2}
	gold := object(t, fixture(t, "desired.envelope"))
	var e error
	key, e = parseDigest(gold["key_id"].(string))
	if e != nil {
		t.Fatal(e)
	}
	a, e := BuildDesiredEnvelope(p, d, key, sig)
	if e != nil || a == nil {
		t.Fatal(e)
	}
	defer a.Clear()
	if !bytes.Equal(a.Envelope(), fixture(t, "desired.envelope")) {
		t.Fatal("not unique envelope")
	}
	for _, tc := range []struct {
		p    []byte
		d, k contracts.Digest
		s    [64]byte
	}{{p, contracts.Digest{1}, key, sig}, {p, d, contracts.Digest{}, sig}, {p, d, key, [64]byte{}}, {append(bytes.Clone(p), ' '), d, key, sig}} {
		if a, e := BuildDesiredEnvelope(tc.p, tc.d, tc.k, tc.s); a != nil || e != ErrInvalidFormat {
			t.Fatal("invalid build", e)
		}
	}
}

func TestArtifactOwnership(t *testing.T) {
	input := fixture(t, "desired.envelope")
	original := bytes.Clone(input)
	a, e := ParseEnvelope(Desired, input)
	if e != nil {
		t.Fatal(e)
	}
	input[0] ^= 1
	if !bytes.Equal(a.Envelope(), original) {
		t.Fatal("input alias")
	}
	for _, get := range []func() []byte{a.Payload, a.Envelope, func() []byte { v, _ := a.Transcript(); return v }} {
		x := get()
		want := bytes.Clone(x)
		x[0] ^= 1
		if !bytes.Equal(get(), want) {
			t.Fatal("accessor alias")
		}
	}
	b := a.Bindings()
	b.Signature[0] ^= 1
	if a.Bindings().Signature[0] != 2 {
		t.Fatal("bindings alias")
	}
	p, eBytes := a.payload, a.envelope
	a.Clear()
	for _, buf := range [][]byte{p, eBytes} {
		if !bytes.Equal(buf, make([]byte, len(buf))) {
			t.Fatal("owned bytes not cleared")
		}
	}
	a.Clear()
	for _, empty := range []*ParsedArtifact{nil, {}, a} {
		if empty.Payload() != nil || empty.Envelope() != nil || empty.Bindings() != (Bindings{}) || empty.PayloadDigest() != (contracts.Digest{}) || empty.EnvelopeDigest() != (contracts.Digest{}) {
			t.Fatal("empty artifact")
		}
		if v, e := empty.Transcript(); v != nil || e != ErrInvalidFormat {
			t.Fatal("empty transcript")
		}
		empty.Clear()
	}
}
