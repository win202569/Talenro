package nodeissuer

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	"testing"
	"time"
)

// Catches dropping/reordering obligations or accepting an incomplete lower proof.
func TestIssuerLowerProofs(t *testing.T) {
	s, e := evaluateLower(context.Background(), fixtureInput(t))
	if e != nil {
		t.Fatal(e)
	}
	if e = checkLowerProofs(s); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*lowerSummaries){func(s *lowerSummaries) { s.ca.Proof = nodeca.None }, func(s *lowerSummaries) { s.leaf.Proof = nodeleafkeys.None }, func(s *lowerSummaries) { s.leaf.Leaf.Proof = nodeleaf.None }} {
		bad := s
		mutate(&bad)
		if checkLowerProofs(bad) != ErrInvalidMaterial {
			t.Fatal("partial proof")
		}
	}
	for i := 0; i < 6; i++ {
		bad := s
		bad.ca.Pending[i] = nodeca.UnknownObligation
		if checkLowerProofs(bad) != ErrInvalidMaterial {
			t.Fatal("CA obligation", i)
		}
		bad = s
		bad.leaf.Leaf.Pending[i] = nodeleaf.UnknownObligation
		if checkLowerProofs(bad) != ErrInvalidMaterial {
			t.Fatal("leaf obligation", i)
		}
	}
	bad := s
	bad.ca.Pending[0], bad.ca.Pending[1] = bad.ca.Pending[1], bad.ca.Pending[0]
	if checkLowerProofs(bad) != ErrInvalidMaterial {
		t.Fatal("CA order")
	}
	bad = s
	bad.leaf.Leaf.Pending[0], bad.leaf.Leaf.Pending[1] = bad.leaf.Leaf.Pending[1], bad.leaf.Leaf.Pending[0]
	if checkLowerProofs(bad) != ErrInvalidMaterial {
		t.Fatal("leaf order")
	}
}

// Catches remapping, wrapping, or executing leaf before a failing CA.
func TestIssuerLowerErrorIdentity(t *testing.T) {
	badCA := fixtureInput(t)
	badCA.CA.DER = []byte{0x30, 0}
	badCA.Leaf.DER = []byte{0x30, 0}
	badLeaf := fixtureInput(t)
	badLeaf.Leaf.DER = []byte{0x30, 0}
	badID := fixtureSetKeyID(t, fixtureInput(t), "leafAKI", make([]byte, 32))
	badCAKey := fixtureInput(t)
	badCAKey.CA.DeclaredSPKI = badCAKey.Leaf.ExpectedSPKI
	for _, c := range []struct {
		in   Input
		want error
	}{{badCA, nodeca.ErrInvalidMaterial}, {badLeaf, nodeleaf.ErrInvalidMaterial}, {badID, nodeleafkeys.ErrBindingMismatch}, {badCAKey, nodeca.ErrBindingMismatch}} {
		got, e := evaluateLower(context.Background(), c.in)
		if e != c.want || got != (lowerSummaries{}) {
			t.Fatalf("error identity: %v want %v", e, c.want)
		}
	}
}

// Catches an observed cancellation being masked by a stage result.
func TestIssuerStageError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	for _, ordinary := range []error{nil, ErrInvalidMaterial, nodeca.ErrInvalidMaterial, nodeleaf.ErrInvalidMaterial, ErrIssuerRelationMismatch} {
		if stageError(ctx, ordinary) != context.Canceled {
			t.Fatal("cancellation")
		}
		if stageError(expired, ordinary) != context.DeadlineExceeded {
			t.Fatal("deadline")
		}
		if stageError(context.Background(), ordinary) != ordinary {
			t.Fatal("original identity")
		}
	}
	if stageError(nil, nil) != ErrInvalidInput {
		t.Fatal("nil context")
	}
	if s, e := evaluateLower(ctx, fixtureInput(t)); e != context.Canceled || s != (lowerSummaries{}) {
		t.Fatal("lower cancellation", e)
	}
}
