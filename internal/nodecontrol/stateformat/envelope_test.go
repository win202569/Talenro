package stateformat

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func rewrap(t *testing.T, name string, p []byte) []byte {
	m := object(t, fixture(t, name+".envelope"))
	d := sha256.Sum256(p)
	m["payload_b64url"] = base64.RawURLEncoding.EncodeToString(p)
	m["payload_sha256"] = fmt.Sprintf("%x", d)
	return encoded(t, m)
}
func TestEnvelopeExactFields(t *testing.T) {
	for name, kind := range formatKinds {
		m := object(t, fixture(t, name+".envelope"))
		parse := func(b []byte) (artifactBindings, error) {
			a, e := ParseEnvelope(kind, b)
			if e != nil {
				if a != nil {
					t.Fatal("partial result")
				}
				return artifactBindings{}, e
			}
			defer a.Clear()
			return a.Bindings(), nil
		}
		if _, e := parse(encoded(t, m)); e != nil {
			t.Fatal(name, e)
		}
		assertClosedObjects(t, m, parse)
		for _, wrong := range []Kind{Invalid, 255} {
			if a, e := ParseEnvelope(wrong, encoded(t, m)); a != nil || e != ErrInvalidFormat {
				t.Fatal("unknown kind")
			}
		}
	}
}
func TestEnvelopeBindings(t *testing.T) {
	for name, kind := range formatKinds {
		for _, field := range []string{"schema_version", "payload_sha256", "control_plane_authority_epoch", "authority_sequence", "root_version", "previous_root_version"} {
			m := object(t, fixture(t, name+".envelope"))
			if _, ok := m[field]; !ok {
				continue
			}
			m[field] = "20"
			if field == "payload_sha256" {
				m[field] = strings.Repeat("ab", 32)
			}
			if a, e := ParseEnvelope(kind, encoded(t, m)); a != nil || e != ErrInvalidFormat {
				t.Error(name, field, e)
			}
		}
		for other := range formatKinds {
			if other == name {
				continue
			}
			if a, e := ParseEnvelope(kind, rewrap(t, name, fixture(t, other+".payload"))); a != nil || e != ErrInvalidFormat {
				t.Error("mixed payload", name, other, e)
			}
		}
	}
	m := object(t, fixture(t, "root-initial.payload"))
	m["root_version"] = "2"
	outer := object(t, rewrap(t, "root-initial", encoded(t, m)))
	outer["root_version"] = "2"
	if a, e := ParseEnvelope(InitialRoot, encoded(t, outer)); a != nil || e != ErrInvalidFormat {
		t.Fatal("noninitial root")
	}
}
func shareFixture(n int) []any {
	s := make([]any, n)
	for i := range s {
		s[i] = map[string]any{"key_id": fmt.Sprintf("%064x", i+1), "signature": "Ag" + strings.Repeat("A", 84)}
	}
	return s
}
func TestSharesFormat(t *testing.T) {
	for _, tc := range []struct {
		name, field string
		kind        Kind
	}{{"root-rotation", "current_signatures", RootRotation}, {"root-rotation", "new_signatures", RootRotation}, {"metadata", "signatures", Metadata}} {
		for _, n := range []int{0, 64, 65} {
			m := object(t, fixture(t, tc.name+".envelope"))
			m[tc.field] = shareFixture(n)
			a, e := ParseEnvelope(tc.kind, encoded(t, m))
			if (e == nil) != (n <= 64) {
				t.Error(tc, n, e)
			}
			if a != nil {
				a.Clear()
			}
		}
		for _, mut := range []func([]any){func(s []any) { s[1] = s[0] }, func(s []any) { s[0], s[1] = s[1], s[0] }, func(s []any) { s[0].(map[string]any)["signature"] = strings.Repeat("A", 86) }, func(s []any) { delete(s[0].(map[string]any), "key_id") }, func(s []any) { s[0].(map[string]any)["other"] = true }} {
			m := object(t, fixture(t, tc.name+".envelope"))
			s := shareFixture(2)
			mut(s)
			m[tc.field] = s
			if a, e := ParseEnvelope(tc.kind, encoded(t, m)); a != nil || e != ErrInvalidFormat {
				t.Error(tc, e)
			}
		}
	}
}
func TestEnvelopeSizeLayers(t *testing.T) {
	for _, n := range []int{65536, 65537} {
		raw := []byte(strings.Repeat("a", n))
		s := base64.RawURLEncoding.EncodeToString(raw)
		b, e := decodePayload(s)
		if (e == nil) != (n == 65536) {
			t.Error(n, len(s), e)
		}
		if e == nil && len(b) != n {
			t.Fatal("decoded length")
		}
	}
	for _, s := range []string{"", strings.Repeat("A", 87383), "YB", "YQ=", "YQ\n"} {
		if b, e := decodePayload(s); b != nil || e != ErrInvalidFormat {
			t.Errorf("bad payload encoding length %d", len(s))
		}
	}
	m := metadataObject(t)
	ids := make([]string, 800)
	for i := range ids {
		ids[i] = fmt.Sprintf("%064x", i+100)
	}
	m["revoked_key_ids"] = ids
	p := encoded(t, m)
	e := rewrap(t, "metadata", p)
	if len(p) > 65536 || len(e) <= 65536 {
		t.Fatal("layer fixture bounds")
	}
	a, err := ParseEnvelope(Metadata, e)
	if err != nil {
		t.Fatal(err)
	}
	a.Clear()
	if a, e := ParseEnvelope(Metadata, []byte(strings.Repeat(" ", 262145))); a != nil || e != ErrInvalidFormat {
		t.Fatal("outer size")
	}
}
