package stateverify

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"sort"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

func testPrivate(n byte) ed25519.PrivateKey {
	seed := make([]byte, 32)
	seed[0] = n
	return ed25519.NewKeyFromSeed(seed)
}
func thresholdFixture(n int, msg []byte) (stateformat.RootMaterial, []stateformat.SignatureShare) {
	r := stateformat.RootMaterial{Version: 2, Threshold: 1}
	s := []stateformat.SignatureShare{}
	for i := 0; i < n; i++ {
		k := testPrivate(byte(i + 1))
		pub := k.Public().(ed25519.PublicKey)
		id := sha256.Sum256(pub)
		r.Keys = append(r.Keys, stateformat.PublicKeyMaterial{KeyID: id, PublicKey: [32]byte(pub)})
		s = append(s, stateformat.SignatureShare{KeyID: id, Signature: [64]byte(ed25519.Sign(k, msg))})
	}
	sort.Slice(r.Keys, func(i, j int) bool { return bytes.Compare(r.Keys[i].KeyID[:], r.Keys[j].KeyID[:]) < 0 })
	sort.Slice(s, func(i, j int) bool { return bytes.Compare(s[i].KeyID[:], s[j].KeyID[:]) < 0 })
	return r, s
}
func TestThresholdMatrix(t *testing.T) {
	msg := []byte("TALENRO-NODE-STATE-METADATA-V1\x00{}")
	for n := 1; n <= 5; n++ {
		r, s := thresholdFixture(n, msg)
		for q := 1; q <= n; q++ {
			r.Threshold = uint64(q)
			for votes := 0; votes <= n; votes++ {
				want := error(nil)
				if votes < q {
					want = ErrSignatureMismatch
				}
				if e := verifyThreshold(context.Background(), r, s[:votes], msg); e != want {
					t.Fatalf("n=%d q=%d votes=%d got %v", n, q, votes, e)
				}
			}
		}
	}
	r, s := thresholdFixture(1, msg)
	unknown := stateformat.SignatureShare{KeyID: [32]byte{9}, Signature: [64]byte{1}}
	if e := verifyThreshold(context.Background(), r, []stateformat.SignatureShare{unknown}, msg); e != ErrSignatureMismatch {
		t.Fatal("unknown counted", e)
	}
	if e := verifyThreshold(context.Background(), r, append(s, unknown), msg); e != nil {
		t.Fatal("unknown must be ignored", e)
	}
	for _, wrong := range [][]byte{[]byte("wrong domain"), {1}} {
		_, bad := thresholdFixture(1, wrong)
		if e := verifyThreshold(context.Background(), r, bad, msg); e != ErrSignatureMismatch {
			t.Fatal("bad message", e)
		}
	}
	digest := sha256.Sum256(msg)
	_, bad := thresholdFixture(1, digest[:])
	if e := verifyThreshold(context.Background(), r, bad, msg); e != ErrSignatureMismatch {
		t.Fatal("hash message", e)
	}
	bad[0].Signature = [64]byte{1}
	if e := verifyThreshold(context.Background(), r, bad, msg); e != ErrSignatureMismatch {
		t.Fatal("fake signature", e)
	}
	bad[0].Signature = [64]byte(ed25519.Sign(testPrivate(99), msg))
	if e := verifyThreshold(context.Background(), r, bad, msg); e != ErrSignatureMismatch {
		t.Fatal("wrong key", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := verifyThreshold(ctx, r, s, msg); e != ErrCanceled {
		t.Fatal("cancel", e)
	}
}
func TestThresholdKnownBadAfterQuorum(t *testing.T) {
	msg := []byte("message")
	r, s := thresholdFixture(3, msg)
	r.Threshold = 1
	s[2].Signature[0] ^= 1
	if e := verifyThreshold(context.Background(), r, s, msg); e != ErrSignatureMismatch {
		t.Fatal("known bad after quorum", e)
	}
}
func TestThresholdSeparateGroups(t *testing.T) {
	msg := []byte("rotation")
	r, s := thresholdFixture(2, msg)
	r.Threshold = 2
	if e := verifyThreshold(context.Background(), r, s, msg); e != nil {
		t.Fatal(e)
	}
	if e := verifyThreshold(context.Background(), r, s, msg); e != nil {
		t.Fatal("same keys allowed each generation", e)
	}
	for _, part := range [][]stateformat.SignatureShare{s[:1], s[1:]} {
		if e := verifyThreshold(context.Background(), r, part, msg); e != ErrSignatureMismatch {
			t.Fatal("pooled groups", e)
		}
	}
}
