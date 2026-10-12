package stateverify

import (
	"context"
	"fmt"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

// Skipping format validation for a different active artifact must fail this test.
func TestVerifyActiveRootErrorClass(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		for _, mode := range []string{"identical", "malformed", "noncanonical", "bad-digest", "wrong-kind", "valid-different"} {
			t.Run(fmt.Sprintf("%d/%s", n, mode), func(t *testing.T) {
				b := signedBatch(t, n)
				want := error(nil)
				switch mode {
				case "malformed":
					b.ActiveRootEnvelope = []byte("{")
					want = ErrInvalidInput
				case "noncanonical":
					b.ActiveRootEnvelope = append(append([]byte(nil), b.ActiveRootEnvelope...), ' ')
					want = ErrInvalidInput
				case "bad-digest":
					e := jsonObject(t, b.ActiveRootEnvelope)
					e["payload_sha256"] = "abababababababababababababababababababababababababababababababab"
					b.ActiveRootEnvelope = jsonBytes(t, e)
					want = ErrInvalidInput
				case "wrong-kind":
					b.ActiveRootEnvelope = b.MetadataEnvelope
					want = ErrInvalidInput
				case "valid-different":
					name := "anchor"
					if n > 0 {
						name = "rotation"
					}
					b.ActiveRootEnvelope = changePayload(t, b.ActiveRootEnvelope, name, func(p map[string]any) {
						p["authority_sequence"] = "100"
						if n > 0 {
							p["new_root_set"].(map[string]any)["authority_sequence"] = "100"
						}
					})
					want = ErrBindingMismatch
				}
				assertBatch(t, b, want)
			})
		}
	}
}

func TestVerifyActiveRootMismatchCancellation(t *testing.T) {
	b := signedBatch(t, 1)
	b.ActiveRootEnvelope = changePayload(t, b.ActiveRootEnvelope, "rotation", func(p map[string]any) {
		p["authority_sequence"] = "100"
		p["new_root_set"].(map[string]any)["authority_sequence"] = "100"
	})
	c := &countingContext{Context: context.Background()}
	if e, err := Verify(c, b); e != nil || err != ErrBindingMismatch {
		t.Fatal("valid mismatch", err)
	}
	for stop := 1; stop <= c.calls; stop++ {
		ctx := &countingContext{Context: context.Background(), stop: stop}
		if e, err := Verify(ctx, b); e != nil || err != ErrCanceled {
			t.Fatalf("cancel checkpoint %d: %v", stop, err)
		}
	}
}

func TestActiveRootOwnerSlot(t *testing.T) {
	o := &verificationOwner{}
	ctx := &countingContext{Context: context.Background(), stop: 2}
	a, e := o.parse(ctx, 35, stateformat.InitialRoot, frozenFile(t, "anchor.envelope.json"))
	if a != nil || e != ErrCanceled || o.artifacts[35] == nil {
		t.Fatal("cancel ownership", e)
	}
	held := o.artifacts[35]
	o.Clear()
	if held.Bindings() != (stateformat.Bindings{}) || held.Payload() != nil {
		t.Fatal("active artifact retained")
	}
}
