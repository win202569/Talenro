package nodeleafkeys

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strings"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
)

func keysBudgetInput() Input {
	return Input{Leaf: nodeleaf.Input{DER: []byte{1}, ExpectedSPKI: []byte{2}, NodeID: uuid.MustParse("11111111-2222-4333-8444-555555555555"), TrustDomain: "node.test"}, DeclaredIssuerSPKI: []byte{3}}
}
func TestKeysBudget(t *testing.T) {
	for _, field := range []string{"der", "expected", "issuer"} {
		limit := map[string]int{"der": 65536, "expected": 512, "issuer": 8192}[field]
		for _, n := range []int{0, 1, limit, limit + 1} {
			in := keysBudgetInput()
			b := make([]byte, n)
			switch field {
			case "der":
				in.Leaf.DER = b
			case "expected":
				in.Leaf.ExpectedSPKI = b
			case "issuer":
				in.DeclaredIssuerSPKI = b
			}
			var want error
			if n == 0 {
				want = ErrInvalidInput
			}
			if n > limit {
				want = ErrLimitExceeded
			}
			if got := preflight(context.Background(), in); got != want {
				t.Fatalf("%s/%d: %v want %v", field, n, got, want)
			}
		}
	}
	in := Input{DeclaredIssuerSPKI: make([]byte, 8193)}
	if preflight(context.Background(), in) != ErrLimitExceeded {
		t.Fatal("limit before empty")
	}
	in = keysBudgetInput()
	in.Leaf.DER = make([]byte, 65537)
	in.Leaf.NodeID = uuid.Nil
	if preflight(context.Background(), in) != ErrLimitExceeded {
		t.Fatal("limit before scope")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if preflight(ctx, in) != context.Canceled {
		t.Fatal("context first")
	}
	if preflight(nil, in) != ErrInvalidInput {
		t.Fatal("nil first")
	}
}
func TestKeysScope(t *testing.T) {
	good253 := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, d := range []string{"a", "0", "a-b.node", good253, strings.Repeat("a", 63)} {
		in := keysBudgetInput()
		in.Leaf.TrustDomain = d
		if err := preflight(context.Background(), in); err != nil {
			t.Fatalf("good domain %q: %v", d, err)
		}
	}
	for _, d := range []string{"", good253 + "e", strings.Repeat("a", 64), "A.test", "é.test", " node.test", "node.test ", "node.test.", "-a", "a-", "a..b", "a_b", "a/b"} {
		in := keysBudgetInput()
		in.Leaf.TrustDomain = d
		if preflight(context.Background(), in) != ErrInvalidInput {
			t.Fatalf("bad domain %q", d)
		}
	}
	for _, variant := range []byte{0x00, 0x40, 0xc0, 0xe0} {
		in := keysBudgetInput()
		in.Leaf.NodeID[8] = variant
		if preflight(context.Background(), in) != ErrInvalidInput {
			t.Fatal("variant")
		}
	}
	in := keysBudgetInput()
	in.Leaf.NodeID = uuid.Nil
	if preflight(context.Background(), in) != ErrInvalidInput {
		t.Fatal("zero uuid")
	}
	in = keysBudgetInput()
	in.Leaf.NodeID[6] = 0 // no version policy
	if preflight(context.Background(), in) != nil {
		t.Fatal("unexpected UUID version policy")
	}
}
func TestKeysFixedErrors(t *testing.T) {
	got := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidMaterial, ErrUnsupportedLeafProfile, ErrBindingMismatch}
	want := []error{nodeleaf.ErrInvalidInput, nodeleaf.ErrLimitExceeded, nodeleaf.ErrInvalidMaterial, nodeleaf.ErrUnsupportedLeafProfile, nodeleaf.ErrBindingMismatch}
	for i := range got {
		if got[i] != want[i] || !errors.Is(got[i], want[i]) {
			t.Fatal("alias identity")
		}
	}
	if ErrUnsupportedKeyIdentifierProfile.Error() != "node leaf keys: unsupported key identifier profile" {
		t.Fatal("fixed message")
	}
}
