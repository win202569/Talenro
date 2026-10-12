package stateformat

import (
	"fmt"
	"sort"
	"testing"
)

func metadataObject(t *testing.T) map[string]any { return object(t, fixture(t, "metadata.payload")) }
func metadataKeys(n int) []any {
	keys := rootKeys(n)
	for _, v := range keys {
		k := v.(map[string]any)
		k["status"] = "active"
		k["not_before"] = "2026-10-09T00:00:00Z"
		k["not_after"] = "2026-10-16T00:00:00Z"
		k["revoked_at"] = nil
		k["revocation_reason"] = nil
		k["incident_id"] = nil
	}
	return keys
}
func TestMetadataPayload(t *testing.T) {
	b, e := parseMetadataPayload(fixture(t, "metadata.payload"))
	if e != nil || b.Kind != Metadata || b.Epoch != 19 || b.Sequence != 5 || b.RootVersion != 2 || b.MetadataVersion != 6 || b.Generation != 0 {
		t.Fatal(b, e)
	}
	assertClosedObjects(t, metadataObject(t), parseMetadataPayload)
	for _, n := range []int{0, 1, 64, 65} {
		m := metadataObject(t)
		m["keys"] = metadataKeys(n)
		_, e := parseMetadataPayload(encoded(t, m))
		if (e == nil) != (n >= 1 && n <= 64) {
			t.Errorf("count %d: %v", n, e)
		}
	}
	for _, tc := range []struct {
		until, refresh string
		ok             bool
	}{
		{"2026-10-11T00:00:00Z", "2026-10-09T00:00:00Z", true},
		{"2026-10-10T23:59:59Z", "2026-10-09T00:00:00Z", false},
		{"2026-10-16T00:00:00Z", "2026-10-14T00:00:00Z", true},
		{"2026-10-16T00:00:01Z", "2026-10-14T00:00:00Z", false},
		{"2026-10-16T00:00:00Z", "2026-10-14T00:00:01Z", false},
		{"2026-10-16T00:00:00Z", "2026-10-08T23:59:59Z", false},
	} {
		m := metadataObject(t)
		m["valid_until"] = tc.until
		m["next_refresh_at"] = tc.refresh
		_, e := parseMetadataPayload(encoded(t, m))
		if (e == nil) != tc.ok {
			t.Errorf("%+v: %v", tc, e)
		}
	}
	for _, field := range []string{"role", "schema_version", "root_version", "metadata_version", "control_plane_authority_epoch", "authority_sequence"} {
		m := metadataObject(t)
		m[field] = "wrong"
		if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Error(field, e)
		}
	}
	for _, mut := range []func(map[string]any){
		func(m map[string]any) { m["keys"] = []any{metadataKeys(1)[0], metadataKeys(1)[0]} },
		func(m map[string]any) { k := metadataKeys(2); k[0], k[1] = k[1], k[0]; m["keys"] = k },
		func(m map[string]any) { m["keys"].([]any)[0].(map[string]any)["key_id"] = fmt.Sprintf("%064x", 99) },
		func(m map[string]any) { m["keys"].([]any)[0].(map[string]any)["algorithm"] = "rsa" },
	} {
		m := metadataObject(t)
		mut(m)
		if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Fatal(m, e)
		}
	}
}

func TestMetadataRevocationPresence(t *testing.T) {
	for _, status := range []string{"active", "retiring", "revoked"} {
		m := metadataObject(t)
		keys := metadataKeys(2)
		m["keys"] = keys
		k := keys[0].(map[string]any)
		k["status"] = status
		if status == "revoked" {
			k["revoked_at"] = "2026-10-09T00:00:00Z"
			k["revocation_reason"] = "scheduled"
			m["revoked_key_ids"] = []string{k["key_id"].(string)}
		}
		if _, e := parseMetadataPayload(encoded(t, m)); e != nil {
			t.Fatal(status, e)
		}
		for _, field := range []string{"revoked_at", "revocation_reason", "incident_id"} {
			v := k[field]
			delete(k, field)
			if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
				t.Error("absent", status, field, e)
			}
			k[field] = v
		}
		if status != "revoked" {
			for _, field := range []string{"revoked_at", "revocation_reason", "incident_id"} {
				k[field] = "2026-10-09T00:00:00Z"
				if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
					t.Error(status, field, e)
				}
				k[field] = nil
			}
			m["revoked_key_ids"] = []string{k["key_id"].(string)}
			if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
				t.Error("active ledger", e)
			}
			continue
		}
		for _, mut := range []func(map[string]any){
			func(k map[string]any) { k["revoked_at"] = nil }, func(k map[string]any) { k["revoked_at"] = "2026-10-09T00:00:01Z" }, func(k map[string]any) { k["revocation_reason"] = nil },
			func(k map[string]any) { k["revocation_reason"] = "other" }, func(k map[string]any) { k["incident_id"] = "12345678-1234-4234-8234-123456789abc" },
			func(k map[string]any) { k["revocation_reason"] = "security_incident" },
		} {
			copy := object(t, encoded(t, m))
			mut(copy["keys"].([]any)[0].(map[string]any))
			if _, e := parseMetadataPayload(encoded(t, copy)); e != ErrInvalidFormat {
				t.Error(copy, e)
			}
		}
		k["revocation_reason"] = "security_incident"
		k["incident_id"] = "12345678-1234-4234-8234-123456789abc"
		if _, e := parseMetadataPayload(encoded(t, m)); e != nil {
			t.Fatal(e)
		}
		m["revoked_key_ids"] = []string{}
		if _, e := parseMetadataPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Error("missing ledger", e)
		}
	}
}

