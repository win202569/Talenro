package statetime

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeEvidenceRecheck(t *testing.T) {
	i := base(t)
	r, e := EvaluateModel(context.Background(), i)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Clear()
	if e = RecheckModel(context.Background(), r, i); e != nil {
		t.Fatal(e)
	}
	j := i
	j.Time.Now = j.Time.Now.Add(time.Second)
	j.Time.Floor = j.Time.Floor.Add(time.Second)
	j.Time.ObservationID = digest(88)
	if e = RecheckModel(context.Background(), r, j); e != nil {
		t.Fatal("fresh time", e)
	}
	for name, mut := range map[string]func(*ModelInput){"provider": func(i *ModelInput) { i.Time.ProviderID = digest(90) }, "now backward": func(i *ModelInput) { i.Time.Now = i.Time.Now.Add(-time.Second) }, "floor backward": func(i *ModelInput) { i.Time.Floor = i.Time.Floor.Add(-time.Second) }, "expired": func(i *ModelInput) { i.Time.Now = instant("2026-10-09T00:05:00Z") }, "request": func(i *ModelInput) { i.Request.RequestID = digest(90) }, "nonce": func(i *ModelInput) {
		i.Request.ExpectedNonce[0] = 1
		changeAtt(t, i, "request_nonce", b64(i.Request.ExpectedNonce[:]))
	}, "checkpoint": func(i *ModelInput) { i.Checkpoint.Sequence = 6 }, "metadata source": func(i *ModelInput) { i.Metadata.SourceID = digest(90) }, "status": func(i *ModelInput) { i.Metadata.Status = Unknown }} {
		t.Run(name, func(t *testing.T) {
			j := base(t)
			mut(&j)
			if e := RecheckModel(context.Background(), r, j); e == nil {
				t.Fatal("stale accepted")
			}
		})
	}
}
func TestTimeEvidenceOwnership(t *testing.T) {
	i := base(t)
	r, e := EvaluateModel(context.Background(), i)
	if e != nil {
		t.Fatal(e)
	}
	s := r.Summary()
	i.AttestationEnvelope[0] ^= 1
	if r.Summary() != s {
		t.Fatal("borrowed")
	}
	if e := RecheckModel(context.Background(), r, i); e == nil {
		t.Fatal("not reverified")
	}
	s.Nonce[0] ^= 1
	if r.Summary().Nonce == s.Nonce {
		t.Fatal("summary alias")
	}
	r.Clear()
	r.Clear()
	if r.Summary() != (EvidenceSummary{}) || !errors.Is(RecheckModel(context.Background(), r, base(t)), ErrStaleEvidence) {
		t.Fatal("cleared usable")
	}
	if !errors.Is(RecheckModel(context.Background(), nil, base(t)), ErrStaleEvidence) {
		t.Fatal("nil")
	}
	if !errors.Is(RecheckModel(context.Background(), &EvidenceModel{}, base(t)), ErrStaleEvidence) {
		t.Fatal("zero")
	}
}

type cancelAfter struct {
	context.Context
	calls, after int
}

func (c *cancelAfter) Err() error {
	c.calls++
	if c.calls >= c.after {
		return context.Canceled
	}
	return nil
}
func TestTimeEvidenceCancellation(t *testing.T) {
	for n := 1; n <= 5; n++ {
		ctx := &cancelAfter{Context: context.Background(), after: n}
		r, e := EvaluateModel(ctx, base(t))
		if r != nil || !errors.Is(e, context.Canceled) {
			t.Fatal(n, e)
		}
	}
	i := base(t)
	r, e := EvaluateModel(context.Background(), i)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Clear()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(RecheckModel(ctx, r, i), context.Canceled) {
		t.Fatal("recheck cancel")
	}
	if !errors.Is(RecheckModel(nil, r, i), ErrInvalidInput) {
		t.Fatal("nil context")
	}
}
