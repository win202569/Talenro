package staterecovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRecoveryOwnership(t *testing.T) {
	in := memoryScan()
	r, e := RecoverModel(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	rs := r.Summary()
	in.A.BlobDigest = digest(99)
	copySummary := r.Summary()
	copySummary.BlobDigest = digest(98)
	if r.Summary() != rs {
		t.Fatal("borrowed recovery")
	}
	m := memoryPrepared(t)
	before := m.Summary()
	cs := m.Summary()
	cs.AttemptID = digest(99)
	if m.Summary() != before {
		t.Fatal("borrowed commit")
	}
	n, e := AdvanceModel(context.Background(), m, memoryEvent(m, CandidateWritten))
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	m.Clear()
	if n.Summary().Phase != Written || m.Summary() != (CommitSummary{}) {
		t.Fatal("shared successor")
	}
	requireAdvanceError(t, m, Event{}, ErrStaleModel)
	done := memoryPhase(t, NeedsRecovery)
	ci := memoryConfirmation(true)
	c, e := ConfirmModel(context.Background(), done, ci)
	if e != nil {
		t.Fatal(e)
	}
	want := c.Summary()
	ci.Scan.B.BlobDigest = digest(99)
	cr := c.Summary()
	cr.AttemptID = digest(99)
	done.Clear()
	if c.Summary() != want {
		t.Fatal("borrowed confirmation")
	}
	requireConfirmError(t, done, ci, ErrStaleModel)
	c.Clear()
	c.Clear()
	if c.Summary() != (ConfirmationSummary{}) {
		t.Fatal("clear confirmation")
	}
	var nilCommit *CommitModel
	nilCommit.Clear()
	requireAdvanceError(t, nilCommit, Event{}, ErrStaleModel)
	requireConfirmError(t, nilCommit, ConfirmationInput{}, ErrStaleModel)
	var nilResult *ConfirmationModel
	nilResult.Clear()
	if nilResult.Summary() != (ConfirmationSummary{}) {
		t.Fatal("nil summary")
	}
	if (&CommitModel{}).Summary() != (CommitSummary{}) {
		t.Fatal("zero summary")
	}
	for reason := CorruptSlot; reason <= UnexpectedFuture; reason++ {
		err := &SecurityError{Reason: reason}
		if !errors.Is(err, ErrSecurityFault) || strings.Contains(err.Error(), memoryScan().ExpectedScope.NodeID) || strings.Contains(err.Error(), fmt.Sprintf("%x", digest(10))) {
			t.Fatal("error identity leak")
		}
	}
}

// cancelOnCheck provides deterministic cancellation at the final check, not a
// sleep-dependent race. Counts pin each public entry's post-validation check.
type cancelOnCheck struct {
	context.Context
	calls, at int
}

func (c *cancelOnCheck) Err() error {
	c.calls++
	if c.calls >= c.at {
		return context.Canceled
	}
	return nil
}
func TestRecoveryCancellation(t *testing.T) {
	in := memoryScan()
	p := memoryProposal(in)
	m := memoryPrepared(t)
	ev := memoryEvent(m, CandidateWritten)
	done := memoryPhase(t, NeedsRecovery)
	ci := memoryConfirmation(true)
	calls := []struct {
		name string
		last int
		call func(context.Context) (bool, error)
	}{
		{"recover", 2, func(ctx context.Context) (bool, error) { r, e := RecoverModel(ctx, in); return r == nil, e }},
		{"prepare", 3, func(ctx context.Context) (bool, error) { r, e := PrepareModel(ctx, in, p); return r == nil, e }},
		{"advance", 2, func(ctx context.Context) (bool, error) { r, e := AdvanceModel(ctx, m, ev); return r == nil, e }},
		{"confirm", 4, func(ctx context.Context) (bool, error) { r, e := ConfirmModel(ctx, done, ci); return r == nil, e }},
	}
	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			nilResult, e := c.call(nil)
			if !nilResult || !errors.Is(e, ErrInvalidInput) {
				t.Fatal("nil context", e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			nilResult, e = c.call(ctx)
			if !nilResult || !errors.Is(e, context.Canceled) {
				t.Fatal("precancel", e)
			}
			late := &cancelOnCheck{Context: context.Background(), at: c.last}
			nilResult, e = c.call(late)
			if !nilResult || !errors.Is(e, context.Canceled) {
				t.Fatal("late cancel", e)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, e := RecoverModel(ctx, RecoveryInput{})
	if r != nil || !errors.Is(e, context.Canceled) {
		t.Fatal("cancel priority", e)
	}
	cm, e := PrepareModel(ctx, RecoveryInput{}, ProposalObservation{})
	if cm != nil || !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	cm, e = AdvanceModel(ctx, nil, Event{})
	if cm != nil || !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	conf, e := ConfirmModel(ctx, nil, ConfirmationInput{})
	if conf != nil || !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
