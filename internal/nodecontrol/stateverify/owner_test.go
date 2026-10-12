package stateverify

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

func TestOwnerExitCleanup(t *testing.T) {
	for _, exit := range []string{"error", "cancel", "panic"} {
		t.Run(exit, func(t *testing.T) {
			o := &verificationOwner{}
			a, e := stateformat.ParseEnvelope(stateformat.Metadata, frozenFile(t, "metadata.envelope.json"))
			if e != nil {
				t.Fatal(e)
			}
			o.artifacts[0] = a
			o.metadata, e = a.MetadataMaterial()
			if e != nil {
				t.Fatal(e)
			}
			o.roots[0] = stateformat.RootMaterial{Keys: []stateformat.PublicKeyMaterial{{PublicKey: [32]byte{1}}}}
			o.shares[0] = stateformat.SignatureMaterial{Metadata: []stateformat.SignatureShare{{Signature: [64]byte{1}}}}
			o.transcripts[0] = []byte{1, 2, 3}
			o.evidence = &SignatureEvidence{rotations: []ArtifactIdentity{{}}}
			o.evidence.rotations[0].PayloadDigest[0] = 1
			root, keys, shares, msg, ev := o.roots[0].Keys, o.metadata.Keys, o.shares[0].Metadata, o.transcripts[0], o.evidence.rotations
			func() {
				defer func() {
					v := recover()
					if (v != nil) != (exit == "panic") {
						t.Fatal("panic propagation")
					}
				}()
				defer o.Clear()
				if exit == "panic" {
					panic("test-owned exit")
				}
			}()
			if root[0] != (stateformat.PublicKeyMaterial{}) || keys[0] != (stateformat.MetadataKeyMaterial{}) || shares[0] != (stateformat.SignatureShare{}) || msg[0] != 0 || msg[1] != 0 || msg[2] != 0 || ev[0] != (ArtifactIdentity{}) || a.Bindings() != (stateformat.Bindings{}) {
				t.Fatal("owned exit cleanup")
			}
			o.Clear()
		})
	}
}

type countingContext struct {
	context.Context
	calls, stop int
}

func (c *countingContext) Err() error {
	c.calls++
	if c.stop > 0 && c.calls >= c.stop {
		return context.Canceled
	}
	return nil
}
func TestVerifyLateFailureAndCancel(t *testing.T) {
	b := signedBatch(t, 1)
	bad := jsonObject(t, b.DesiredEnvelope)
	bad["signature"] = "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	b.DesiredEnvelope = jsonBytes(t, bad)
	assertBatch(t, b, ErrSignatureMismatch)
	if e, err := Verify(nil, Batch{}); e != nil || err != ErrInvalidInput {
		t.Fatal("nil context", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e, err := Verify(ctx, Batch{}); e != nil || err != ErrCanceled {
		t.Fatal("cancel priority", err)
	}
	b = signedBatch(t, 1)
	counter := &countingContext{Context: context.Background()}
	e, err := Verify(counter, b)
	if err != nil {
		t.Fatal(err)
	}
	e.Clear()
	calls := counter.calls
	for stop := 1; stop <= calls; stop++ {
		c := &countingContext{Context: context.Background(), stop: stop}
		e, err := Verify(c, b)
		if e != nil || err != ErrCanceled {
			t.Fatalf("cancel checkpoint %d/%d: %v", stop, calls, err)
		}
	}
}
