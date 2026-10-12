package statepolicy

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/stateverify"
	"testing"
)

func wantCode(t *testing.T, err error, code Code) {
	t.Helper()
	if code == 0 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var p *PolicyError
	if !errors.As(err, &p) || p.Code != code {
		t.Fatalf("want code %v, got %v", code, err)
	}
}
func TestLoadFactsBindings(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		b := signedBatch(t, n)
		f, e := loadFacts(context.Background(), b)
		wantCode(t, e, 0)
		if f == nil {
			t.Fatal("missing facts")
		}
		defer f.Clear()
		if len(f.rotations) != n || f.metadata.bindings.MetadataVersion != 6 || f.desired.bindings.Generation != 7 || f.desired.bindings.Sequence != 9 {
			t.Fatal("wrong facts")
		}
		if f.desired.id.EnvelopeDigest != sha256.Sum256(b.DesiredEnvelope) {
			t.Fatal("unbound envelope")
		}
		if f.metadata.metadata.Keys[0].Status != "active" {
			t.Fatal("missing key facts")
		}
	}
	for _, part := range []string{"anchor", "rotation", "metadata", "desired", "active"} {
		t.Run(part, func(t *testing.T) {
			b := signedBatch(t, 1)
			switch part {
			case "anchor":
				b.AnchorEnvelope = []byte("{}")
			case "rotation":
				b.Rotations[0] = []byte("{}")
			case "metadata":
				b.MetadataEnvelope = []byte("{}")
			case "desired":
				b.DesiredEnvelope = []byte("{}")
			case "active":
				b.ActiveRootEnvelope = []byte("{}")
			}
			f, e := loadFacts(context.Background(), b)
			if e == nil || f != nil {
				t.Fatal("substitution accepted or partial output")
			}
		})
	}
	b := signedBatch(t, 1)
	p := jsonObject(t, b.DesiredEnvelope)
	sig, decodeErr := base64.RawURLEncoding.DecodeString(p["signature"].(string))
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	sig[0] ^= 1
	p["signature"] = base64.RawURLEncoding.EncodeToString(sig)
	b.DesiredEnvelope = jsonBytes(t, p)
	f, e := loadFacts(context.Background(), b)
	wantCode(t, e, Rejected)
	if f != nil {
		t.Fatal("partial facts")
	}
	b = signedBatch(t, 33)
	f, e = loadFacts(context.Background(), b)
	wantCode(t, e, AssuranceMissing)
	if f != nil {
		t.Fatal("partial range")
	}
}
func TestLoadFactsCancellation(t *testing.T) {
	f, e := loadFacts(nil, stateverify.Batch{})
	wantCode(t, e, InvalidInput)
	if f != nil {
		t.Fatal("nil ctx")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f, e = loadFacts(ctx, stateverify.Batch{})
	wantCode(t, e, Canceled)
	if f != nil {
		t.Fatal("cancel partial")
	}
}
func TestModelOwnership(t *testing.T) {
	b := signedBatch(t, 1)
	before := append([]byte(nil), b.AnchorEnvelope...)
	f, e := loadFacts(context.Background(), b)
	wantCode(t, e, 0)
	keys := f.anchor.root.Keys
	meta := f.metadata.metadata.Keys
	f.Clear()
	f.Clear()
	if !reflect.DeepEqual(before, b.AnchorEnvelope) {
		t.Fatal("borrowed input cleared")
	}
	if !reflect.DeepEqual(keys[0], stateformat.PublicKeyMaterial{}) || !reflect.DeepEqual(meta[0], stateformat.MetadataKeyMaterial{}) {
		t.Fatal("owned key storage retained")
	}
	s := &StateModel{Roots: []RootRecord{{Root: stateformat.RootMaterial{Keys: []stateformat.PublicKeyMaterial{{PublicKey: [32]byte{1}}}}}}}
	sk := s.Roots[0].Root.Keys
	s.Clear()
	s.Clear()
	if sk[0].PublicKey[0] != 0 {
		t.Fatal("state nested owner retained")
	}
}
