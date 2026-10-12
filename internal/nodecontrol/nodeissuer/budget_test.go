package nodeissuer

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strings"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	"testing"
	"time"
)

func contractInput() Input {
	return Input{Leaf: nodeleaf.Input{DER: []byte{1}, ExpectedSPKI: []byte{2}, NodeID: uuid.UUID{1, 0, 0, 0, 0, 0, 4, 0, 0x80, 0, 0, 0, 0, 0, 0, 1}, TrustDomain: "node.test"}, CA: nodeca.Input{DER: []byte{3}, DeclaredSPKI: []byte{4}}}
}

// Catches late budget checks, relaxed declarations, and context error mapping.
func TestIssuerPreflight(t *testing.T) {
	valid := contractInput()
	if err := preflight(context.Background(), valid); err != nil {
		t.Fatal(err)
	}
	if err := preflight(nil, valid); err != ErrInvalidInput {
		t.Fatal("nil context", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if err := preflight(canceled, valid); err != context.Canceled {
		t.Fatal(err)
	}
	if err := preflight(expired, valid); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
	setters := []func(*Input, []byte){func(in *Input, b []byte) { in.Leaf.DER = b }, func(in *Input, b []byte) { in.Leaf.ExpectedSPKI = b }, func(in *Input, b []byte) { in.CA.DER = b }, func(in *Input, b []byte) { in.CA.DeclaredSPKI = b }}
	for i, limit := range []int{65536, 512, 65536, 8192} {
		for _, n := range []int{0, 1, limit, limit + 1} {
			in := contractInput()
			setters[i](&in, make([]byte, n))
			want := error(nil)
			if n == 0 {
				want = ErrInvalidInput
			}
			if n > limit {
				want = ErrLimitExceeded
			}
			if got := preflight(context.Background(), in); got != want {
				t.Errorf("field %d size %d: %v want %v", i, n, got, want)
			}
		}
		in := Input{}
		setters[i](&in, make([]byte, limit+1))
		if err := preflight(context.Background(), in); err != ErrLimitExceeded {
			t.Fatal("budget before empty", err)
		}
		if err := preflight(canceled, in); err != context.Canceled {
			t.Fatal("context before budget", err)
		}
	}
	for _, id := range []uuid.UUID{uuid.Nil, {1}, {1, 0, 0, 0, 0, 0, 0, 0, 0xc0}} {
		in := valid
		in.Leaf.NodeID = id
		if preflight(context.Background(), in) != ErrInvalidInput {
			t.Fatal("UUID accepted", id)
		}
	}
	maxDomain := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, domain := range []string{"a", "0", "a-b.example", strings.Repeat("a", 63), maxDomain} {
		in := valid
		in.Leaf.TrustDomain = domain
		if err := preflight(context.Background(), in); err != nil {
			t.Fatal("valid domain", domain, err)
		}
	}
	for _, domain := range []string{"", ".a", "a.", "a..b", "-a", "a-", "A", "a_b", "é", "a b", strings.Repeat("a", 64), maxDomain + "e"} {
		in := valid
		in.Leaf.TrustDomain = domain
		if preflight(context.Background(), in) != ErrInvalidInput {
			t.Fatal("invalid domain accepted", domain)
		}
	}
}

// Catches remapping or aliasing failures across the public error boundary.
func TestIssuerFixedErrors(t *testing.T) {
	cases := []struct {
		err  error
		text string
	}{
		{ErrInvalidInput, "node issuer: invalid input"}, {ErrLimitExceeded, "node issuer: limit exceeded"}, {ErrInvalidMaterial, "node issuer: invalid material"},
		{ErrUnsupportedSignatureProfile, "node issuer: unsupported leaf signature profile"}, {ErrIssuerRelationMismatch, "node issuer: issuer relation mismatch"}, {ErrInvalidSignature, "node issuer: invalid leaf signature"},
	}
	for i, c := range cases {
		if c.err.Error() != c.text {
			t.Fatal("fixed error", i)
		}
		for j, d := range cases {
			if errors.Is(c.err, d.err) != (i == j) {
				t.Fatal("error identity", i, j)
			}
		}
		for _, lower := range []error{nodeleaf.ErrInvalidInput, nodeleaf.ErrInvalidMaterial, nodeleafkeys.ErrInvalidMaterial, nodeca.ErrInvalidMaterial, nodeca.ErrLimitExceeded} {
			if errors.Is(c.err, lower) {
				t.Fatal("lower alias", i)
			}
		}
	}
}
