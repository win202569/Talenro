package statepolicy

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func TestEvaluateModel(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := modelInput(t, n)
		c, e := EvaluateModel(context.Background(), in)
		wantCode(t, e, 0)
		if c == nil {
			t.Fatal("no candidate")
		}
		defer c.Clear()
		s := c.State()
		defer s.Clear()
		if s.Metadata.Version != 6 || s.Desired.Version != 7 || len(s.Roots) != n+1 || c.NoChange() || !c.ValidUntil().Equal(stamp("2020-01-02T00:00:00Z")) {
			t.Fatal("candidate coordinates")
		}
		// A proposal cannot be reused as an observed committed snapshot.
		in.Old = s
		c2, e := EvaluateModel(context.Background(), in)
		if e == nil || c2 != nil {
			t.Fatal("proposal adopted as trusted state")
		}
	}
	for _, name := range []string{"repeat", "signature", "format", "missing deployment", "old rollback", "time expired", "old content", "history missing"} {
		t.Run(name, func(t *testing.T) {
			in := modelInput(t, 1)
			in.Old = previousModel(t, in)
			want := Code(0)
			switch name {
			case "signature":
				in.Batch.DesiredEnvelope = changePayload(t, in.Batch.DesiredEnvelope, "desired", func(p map[string]any) { p["node_id"] = "22345678-1234-4234-8234-123456789abc" })
				want = Rejected
			case "format":
				in.Batch.DesiredEnvelope = []byte("{}")
				want = InvalidInput
			case "missing deployment":
				in.Deployment.Status = Unavailable
				want = AssuranceMissing
			case "old rollback":
				in.Old.Status = Rollback
				want = SecurityFault
			case "time expired":
				in.Time.Now = stamp("2020-01-02T00:00:00Z")
				want = Rejected
			case "old content":
				in.Old.Metadata.Payload[0] ^= 1
				want = SecurityFault
			case "history missing":
				in.History = []KeyHistoryObservation{{Status: Available, ObservationID: contracts.Digest{8}}}
				want = AssuranceMissing
			}
			before := append([]byte(nil), in.Batch.DesiredEnvelope...)
			c, e := EvaluateModel(context.Background(), in)
			wantCode(t, e, want)
			if e != nil && c != nil {
				t.Fatal("partial candidate")
			}
			if c != nil {
				defer c.Clear()
				if !c.NoChange() {
					t.Fatal("repeat writes state")
				}
			}
			if !reflect.DeepEqual(before, in.Batch.DesiredEnvelope) {
				t.Fatal("input modified")
			}
		})
	}
	in := modelInput(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in.Batch.DesiredEnvelope = nil
	c, e := EvaluateModel(ctx, in)
	wantCode(t, e, Canceled)
	if c != nil {
		t.Fatal("cancel partial")
	}
	c, e = EvaluateModel(nil, in)
	wantCode(t, e, InvalidInput)
	if c != nil {
		t.Fatal("nil ctx partial")
	}
}
func TestCandidateRecheck(t *testing.T) {
	for _, name := range []string{"fresh", "expired", "clock back", "floor back", "provider", "token", "same token changed state", "bytes", "history", "zero candidate", "cleared", "ownership"} {
		t.Run(name, func(t *testing.T) {
			in := modelInput(t, 1)
			in.Old = previousModel(t, in)
			c, e := EvaluateModel(context.Background(), in)
			wantCode(t, e, 0)
			defer c.Clear()
			want := true
			switch name {
			case "fresh":
				in.Time.Now = in.Time.Now.Add(time.Second)
				in.Time.ObservationID[0]++
			case "expired":
				in.Time.Now = c.ValidUntil()
				want = false
			case "clock back":
				in.Time.Now = in.Time.Now.Add(-time.Second)
				want = false
			case "floor back":
				in.Time.Floor = in.Time.Floor.Add(-time.Second)
				want = false
			case "provider":
				in.Time.ProviderID[0]++
				want = false
			case "token":
				in.Old.StateID[0]++
				want = false
			case "same token changed state":
				in.Old.Keys[0].DeadlinesKnown = true
				in.Old.Keys[0].LastSnapshotUntil = in.Time.Now
				in.Old.Keys[0].LastCompositeDeadline = in.Time.Now
				want = false
			case "bytes":
				in.Batch.MetadataEnvelope = alternateMetadata(t, in.Batch.MetadataEnvelope)
				want = false
			case "history":
				f := factsFor(t, 1)
				in.History = []KeyHistoryObservation{deadlineObservation(f, in)}
				want = false
			case "zero candidate":
				c = &Candidate{}
				want = false
			case "cleared":
				c.Clear()
				want = false
			case "ownership":
				s := c.State()
				s.Roots[0].Root.Keys[0].PublicKey[0] ^= 1
				s.Keys[0].Material.PublicKey[0] ^= 1
				s.Clear()
			}
			e = RecheckModel(context.Background(), c, in)
			if (e == nil) != want {
				t.Fatalf("recheck want success %v, got %v", want, e)
			}
		})
	}
	var nilCandidate *Candidate
	if nilCandidate.State() != nil || nilCandidate.NoChange() || !nilCandidate.ValidUntil().IsZero() {
		t.Fatal("nil candidate")
	}
	nilCandidate.Clear()
}
