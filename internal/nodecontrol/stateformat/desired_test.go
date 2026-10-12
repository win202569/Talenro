package stateformat

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func object(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if e := json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	return m
}

// Standard JSON serialization is only a test mutation transport for ASCII fixtures,
// never the oracle for production canonical bytes or golden vectors.
func encoded(t *testing.T, m any) []byte {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func processFixture(t *testing.T) map[string]any {
	return object(t, []byte(`{"adapter":"fixture","capacity_limits":{"connection_limit":"1","cpu_limit_basis_points":"1","cpu_quota_millicores":"100","egress_limit_bps":"1000000","file_descriptor_limit":"64","handshake_limit_per_second":"1","memory_limit_bytes":"67108864","packet_loss_limit_basis_points":"1","queue_limit":"1","task_limit":"32"},"capacity_profile_id":"capacity.v1","capacity_profile_version":"1","lifecycle":"running","probe_profile":"probe.v1","release_id":"release.v1","required":false,"required_metrics":["bytes"],"restart_profile":"restart.v1","slot_id":"slot-a","test_profile":"test.v1"}`))
}
func desiredObject(t *testing.T) map[string]any {
	m := object(t, fixture(t, "desired.payload"))
	m["processes"] = []any{processFixture(t)}
	return m
}

// Required/null/type checks must apply at every object depth, including false.
func assertClosedObjects(t *testing.T, m map[string]any, parse func([]byte) (artifactBindings, error)) {
	t.Helper()
	var visit func(map[string]any, string)
	visit = func(o map[string]any, path string) {
		for k, v := range o {
			delete(o, k)
			if _, e := parse(encoded(t, m)); e != ErrInvalidFormat {
				t.Errorf("missing %s/%s", path, k)
			}
			o[k] = v
			if v != nil {
				o[k] = nil
				if _, e := parse(encoded(t, m)); e != ErrInvalidFormat {
					t.Errorf("null %s/%s", path, k)
				}
				o[k] = v
			}
			o[k] = float64(2)
			if _, e := parse(encoded(t, m)); e != ErrInvalidFormat {
				t.Errorf("number %s/%s", path, k)
			}
			o[k] = v
			switch x := v.(type) {
			case map[string]any:
				visit(x, path+"/"+k)
			case []any:
				for i, item := range x {
					if child, ok := item.(map[string]any); ok {
						visit(child, fmt.Sprintf("%s/%s/%d", path, k, i))
					}
				}
			}
		}
		o["unknown_field"] = false
		if _, e := parse(encoded(t, m)); e != ErrInvalidFormat {
			t.Errorf("extra %s", path)
		}
		delete(o, "unknown_field")
	}
	visit(m, "")
}

func TestDesiredPayload(t *testing.T) {
	b, e := parseDesiredPayload(fixture(t, "desired.payload"))
	if e != nil || b.Kind != Desired || b.Epoch != 19 || b.Sequence != 9007199254740993 || b.Generation != 7 || b.InventoryVersion != 8 || b.ResourceEnvelopeVersion != 9 || b.NodeID.String() != "12345678-1234-4234-8234-123456789abc" || b.RootVersion != 0 || b.MetadataVersion != 0 || b.ResourceEnvelopeDigest[0] != 0xab {
		t.Fatal(b, e)
	}
	m := desiredObject(t)
	if _, e := parseDesiredPayload(encoded(t, m)); e != nil {
		t.Fatal("required false", e)
	}
	assertClosedObjects(t, m, parseDesiredPayload)
	for _, n := range []int{0, 8, 9} {
		m := desiredObject(t)
		p := make([]any, n)
		for i := range p {
			x := processFixture(t)
			x["slot_id"] = fmt.Sprintf("slot-%02d", i)
			p[i] = x
		}
		m["processes"] = p
		_, e := parseDesiredPayload(encoded(t, m))
		if (e == nil) != (n <= 8) {
			t.Errorf("process count %d: %v", n, e)
		}
	}
	for _, n := range []int{0, 1, 64, 65} {
		m := desiredObject(t)
		p := m["processes"].([]any)[0].(map[string]any)
		metrics := make([]string, n)
		for i := range metrics {
			metrics[i] = fmt.Sprintf("metric_%02d", i)
		}
		p["required_metrics"] = metrics
		_, e := parseDesiredPayload(encoded(t, m))
		if (e == nil) != (n >= 1 && n <= 64) {
			t.Errorf("metrics %d: %v", n, e)
		}
	}
	for _, reason := range []string{"initial", "operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize"} {
		m := desiredObject(t)
		m["reason"] = reason
		if _, e := parseDesiredPayload(encoded(t, m)); e != nil {
			t.Error(reason, e)
		}
	}
	for _, tc := range []struct {
		field string
		value any
	}{
		{"adapter", "shell"}, {"lifecycle", "paused"}, {"slot_id", "bad/id"}, {"release_id", strings.Repeat("x", 129)}, {"test_profile", ""}, {"capacity_profile_id", "-bad"}, {"probe_profile", "x y"}, {"restart_profile", "é"}, {"capacity_profile_version", "0"}, {"required_metrics", []string{"a", "a"}}, {"required_metrics", []string{"z", "a"}}, {"required_metrics", []string{"Upper"}}, {"required_metrics", []string{strings.Repeat("x", 65)}},
	} {
		m := desiredObject(t)
		m["processes"].([]any)[0].(map[string]any)[tc.field] = tc.value
		if _, e := parseDesiredPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Error(tc.field, tc.value, e)
		}
	}
	for _, field := range []string{"reason", "minimum_agent_version", "minimum_supervisor_version", "node_id", "node_resource_envelope_digest", "schema_version"} {
		m := desiredObject(t)
		m[field] = "invalid"
		if _, e := parseDesiredPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Error(field, e)
		}
	}
	for _, slots := range [][]string{{"a", "a"}, {"z", "a"}} {
		m := desiredObject(t)
		p := []any{processFixture(t), processFixture(t)}
		for i := range p {
			p[i].(map[string]any)["slot_id"] = slots[i]
		}
		m["processes"] = p
		if _, e := parseDesiredPayload(encoded(t, m)); e != ErrInvalidFormat {
			t.Error(slots, e)
		}
	}
}

