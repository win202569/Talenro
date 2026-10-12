package statecomposition

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
	"testing"
	"time"
)

func requireOutcome(t *testing.T, m *CompositionModel, e, want error) {
	t.Helper()
	if want == nil {
		if e != nil || m == nil {
			t.Fatalf("missing success: %v", e)
		}
		t.Cleanup(m.Clear)
	} else if !errors.Is(e, want) || m != nil {
		if m != nil {
			m.Clear()
		}
		t.Fatalf("want %v, got model=%v err=%v", want, m != nil, e)
	}
}
func outcome(t *testing.T, in Input, want error) {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	requireOutcome(t, m, e, want)
}
func badSignature(t *testing.T, raw []byte, field string) []byte {
	e := object(t, raw)
	if field == "signature" {
		b, _ := base64.RawURLEncoding.DecodeString(e[field].(string))
		b[0] ^= 1
		e[field] = base64.RawURLEncoding.EncodeToString(b)
	} else {
		s := e[field].([]any)[0].(map[string]any)
		b, _ := base64.RawURLEncoding.DecodeString(s["signature"].(string))
		b[0] ^= 1
		s["signature"] = base64.RawURLEncoding.EncodeToString(b)
	}
	return encode(t, e)
}
func TestCompositionEvaluate(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		outcome(t, fixture(t, n), nil)
	}
	for _, field := range []string{"signature", "signatures", "current_signatures", "new_signatures", "empty current", "empty new"} {
		t.Run(field, func(t *testing.T) {
			in := fixture(t, 1)
			r := in.Target.Desired
			if field == "signatures" {
				r = in.Target.Metadata
			} else if field != "signature" {
				r = in.Target.Links[0].Rotation
			}
			raw := material(&in, r).Raw
			if field == "empty current" || field == "empty new" {
				e := object(t, raw)
				f := "current_signatures"
				if field == "empty new" {
					f = "new_signatures"
				}
				e[f] = []any{}
				raw = encode(t, e)
			} else {
				raw = badSignature(t, raw, field)
			}
			nr := replaceEnvelope(t, &in, r, raw)
			switch r.Kind {
			case statematerials.Desired:
				in.Target.Desired = nr
			case statematerials.Metadata:
				in.Target.Metadata = nr
			default:
				in.Target.Links[0].Rotation = nr
			}
			m, e := statematerials.Evaluate(context.Background(), in.Materials)
			if e != nil {
				t.Fatal("masked crypto test", e)
			}
			m.Clear()
			outcome(t, in, ErrSignatureMismatch)
		})
	}
	cases := []struct {
		name string
		mut  func(*Input)
		want error
	}{
		{"duplicate", func(in *Input) { in.Materials.Materials = append(in.Materials.Materials, in.Materials.Materials[0]) }, ErrDuplicateMaterial},
		{"missing", func(in *Input) { in.Materials.Materials = in.Materials.Materials[1:] }, ErrMissingMaterial},
		{"bad raw", func(in *Input) { material(in, in.Target.Metadata).Raw = []byte("bad") }, ErrInvalidMaterial},
		{"expected mismatch", func(in *Input) { expected(in, in.Target.Desired).Desired.Generation++ }, ErrBindingMismatch},
		{"revoked", func(in *Input) {
			r := compatibleMetadata(t, in, 7, true)
			in.Target.Metadata = r
			expected(in, in.Target.Desired).Desired.MetadataPayload = r.Payload
		}, ErrBindingMismatch},
		{"missing key", func(in *Input) {
			r := in.Target.Desired
			e := object(t, material(in, r).Raw)
			e["key_id"] = wireKey(3)["key_id"]
			in.Target.Desired = replaceEnvelope(t, in, r, encode(t, e))
			expected(in, in.Target.Desired).Desired.SigningKeyID = digest(t, e["key_id"].(string))
		}, ErrBindingMismatch},
		{"key static window", func(in *Input) {
			p := object(t, payload(t, material(in, in.Target.Metadata).Raw))
			p["keys"].([]any)[0].(map[string]any)["not_before"] = "2020-01-01T00:00:01Z"
			addActiveKey(p)
			r := addEnvelope(t, in, statematerials.Metadata, mint(t, "metadata", p, private(1)), statematerials.MaterialRef{}, tailRef(*in), statematerials.MaterialRef{})
			in.Target.Metadata = r
			expected(in, in.Target.Desired).Desired.MetadataPayload = r.Payload
		}, ErrBindingMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { in := fixture(t, 1); tc.mut(&in); outcome(t, in, tc.want) })
	}
	m, e := Evaluate(nil, fixture(t, 1))
	requireOutcome(t, m, e, ErrInvalidInput)
}
func addActiveKey(p map[string]any) {
	k := wireKey(3)
	k["status"] = "active"
	k["not_before"] = "2019-12-31T00:00:00Z"
	k["not_after"] = "2020-01-10T00:00:00Z"
	k["revoked_at"] = nil
	k["revocation_reason"] = nil
	k["incident_id"] = nil
	keys := append(p["keys"].([]any), k)
	if keys[0].(map[string]any)["key_id"].(string) > keys[1].(map[string]any)["key_id"].(string) {
		keys[0], keys[1] = keys[1], keys[0]
	}
	p["keys"] = keys
}
func TestCompositionErrorMapping(t *testing.T) {
	cases := []struct{ in, want error }{
		{nil, nil}, {statematerials.ErrInvalidInput, ErrInvalidInput}, {statematerials.ErrLimitExceeded, ErrLimitExceeded}, {statematerials.ErrInvalidMaterial, ErrInvalidMaterial}, {statematerials.ErrMissingMaterial, ErrMissingMaterial}, {statematerials.ErrDuplicateMaterial, ErrDuplicateMaterial}, {statematerials.ErrClearedModel, ErrClearedModel}, {statematerials.ErrReferenceMismatch, ErrBindingMismatch}, {stateverify.ErrBindingMismatch, ErrBindingMismatch}, {stateformat.ErrInvalidFormat, ErrInvalidMaterial}, {stateverify.ErrInvalidInput, ErrInvalidMaterial}, {stateverify.ErrSignatureMismatch, ErrSignatureMismatch}, {stateverify.ErrCanceled, context.Canceled}, {context.Canceled, context.Canceled}, {context.DeadlineExceeded, context.DeadlineExceeded}, {errors.New("sensitive dynamic material"), ErrInvalidInput},
	}
	for _, tc := range cases {
		for _, wrapped := range []bool{false, true} {
			e := tc.in
			if wrapped && e != nil {
				e = fmt.Errorf("lower: %w", e)
			}
			if got := mapError(context.Background(), e); !errors.Is(got, tc.want) {
				t.Fatalf("%v: %v", tc.in, got)
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := mapError(ctx, stateverify.ErrSignatureMismatch); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if e := mapError(ctx, nil); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
func TestCompositionEvidenceBinding(t *testing.T) {
	names := []string{"anchor P", "anchor E", "metadata P", "metadata E", "desired P", "desired E", "rotation 0 P", "rotation 0 E", "rotation 1 P", "rotation 1 E", "rotation order", "rotation count", "node", "epoch", "root version", "metadata version", "desired generation", "desired sequence", "desired key"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			in := checkedOwned(t, fixture(t, 2))
			s, e := selectBatch(context.Background(), in)
			if e != nil {
				t.Fatal(e)
			}
			defer s.Clear()
			proof, e := stateverify.Verify(context.Background(), s.batch)
			if e != nil {
				t.Fatal(e)
			}
			defer proof.Clear()
			if e := checkEvidence(context.Background(), in, s, proof); e != nil {
				t.Fatal(e)
			}
			switch name {
			case "anchor P":
				in.Anchor.Payload[0] ^= 1
			case "anchor E":
				in.Anchor.Envelope[0] ^= 1
			case "metadata P":
				in.Target.Metadata.Payload[0] ^= 1
			case "metadata E":
				in.Target.Metadata.Envelope[0] ^= 1
			case "desired P":
				in.Target.Desired.Payload[0] ^= 1
			case "desired E":
				in.Target.Desired.Envelope[0] ^= 1
			case "rotation 0 P":
				in.Target.Links[0].Rotation.Payload[0] ^= 1
			case "rotation 0 E":
				in.Target.Links[0].Rotation.Envelope[0] ^= 1
			case "rotation 1 P":
				in.Target.Links[1].Rotation.Payload[0] ^= 1
			case "rotation 1 E":
				in.Target.Links[1].Rotation.Envelope[0] ^= 1
			case "rotation order":
				in.Target.Links[0], in.Target.Links[1] = in.Target.Links[1], in.Target.Links[0]
			case "rotation count":
				in.Target.Links = in.Target.Links[:1]
			case "node":
				in.Anchor.Scope.NodeID = "22345678-1234-4234-8234-123456789abc"
			case "epoch":
				in.Anchor.Scope.Epoch++
			case "root version":
				s.rootVersion++
			case "metadata version":
				s.metadataVersion++
			case "desired generation":
				s.desiredGeneration++
			case "desired sequence":
				s.desiredSequence++
			case "desired key":
				s.desiredKeyID = contracts.Digest{1}
			}
			if e := checkEvidence(context.Background(), in, s, proof); e != ErrBindingMismatch {
				t.Fatal(e)
			}
		})
	}
	in := checkedOwned(t, fixture(t, 1))
	s, e := selectBatch(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Clear()
	if e := checkEvidence(context.Background(), in, s, nil); e != ErrBindingMismatch {
		t.Fatal(e)
	}
	p, e := stateverify.Verify(context.Background(), s.batch)
	if e != nil {
		t.Fatal(e)
	}
	p.Clear()
	if e := checkEvidence(context.Background(), in, s, p); e != ErrBindingMismatch {
		t.Fatal(e)
	}
}
func TestCompositionDeclaredAssociation(t *testing.T) {
	in := fixture(t, 1)
	other := compatibleMetadata(t, &in, 7, false)
	outcome(t, in, nil)
	in.Target.Metadata = other
	outcome(t, in, ErrBindingMismatch)
	expected(&in, in.Target.Desired).Desired.MetadataPayload = other.Payload
	outcome(t, in, nil)
}
func TestCompositionHistoricalTarget(t *testing.T) {
	in := fixture(t, 1)
	revoked := compatibleMetadata(t, &in, 7, true)
	outcome(t, in, nil)
	in.Target.Metadata = revoked
	expected(&in, in.Target.Desired).Desired.MetadataPayload = revoked.Payload
	outcome(t, in, ErrBindingMismatch)
	in = fixture(t, 1)
	other := compatibleMetadata(t, &in, 7, false)
	replaceEnvelope(t, &in, other, badSignature(t, material(&in, other).Raw, "signatures"))
	outcome(t, in, nil)
	for _, mode := range []string{"schema", "raw", "expected", "missing relation"} {
		t.Run(mode, func(t *testing.T) {
			in := fixture(t, 1)
			other := compatibleMetadata(t, &in, 7, false)
			want := ErrInvalidMaterial
			switch mode {
			case "schema":
				o := object(t, material(&in, other).Raw)
				o["schema_version"] = "unknown"
				replaceEnvelope(t, &in, other, encode(t, o))
			case "raw":
				material(&in, other).Raw = []byte("broken")
			case "expected":
				expected(&in, other).Metadata.Version++
				want = ErrBindingMismatch
			case "missing relation":
				expected(&in, other).Metadata.RootPayload[0] ^= 1
				want = ErrMissingMaterial
			}
			outcome(t, in, want)
		})
	}
}
