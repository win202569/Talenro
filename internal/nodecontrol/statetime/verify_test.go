package statetime

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

// Removing E pinning would accept replacement raw bytes despite equal P.
func TestTimeVerifyMetadataEnvelopeSubstitution(t *testing.T) {
	in := base(t)
	var envelope map[string]any
	decodeObject(t, in.MetadataEnvelope, &envelope)
	p, err := base64.RawURLEncoding.DecodeString(envelope["payload_b64url"].(string))
	if err != nil {
		t.Fatal(err)
	}
	message := append([]byte("TALENRO-NODE-STATE-METADATA-V1\x00"), p...)
	makeEnvelope := func(n byte) []byte {
		envelope["signatures"] = []any{map[string]any{"key_id": hexID(keyID(n)), "signature": b64(ed25519.Sign(private(n), message))}}
		return canon(t, envelope)
	}
	first, err := stateformat.ParseEnvelope(stateformat.Metadata, makeEnvelope(1))
	if err != nil {
		t.Fatal(err)
	}
	defer first.Clear()
	second, err := stateformat.ParseEnvelope(stateformat.Metadata, makeEnvelope(2))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Clear()
	if !bytes.Equal(first.Payload(), second.Payload()) || first.PayloadDigest() != second.PayloadDigest() || bytes.Equal(first.Envelope(), second.Envelope()) || first.EnvelopeDigest() == second.EnvelopeDigest() {
		t.Fatal("fixture must have identical P and distinct E")
	}
	in.MetadataEnvelope = first.Envelope()
	in.Metadata.Payload = first.PayloadDigest()
	in.Metadata.Envelope = first.EnvelopeDigest()
	accepted, err := EvaluateModel(context.Background(), in)
	if err != nil {
		t.Fatal("original model", err)
	}
	accepted.Clear()
	in.MetadataEnvelope = second.Envelope()
	if result, err := EvaluateModel(context.Background(), in); result != nil || !errors.Is(err, ErrBindingMismatch) {
		result.Clear()
		t.Fatalf("old observation accepted substituted E: %v", err)
	}
	// Updating this model is not root authentication or a durable commit.
	in.Metadata.Envelope = second.EnvelopeDigest()
	accepted, err = EvaluateModel(context.Background(), in)
	if err != nil {
		t.Fatal("updated model", err)
	}
	defer accepted.Clear()
	if accepted.Summary().MetadataPayload != first.PayloadDigest() || accepted.Summary().MetadataEnvelope != second.EnvelopeDigest() {
		t.Fatal("wrong bound identity")
	}
}

// Dropping explicit scope/identity comparisons must fail these signed-input tests.
func TestTimeVerifyBindings(t *testing.T) {
	for name, mut := range map[string]func(*ModelInput){"nonce": func(i *ModelInput) { i.Request.ExpectedNonce[0] = 1 }, "node": func(i *ModelInput) { i.Checkpoint.NodeID[0] ^= 1 }, "epoch": func(i *ModelInput) { i.Request.Epoch++ }, "version": func(i *ModelInput) { i.Metadata.Version++ }, "state": func(i *ModelInput) { i.Checkpoint.StateID = digest(99) }, "payload": func(i *ModelInput) { i.Metadata.Payload = digest(99) }, "envelope": func(i *ModelInput) { i.Metadata.Envelope = digest(99) }, "zero checkpoint": func(i *ModelInput) { i.Checkpoint.Sequence = 0 }, "overflow": func(i *ModelInput) { i.Checkpoint.Sequence = 1 << 63 }, "zero source": func(i *ModelInput) { i.Metadata.SourceID = contracts.Digest{} }, "zero request": func(i *ModelInput) { i.Request.RequestID = contracts.Digest{} }, "zero provider": func(i *ModelInput) { i.Time.ProviderID = contracts.Digest{} }} {
		t.Run(name, func(t *testing.T) {
			i := base(t)
			mut(&i)
			f, e := verifyModel(context.Background(), i)
			if e == nil || f != nil {
				t.Fatal("accepted", e)
			}
		})
	}
	for _, s := range []ObservationStatus{Unknown, Unavailable, Rollback, 255} {
		for n := 0; n < 4; n++ {
			i := base(t)
			switch n {
			case 0:
				i.Metadata.Status = s
			case 1:
				i.Time.Status = s
			case 2:
				i.Request.Status = s
			case 3:
				i.Checkpoint.Status = s
			}
			f, e := verifyModel(context.Background(), i)
			if e == nil || f != nil {
				t.Fatal("status")
			}
		}
	}
	if f, e := verifyModel(nil, base(t)); f != nil || !errors.Is(e, ErrInvalidInput) {
		t.Fatal("nilctx", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if f, e := verifyModel(ctx, ModelInput{}); f != nil || !errors.Is(e, context.Canceled) {
		t.Fatal("cancel", e)
	}
}
func TestTimeVerifySignature(t *testing.T) {
	i := base(t)
	f, e := verifyModel(context.Background(), i)
	if e != nil {
		t.Fatal(e)
	}
	f.Clear()
	for _, domain := range []string{"TALENRO-NODE-DESIRED-STATE-V1", "TALENRO-NODE-TIME-ATTESTATION-V1\x00"} {
		j := base(t)
		j.AttestationEnvelope = attestation(t, payload(t, j.AttestationEnvelope), 1, domain)
		if f, e := verifyModel(context.Background(), j); f != nil || !errors.Is(e, ErrSignatureMismatch) {
			t.Fatal("domain", e)
		}
	}
	for _, k := range []string{"signature", "key_id"} {
		j := base(t)
		var m map[string]any
		decodeObject(t, j.AttestationEnvelope, &m)
		if k == "signature" {
			m[k] = b64(makeNonzero(64))
		} else {
			m[k] = hexID(keyID(2))
		}
		j.AttestationEnvelope = canon(t, m)
		if f, e := verifyModel(context.Background(), j); e == nil || f != nil {
			t.Fatal(k, e)
		}
	}
}
func TestTimeVerifyKeyStatus(t *testing.T) {
	for _, status := range []string{"active", "retiring", "revoked", "missing"} {
		i := base(t)
		p := payload(t, i.MetadataEnvelope)
		k := signingKey(p)
		if status == "missing" {
			keys := p["keys"].([]any)
			out := []any{}
			for _, v := range keys {
				if v.(map[string]any)["key_id"] != k["key_id"] {
					out = append(out, v)
				}
			}
			p["keys"] = out
		} else {
			k["status"] = status
			if status == "revoked" {
				k["revoked_at"] = "2026-10-09T00:00:00Z"
				k["revocation_reason"] = "scheduled"
				p["revoked_key_ids"] = []any{k["key_id"]}
			}
		}
		setMetadata(t, &i, p)
		f, e := verifyModel(context.Background(), i)
		want := status == "active" || status == "retiring"
		if (e == nil) != want {
			t.Fatal(status, e)
		}
		f.Clear()
	}
}
