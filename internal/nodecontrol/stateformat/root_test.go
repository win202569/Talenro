package stateformat

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
)

// Distinct test keys, no cryptographic signing/trust claims.
func rootKeys(n int) []any {
	keys := make([]any, n)
	for i := range keys {
		var key [32]byte
		key[0] = byte(i + 1)
		d := sha256.Sum256(key[:])
		keys[i] = map[string]any{"algorithm": "ed25519", "key_id": hex.EncodeToString(d[:]), "public_key": base64.RawURLEncoding.EncodeToString(key[:])}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	return keys
}
func TestRootPayload(t *testing.T) {
	b, e := parseRootPayload(fixture(t, "root-initial.payload"))
	if e != nil || b.Kind != InitialRoot || b.Epoch != 19 || b.Sequence != 3 || b.RootVersion != 1 || b.Generation != 0 {
		t.Fatal(b, e)
	}
	assertClosedObjects(t, object(t, fixture(t, "root-initial.payload")), parseRootPayload)
	for _, n := range []int{0, 1, 5, 6} {
		for _, threshold := range []int{0, 1, n, n + 1} {
			m := object(t, fixture(t, "root-initial.payload"))
			m["root_version"] = "2"
			m["keys"] = rootKeys(n)
			m["threshold"] = fmt.Sprint(threshold)
			_, e := parseRootPayload(encoded(t, m))
			want := n >= 1 && n <= 5 && threshold >= 1 && threshold <= n
			if (e == nil) != want {
				t.Errorf("keys %d threshold %d: %v", n, threshold, e)
			}
		}
	}
	for _, mut := range []func(map[string]any){
		func(m map[string]any) { m["keys"] = rootKeys(2) }, func(m map[string]any) { m["threshold"] = "2" },
		func(m map[string]any) {
			k := m["keys"].([]any)[0].(map[string]any)
			k["key_id"] = fmt.Sprintf("%064x", 2)
		},
		func(m map[string]any) { k := m["keys"].([]any)[0].(map[string]any); k["algorithm"] = "rsa" },
		func(m map[string]any) { m["root_version"] = "2"; k := rootKeys(1); m["keys"] = []any{k[0], k[0]} },
		func(m map[string]any) {
			m["root_version"] = "2"
			k := rootKeys(2)
			k[0], k[1] = k[1], k[0]
			m["keys"] = k
		},
	} {
		m := object(t, fixture(t, "root-initial.payload"))
		mut(m)
		if _, e := parseRootPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Fatal(m, e)
		}
	}
}
func TestRotationPayload(t *testing.T) {
	b, e := parseRotationPayload(fixture(t, "root-rotation.payload"))
	if e != nil || b.Kind != RootRotation || b.PreviousRootVersion != 1 || b.RootVersion != 2 || b.Epoch != 19 || b.Sequence != 4 {
		t.Fatal(b, e)
	}
	assertClosedObjects(t, object(t, fixture(t, "root-rotation.payload")), parseRotationPayload)
	for _, mut := range []func(map[string]any){
		func(m map[string]any) { m["previous_root_version"] = "0" }, func(m map[string]any) { m["previous_root_version"] = "9223372036854775807" },
		func(m map[string]any) { m["new_root_set"].(map[string]any)["root_version"] = "3" },
		func(m map[string]any) { m["new_root_set"].(map[string]any)["control_plane_authority_epoch"] = "20" },
		func(m map[string]any) { m["new_root_set"].(map[string]any)["authority_sequence"] = "5" },
		func(m map[string]any) { m["valid_until"] = "2026-10-16T00:00:01Z" }, func(m map[string]any) { m["valid_until"] = m["issued_at"] },
	} {
		m := object(t, fixture(t, "root-rotation.payload"))
		mut(m)
		if _, e := parseRotationPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Fatal(m, e)
		}
	}
	if _, e := parseRotationPayload(fixture(t, "root-initial.payload")); e != ErrInvalidFormat {
		t.Fatal("accepted root without rotation", e)
	}
}