// A one-second hole, retiring-only coverage, or key-order/time-order confusion
// must not turn incomplete active coverage into valid metadata.
func TestMetadataActiveCoverage(t *testing.T) {
	for _, tc := range []struct {
		name, aStart, aEnd, bStart, bEnd, status string
		ok                                       bool
	}{
		{"adjacent reverse time", "2026-10-12T00:00:00Z", "2026-10-16T00:00:00Z", "2026-10-09T00:00:00Z", "2026-10-12T00:00:00Z", "active", true},
		{"gap", "2026-10-09T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-12T00:00:01Z", "2026-10-16T00:00:00Z", "active", false},
		{"first gap", "2026-10-09T00:00:01Z", "2026-10-16T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-16T00:00:00Z", "active", false},
		{"last gap", "2026-10-09T00:00:00Z", "2026-10-15T23:59:59Z", "2026-10-10T00:00:00Z", "2026-10-15T23:59:59Z", "active", false},
		{"overlap", "2026-10-09T00:00:00Z", "2026-10-13T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-16T00:00:00Z", "active", true},
		{"contained", "2026-10-09T00:00:00Z", "2026-10-16T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-13T00:00:00Z", "active", true},
		{"retiring gap", "2026-10-09T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-12T00:00:00Z", "2026-10-16T00:00:00Z", "retiring", false},
		{"nonintersecting", "2026-10-09T00:00:00Z", "2026-10-16T00:00:00Z", "2026-10-16T00:00:00Z", "2026-10-17T00:00:00Z", "retiring", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := metadataObject(t)
			keys := metadataKeys(2)
			a, b := keys[0].(map[string]any), keys[1].(map[string]any)
			a["not_before"] = tc.aStart
			a["not_after"] = tc.aEnd
			b["not_before"] = tc.bStart
			b["not_after"] = tc.bEnd
			b["status"] = tc.status
			m["keys"] = keys
			body := encoded(t, m)
			_, e := parseMetadataPayload(body)
			if (e == nil) != tc.ok {
				t.Fatal(e)
			}
			if string(body) != string(encoded(t, m)) {
				t.Fatal("mutated input")
			}
		})
	}
}

func TestMetadataLedgerBudget(t *testing.T) {
	for _, n := range []int{0, 4096, 4097} {
		ids := make([]string, n)
		for i := range ids {
			ids[i] = fmt.Sprintf("%064x", i+100)
		}
		if validRevokedIDs(ids) != (n <= 4096) {
			t.Error(n)
		}
	}
	for _, ids := range [][]string{{"bad"}, {fmt.Sprintf("%064x", 2), fmt.Sprintf("%064x", 1)}, {fmt.Sprintf("%064x", 1), fmt.Sprintf("%064x", 1)}} {
		if validRevokedIDs(ids) {
			t.Error(ids)
		}
	}
	ids := make([]string, 1000)
	for i := range ids {
		ids[i] = fmt.Sprintf("%064x", i+100)
	}
	sort.Strings(ids)
	m := metadataObject(t)
	m["revoked_key_ids"] = ids
	body := encoded(t, m)
	if len(body) <= 65536 {
		t.Fatal("test must exceed payload budget")
	}
	before := string(body)
	if _, e := parseMetadataPayload(body); e != ErrInvalidFormat {
		t.Fatal(e)
	}
	if string(body) != before || len(ids) != 1000 {
		t.Fatal("truncated input")
	}
}