func TestDesiredCapacityBounds(t *testing.T) {
	for _, tc := range []struct {
		field    string
		min, max uint64
	}{
		{"egress_limit_bps", 1000000, 1000000000000}, {"connection_limit", 1, 10000000}, {"handshake_limit_per_second", 1, 1000000}, {"cpu_quota_millicores", 100, 64000}, {"cpu_limit_basis_points", 1, 10000}, {"memory_limit_bytes", 67108864, 1099511627776}, {"task_limit", 32, 4096}, {"file_descriptor_limit", 64, 1000000}, {"queue_limit", 1, 1000000}, {"packet_loss_limit_basis_points", 1, 10000},
	} {
		for _, v := range []uint64{tc.min - 1, tc.min, tc.max, tc.max + 1} {
			m := desiredObject(t)
			m["processes"].([]any)[0].(map[string]any)["capacity_limits"].(map[string]any)[tc.field] = fmt.Sprint(v)
			_, e := parseDesiredPayload(encoded(t, m))
			if (e == nil) != (v >= tc.min && v <= tc.max) {
				t.Errorf("%s=%d: %v", tc.field, v, e)
			}
		}
	}
}

func TestDesiredWindow(t *testing.T) {
	for _, tc := range []struct {
		issued, until string
		ok            bool
	}{
		{"2026-10-09T00:00:00Z", "2026-10-10T00:00:00Z", true}, {"2026-10-09T00:00:00Z", "2026-10-10T00:00:01Z", false}, {"2026-10-09T00:00:00Z", "2026-10-09T00:00:00Z", false}, {"2026-10-09T00:00:00Z", "2026-10-08T23:59:59Z", false}, {"9999-12-31T23:59:58Z", "9999-12-31T23:59:59Z", true},
	} {
		m := desiredObject(t)
		m["issued_at"] = tc.issued
		m["valid_until"] = tc.until
		_, e := parseDesiredPayload(encoded(t, m))
		if (e == nil) != tc.ok {
			t.Errorf("%+v: %v", tc, e)
		}
	}
}
