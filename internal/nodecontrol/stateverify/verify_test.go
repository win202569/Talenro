package stateverify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

func assertBatch(t *testing.T, b Batch, want error) *SignatureEvidence {
	t.Helper()
	e, err := Verify(context.Background(), b)
	if err != want {
		t.Fatalf("want %v got %v", want, err)
	}
	if err != nil && e != nil {
		t.Fatal("partial evidence")
	}
	if err == nil && e == nil {
		t.Fatal("missing evidence")
	}
	if e != nil {
		t.Cleanup(e.Clear)
	}
	return e
}
func TestVerifyInputAndChain(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			b := signedBatch(t, n)
			e := assertBatch(t, b, nil)
			c := e.Coordinates()
			if c.ExpectedEpoch != 19 || c.ExpectedNodeID.String() != b.ExpectedNodeID || c.FinalRootVersion != uint64(n+1) || c.MetadataVersion != 6 || c.DesiredGeneration != 7 || c.DesiredSequence != 9 || len(e.Rotations()) != n {
				t.Fatal("coordinates", c)
			}
			if e.Anchor().EnvelopeDigest != sha256.Sum256(b.AnchorEnvelope) || e.Metadata().EnvelopeDigest != sha256.Sum256(b.MetadataEnvelope) || e.Desired().EnvelopeDigest != sha256.Sum256(b.DesiredEnvelope) {
				t.Fatal("identity")
			}
			for i, id := range e.Rotations() {
				if id.EnvelopeDigest != sha256.Sum256(b.Rotations[i]) {
					t.Fatal("rotation order")
				}
			}
		})
	}
	cases := []struct {
		name   string
		change func(*Batch)
		want   error
	}{
		{"33", func(b *Batch) { *b = signedBatch(t, 33) }, ErrInvalidInput},
		{"empty", func(b *Batch) { b.AnchorEnvelope = nil }, ErrInvalidInput},
		{"over", func(b *Batch) { b.DesiredEnvelope = bytes.Repeat([]byte("x"), 262145) }, ErrInvalidInput},
		{"uuid", func(b *Batch) { b.ExpectedNodeID = strings.ToUpper(b.ExpectedNodeID) }, ErrInvalidInput},
		{"nil uuid", func(b *Batch) { b.ExpectedNodeID = "00000000-0000-0000-0000-000000000000" }, ErrInvalidInput},
		{"variant", func(b *Batch) { b.ExpectedNodeID = "12345678-1234-4234-1234-123456789abc" }, ErrInvalidInput},
		{"epoch zero", func(b *Batch) { b.ExpectedEpoch = 0 }, ErrInvalidInput},
		{"epoch overflow", func(b *Batch) { b.ExpectedEpoch = 1 << 63 }, ErrInvalidInput},
		{"epoch mismatch", func(b *Batch) { b.ExpectedEpoch = 20 }, ErrBindingMismatch},
		{"node mismatch", func(b *Batch) { b.ExpectedNodeID = "12345678-1234-4234-8234-123456789abd" }, ErrBindingMismatch},
		{"gap", func(b *Batch) { *b = signedBatch(t, 2); b.Rotations = b.Rotations[1:] }, ErrBindingMismatch},
		{"order", func(b *Batch) {
			*b = signedBatch(t, 2)
			b.Rotations[0], b.Rotations[1] = b.Rotations[1], b.Rotations[0]
		}, ErrBindingMismatch},
		{"fork", func(b *Batch) {
			*b = signedBatch(t, 2)
			b.Rotations[1] = b.Rotations[0]
			b.ActiveRootEnvelope = b.Rotations[1]
		}, ErrBindingMismatch},
		{"active wrong kind", func(b *Batch) { b.ActiveRootEnvelope = b.AnchorEnvelope }, ErrInvalidInput},
		{"current quorum", func(b *Batch) {
			e := jsonObject(t, b.Rotations[0])
			e["current_signatures"] = []any{}
			b.Rotations[0] = jsonBytes(t, e)
			b.ActiveRootEnvelope = b.Rotations[0]
		}, ErrSignatureMismatch},
		{"new quorum", func(b *Batch) {
			e := jsonObject(t, b.Rotations[0])
			e["new_signatures"] = []any{}
			b.Rotations[0] = jsonBytes(t, e)
			b.ActiveRootEnvelope = b.Rotations[0]
		}, ErrSignatureMismatch},
		{"replace anchor", func(b *Batch) {
			p := envelopePayload(t, b.AnchorEnvelope)
			p["keys"] = []any{testKeyWire(4)}
			b.AnchorEnvelope = mint(t, "anchor", p, testPrivate(4))
		}, ErrSignatureMismatch},
		{"sequence", func(b *Batch) {
			b.Rotations[0] = changePayload(t, b.Rotations[0], "rotation", func(p map[string]any) {
				p["authority_sequence"] = "3"
				p["new_root_set"].(map[string]any)["authority_sequence"] = "3"
			})
			b.ActiveRootEnvelope = b.Rotations[0]
		}, ErrBindingMismatch},
		{"cross epoch", func(b *Batch) {
			b.Rotations[0] = changePayload(t, b.Rotations[0], "rotation", func(p map[string]any) {
				p["control_plane_authority_epoch"] = "20"
				p["new_root_set"].(map[string]any)["control_plane_authority_epoch"] = "20"
			})
			b.ActiveRootEnvelope = b.Rotations[0]
		}, ErrBindingMismatch},
		{"rootset only", func(b *Batch) { b.Rotations[0] = frozenFile(t, "anchor.payload.json") }, ErrInvalidInput},
		{"kind", func(b *Batch) { b.DesiredEnvelope = b.MetadataEnvelope }, ErrInvalidInput},
		{"C11", func(b *Batch) {
			e := jsonObject(t, b.MetadataEnvelope)
			e["schema_version"] = "node-trust-metadata.v1"
			b.MetadataEnvelope = jsonBytes(t, e)
		}, ErrInvalidInput},
		{"malformed unknown", func(b *Batch) {
			e := jsonObject(t, b.MetadataEnvelope)
			e["signatures"] = []any{map[string]any{"key_id": strings.Repeat("ab", 32), "signature": "AA"}}
			b.MetadataEnvelope = jsonBytes(t, e)
		}, ErrInvalidInput},
		{"duplicate", func(b *Batch) {
			e := jsonObject(t, b.MetadataEnvelope)
			s := e["signatures"].([]any)
			e["signatures"] = append(s, s[0])
			b.MetadataEnvelope = jsonBytes(t, e)
		}, ErrInvalidInput},
		{"unresigned payload", func(b *Batch) {
			e := jsonObject(t, b.DesiredEnvelope)
			p := envelopePayload(t, b.DesiredEnvelope)
			p["generation"] = "8"
			raw := jsonBytes(t, p)
			h := sha256.Sum256(raw)
			e["payload_b64url"] = base64.RawURLEncoding.EncodeToString(raw)
			e["payload_sha256"] = fmt.Sprintf("%x", h)
			b.DesiredEnvelope = jsonBytes(t, e)
		}, ErrSignatureMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { b := signedBatch(t, 1); tc.change(&b); assertBatch(t, b, tc.want) })
	}
	b := signedBatch(t, 0)
	b.ActiveRootEnvelope = frozenFile(t, "rotation.envelope.json")
	assertBatch(t, b, ErrInvalidInput)
}
func TestVerifyBindingsAndWindows(t *testing.T) {
	for _, field := range []string{"root_version", "control_plane_authority_epoch"} {
		t.Run(field, func(t *testing.T) {
			b := signedBatch(t, 1)
			b.MetadataEnvelope = changePayload(t, b.MetadataEnvelope, "metadata", func(p map[string]any) { p[field] = "20" })
			assertBatch(t, b, ErrBindingMismatch)
		})
	}
	for _, field := range []string{"node_id", "control_plane_authority_epoch"} {
		t.Run(field, func(t *testing.T) {
			b := signedBatch(t, 1)
			b.DesiredEnvelope = changePayload(t, b.DesiredEnvelope, "desired", func(p map[string]any) {
				p[field] = "20"
				if field == "node_id" {
					p[field] = "12345678-1234-4234-8234-123456789abd"
				}
			})
			assertBatch(t, b, ErrBindingMismatch)
		})
	}
	for _, status := range []string{"active", "retiring", "revoked", "missing"} {
		t.Run(status, func(t *testing.T) {
			b := signedBatch(t, 1)
			b.MetadataEnvelope = changePayload(t, b.MetadataEnvelope, "metadata", func(p map[string]any) {
				k := p["keys"].([]any)[0].(map[string]any)
				k["status"] = status
				if status == "revoked" {
					k["revoked_at"] = "2020-01-01T00:00:00Z"
					k["revocation_reason"] = "scheduled"
					p["revoked_key_ids"] = []any{k["key_id"]}
				}
				extraActive(t, p)
				if status == "missing" {
					keys := p["keys"].([]any)
					for i, item := range keys {
						if item.(map[string]any)["key_id"] == k["key_id"] {
							p["keys"] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			})
			want := error(nil)
			if status == "revoked" || status == "missing" {
				want = ErrBindingMismatch
			}
			assertBatch(t, b, want)
		})
	}
	windows := []struct {
		name, before, after, issued, until string
		want                               error
	}{
		{"lower exact", "2020-01-01T00:00:00Z", "2020-01-10T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z", nil},
		{"lower before", "2020-01-01T00:00:01Z", "2020-01-10T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z", ErrBindingMismatch},
		{"upper issued", "2019-12-31T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z", ErrBindingMismatch},
		{"key until exact", "2019-12-31T00:00:00Z", "2020-01-02T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z", nil},
		{"key until late", "2019-12-31T00:00:00Z", "2020-01-01T23:59:59Z", "2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z", ErrBindingMismatch},
		{"metadata until exact", "2019-12-31T00:00:00Z", "2020-01-10T00:00:00Z", "2020-01-07T00:00:00Z", "2020-01-08T00:00:00Z", nil},
		{"metadata until late", "2019-12-31T00:00:00Z", "2020-01-10T00:00:00Z", "2020-01-07T00:00:01Z", "2020-01-08T00:00:01Z", ErrBindingMismatch},
	}
	for _, tc := range windows {
		t.Run(tc.name, func(t *testing.T) {
			b := signedBatch(t, 1)
			// Metadata starts before the desired issue time, permitting an expired retiring key as a structural sample.
			b.MetadataEnvelope = changePayload(t, b.MetadataEnvelope, "metadata", func(p map[string]any) {
				p["issued_at"] = "2019-12-31T00:00:00Z"
				p["valid_until"] = "2020-01-07T00:00:00Z"
				p["next_refresh_at"] = "2020-01-05T00:00:00Z"
				k := p["keys"].([]any)[0].(map[string]any)
				k["status"] = "retiring"
				k["not_before"] = tc.before
				k["not_after"] = tc.after
				extraActive(t, p)
				if strings.HasPrefix(tc.name, "metadata until") {
					p["issued_at"] = "2020-01-01T00:00:00Z"
					p["valid_until"] = "2020-01-08T00:00:00Z"
					p["next_refresh_at"] = "2020-01-06T00:00:00Z"
				}
			})
			b.DesiredEnvelope = changePayload(t, b.DesiredEnvelope, "desired", func(p map[string]any) { p["issued_at"] = tc.issued; p["valid_until"] = tc.until })
			assertBatch(t, b, tc.want)
		})
	}
	b := signedBatch(t, 1)
	b.MetadataEnvelope = changePayload(t, b.MetadataEnvelope, "metadata", func(p map[string]any) {
		p["issued_at"] = "2020-01-02T00:00:00Z"
		p["keys"].([]any)[0].(map[string]any)["status"] = "retiring"
		extraActive(t, p)
	})
	assertBatch(t, b, nil)
}
func TestVerifyCompatibleMetadata(t *testing.T) {
	b := signedBatch(t, 1)
	first := assertBatch(t, b, nil)
	b.MetadataEnvelope = changePayload(t, b.MetadataEnvelope, "metadata", func(p map[string]any) { p["metadata_version"] = "7"; p["authority_sequence"] = "1" })
	second := assertBatch(t, b, nil)
	if first.Metadata() == second.Metadata() || first.Desired() != second.Desired() {
		t.Fatal("compatible identities")
	}
}
func TestVerifyConditionalOnly(t *testing.T) {
	b := signedBatch(t, 1)
	first := assertBatch(t, b, nil)
	// Replace the entire batch's root signer and separately supplied anchor.
	p := envelopePayload(t, b.AnchorEnvelope)
	p["keys"] = []any{testKeyWire(4)}
	b.AnchorEnvelope = mint(t, "anchor", p, testPrivate(4))
	p = envelopePayload(t, b.Rotations[0])
	p["new_root_set"].(map[string]any)["keys"] = []any{testKeyWire(4)}
	b.Rotations[0] = mint(t, "rotation", p, testPrivate(4))
	b.ActiveRootEnvelope = b.Rotations[0]
	b.MetadataEnvelope = mint(t, "metadata", envelopePayload(t, b.MetadataEnvelope), testPrivate(4))
	second := assertBatch(t, b, nil)
	if first.Anchor() == second.Anchor() {
		t.Fatal("conditional anchor not recorded")
	}
	a, e := stateformat.ParseEnvelope(stateformat.Desired, b.DesiredEnvelope)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	if a.Bindings().ValidUntil.Year() != 2020 {
		t.Fatal("fixture not past")
	}
	old := second.Desired()
	b.DesiredEnvelope[0] ^= 1
	if second.Desired() != old {
		t.Fatal("output aliases input")
	}
}
