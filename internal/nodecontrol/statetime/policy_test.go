package statetime

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Inclusive expiry, clock ratcheting, or cross-stream sequence comparison breaks these cases.
func TestTimePolicyWindows(t *testing.T) {
	for _, tc := range []struct {
		at    string
		delta time.Duration
		ok    bool
	}{{"2026-10-09T00:00:00Z", 0, true}, {"2026-10-09T00:05:00Z", -1, true}, {"2026-10-09T00:05:00Z", 0, false}, {"2026-10-08T23:59:59Z", 0, false}, {"9999-12-31T23:59:59Z", 0, false}, {"0001-01-01T00:00:01Z", 0, false}} {
		i := base(t)
		i.Time.Now = instant(tc.at).Add(tc.delta)
		i.Time.Floor = instant("0001-01-01T00:00:01Z")
		r, e := EvaluateModel(context.Background(), i)
		if (e == nil) != tc.ok || (!tc.ok && !errors.Is(e, ErrExpired)) {
			t.Fatalf("%+v %v", tc, e)
		}
		r.Clear()
	}
	for name, mut := range map[string]func(*ModelInput){"zero": func(i *ModelInput) { i.Time.Now = time.Time{} }, "floor": func(i *ModelInput) { i.Time.Floor = i.Time.Now.Add(time.Second) }, "future issued": func(i *ModelInput) { changeAtt(t, i, "issued_at", "2026-10-09T00:01:01Z") }, "metadata expired": func(i *ModelInput) { i.Time.Now = instant("2026-10-16T00:00:00Z") }, "key before": func(i *ModelInput) {
		p := payload(t, i.MetadataEnvelope)
		signingKey(p)["not_before"] = "2026-10-09T00:00:01Z"
		setMetadata(t, i, p)
	}, "key until": func(i *ModelInput) {
		p := payload(t, i.MetadataEnvelope)
		signingKey(p)["not_after"] = "2026-10-09T00:04:59Z"
		setMetadata(t, i, p)
	}, "invalid year": func(i *ModelInput) { i.Time.Now = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) }} {
		t.Run(name, func(t *testing.T) {
			i := base(t)
			mut(&i)
			r, e := EvaluateModel(context.Background(), i)
			want := map[string]error{"zero": ErrInvalidTime, "floor": ErrInvalidTime, "future issued": ErrExpired, "metadata expired": ErrExpired, "key before": ErrBindingMismatch, "key until": ErrBindingMismatch, "invalid year": ErrInvalidTime}[name]
			if want == nil || !errors.Is(e, want) || r != nil {
				t.Fatalf("got %v, want %v, result=%v", e, want, r != nil)
			}
		})
	}
	i := base(t)
	changeAtt(t, &i, "issued_at", "2026-10-15T23:59:00Z")
	changeAtt(t, &i, "valid_until", "2026-10-16T00:01:00Z")
	i.Time.Now = instant("2026-10-15T23:59:30Z")
	if r, e := EvaluateModel(context.Background(), i); !errors.Is(e, ErrBindingMismatch) || r != nil {
		t.Fatal("metadata upper bound")
	}
}
func TestTimePolicyCheckpoint(t *testing.T) {
	for _, tc := range []struct {
		s           string
		ok, advance bool
	}{{"6", false, false}, {"7", true, false}, {"8", true, true}} {
		i := base(t)
		changeAtt(t, &i, "node_authority_checkpoint_sequence", tc.s)
		r, e := EvaluateModel(context.Background(), i)
		if (e == nil) != tc.ok || (!tc.ok && (!errors.Is(e, ErrCheckpointRollback) || r != nil)) {
			t.Fatal(tc, e)
		}
		if tc.ok && (r.Summary().AdvanceRequired != tc.advance || r.Summary().OldCheckpoint != 7) {
			t.Fatal("proposal")
		}
		r.Clear()
	}
}
func TestTimePolicyNoAuthority(t *testing.T) {
	i := base(t)
	oldTime, oldCheckpoint := i.Time, i.Checkpoint
	for n := 0; n < 2; n++ {
		r, e := EvaluateModel(context.Background(), i)
		if e != nil || r.Summary().AdvanceRequired {
			t.Fatal("repeat model", e)
		}
		r.Clear()
	}
	if i.Time != oldTime || i.Checkpoint != oldCheckpoint {
		t.Fatal("mutation")
	}
	i.Request.ExpectedNonce[0] = 42
	i.Request.RequestID = digest(42)
	changeAtt(t, &i, "request_nonce", b64(i.Request.ExpectedNonce[:]))
	r, e := EvaluateModel(context.Background(), i)
	if e != nil || r.Summary().AdvanceRequired {
		t.Fatal("same checkpoint new request", e)
	}
	r.Clear()
	if r.Summary() != (EvidenceSummary{}) {
		t.Fatal("clear")
	}
	var nilEvidence *EvidenceModel
	nilEvidence.Clear()
	if nilEvidence.Summary() != (EvidenceSummary{}) {
		t.Fatal("nil")
	}
}
