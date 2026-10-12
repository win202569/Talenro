package nodeleaf

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strings"
	"testing"
	"time"
)

func budgetInput() Input {
	return Input{DER: []byte{1}, ExpectedSPKI: []byte{2}, NodeID: uuid.MustParse("11111111-2222-4333-8444-555555555555"), TrustDomain: "node.test"}
}
func TestLeafBudget(t *testing.T) {
	for _, c := range []struct {
		der, spki int
		want      error
	}{{0, 1, ErrInvalidInput}, {1, 0, ErrInvalidInput}, {1, 1, nil}, {65536, 512, nil}, {65537, 1, ErrLimitExceeded}, {1, 513, ErrLimitExceeded}} {
		in := budgetInput()
		in.DER = make([]byte, c.der)
		in.ExpectedSPKI = make([]byte, c.spki)
		if err := preflight(context.Background(), in); !errors.Is(err, c.want) {
			t.Fatalf("sizes %d/%d: %v", c.der, c.spki, err)
		}
	}
	in := budgetInput()
	in.DER = make([]byte, 65537)
	in.NodeID = uuid.Nil
	if !errors.Is(preflight(context.Background(), in), ErrLimitExceeded) {
		t.Fatal("budget must precede scope")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(preflight(ctx, in), context.Canceled) {
		t.Fatal("context must precede budget")
	}
	deadline, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if !errors.Is(preflight(deadline, in), context.DeadlineExceeded) {
		t.Fatal("deadline")
	}
	if !errors.Is(preflight(nil, budgetInput()), ErrInvalidInput) {
		t.Fatal("nil context")
	}
}
func TestLeafScope(t *testing.T) {
	d253 := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, d := range []string{"a", "node.test", "a-b.test", strings.Repeat("a", 63), d253} {
		in := budgetInput()
		in.TrustDomain = d
		if err := preflight(context.Background(), in); err != nil {
			t.Fatal("valid domain", d, err)
		}
	}
	for _, d := range []string{"", "Node.test", " node.test", "node.test ", "node.test.", "-a.test", "a-.test", "a..test", "nöde.test", "a_b.test", strings.Repeat("a", 64), d253 + "d"} {
		in := budgetInput()
		in.TrustDomain = d
		if !errors.Is(preflight(context.Background(), in), ErrInvalidInput) {
			t.Fatal("invalid domain accepted", d)
		}
	}
	for _, id := range []uuid.UUID{uuid.Nil, {1, 2, 3}} {
		in := budgetInput()
		in.NodeID = id
		if !errors.Is(preflight(context.Background(), in), ErrInvalidInput) {
			t.Fatal("invalid UUID")
		}
	}
	in := budgetInput()
	if expectedURI(in) != "spiffe://node.test/node/11111111-2222-4333-8444-555555555555" {
		t.Fatal("URI normalization")
	}
	// Variant, not UUID version, is the declared condition.
	in.NodeID[6] = 0x12
	if err := preflight(context.Background(), in); err != nil {
		t.Fatal("new UUID version constraint", err)
	}
}
