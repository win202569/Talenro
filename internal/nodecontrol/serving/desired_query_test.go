package serving

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/store"
)

// Capture the real generated query at its external transport boundary. These
// assertions are static contract checks, not PostgreSQL execution evidence.
type desiredQueryDB struct {
	store.DBTX
	query string
	args  []any
}

func (d *desiredQueryDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	d.query, d.args = sql, args
	return &adapterRows{}, nil
}
func captureDesiredQuery(t *testing.T) *desiredQueryDB {
	t.Helper()
	d := &desiredQueryDB{}
	node := uuid.MustParse("41000000-0000-4000-8000-000000000001")
	nodeScope, trustScope := make([]byte, 32), make([]byte, 32)
	nodeScope[0], trustScope[0] = 17, 29
	rows, err := store.New(d).ListExactDesiredProjection(context.Background(), store.ListExactDesiredProjectionParams{
		NodeID: node, AuthorityEpoch: 19, NodeScopeDigest: nodeScope, TrustScopeDigest: trustScope,
	})
	if err != nil || len(rows) != 0 {
		t.Fatalf("query result: %v", err)
	}
	// Parameters are introduced in the input CTE in this fixed order.
	want := []any{node, int64(19), nodeScope, trustScope}
	if !reflect.DeepEqual(d.args, want) {
		t.Fatalf("parameter binding: %#v", d.args)
	}
	return d
}
func TestDesiredQueryArgumentMapping(t *testing.T) { captureDesiredQuery(t) }

func desiredQueryMissingGuards(sql string) []string {
	q := strings.ToLower(strings.Join(strings.Fields(sql), " "))
	required := []string{
		", false)::boolean as pointer_groups_valid", ", false)::boolean as current_valid", ", false)::boolean as rotation_valid",
		"d.root_version = i.active_root_version", "d.metadata_version = i.active_metadata_version", "d.authority_epoch = a.authority_epoch",
		"r.publish_id = i.active_root_publish_id", "r.reserved_version = i.active_root_version",
		"m.publish_id = i.active_metadata_publish_id", "m.reserved_version = i.active_metadata_version",
		"r.authority_epoch = a.authority_epoch", "m.authority_epoch = a.authority_epoch", "h.authority_epoch = a.authority_epoch",
		"m.base_root_version = i.active_root_version", "s.signature_verified_at is not null",
		"from nodecontrol.node_inventory i", "i.node_id = a.node_id",
		"d.generation = i.active_desired_generation", "d.root_publish_id = i.active_root_publish_id",
		"d.metadata_publish_id = i.active_metadata_publish_id", "s.signing_id = d.signing_id",
		"s.canonical_payload = d.canonical_payload", "s.signature = d.signature",
		"s.payload_digest = d.payload_digest", "s.expected_key_id = d.signing_key_id",
		"s.captured_inventory_version = d.inventory_version", "s.signing_kind = 'desired'",
		"s.authority_operation_id = d.authority_operation_id", "s.authority_epoch = d.authority_epoch",
		"s.authority_sequence = d.authority_sequence", "s.reserved_generation = d.generation",
		"s.root_version = d.root_version", "s.metadata_version = d.metadata_version", "s.node_id = d.node_id",
		"df.effect_kind = 'desired_activate'", "df.scope_kind = 'node'",
		"rf.effect_kind = 'root_publish'", "mf.effect_kind = 'metadata_publish'",
		"hf.effect_kind = 'root_publish'", "hf.scope_kind = 'global_node_trust'",
		"h.reason = 'root_rotation'", "isfinite(s.signature_verified_at)",
		"generate_series(2::bigint, case when c.current_valid and c.active_root_version between 1 and 33 then c.active_root_version else 1::bigint end)",
		"left join nodecontrol.node_root_metadata_publish_intents h", "order by rotation_expected_version", "limit 33",
	}
	for _, p := range []string{"s", "r", "m", "h"} {
		required = append(required, p+".status = 'active'", p+".terminal_at is not null", "isfinite("+p+".terminal_at)", p+".failure_reason is null")
	}
	for _, fence := range []struct{ alias, owner, scope, digest string }{
		{"df", "d", "node", "node_scope_digest"},
		{"rf", "r", "global_node_trust", "trust_scope_digest"},
		{"mf", "m", "global_node_trust", "trust_scope_digest"},
		{"hf", "h", "global_node_trust", "trust_scope_digest"},
	} {
		a, p := fence.alias, fence.owner
		required = append(required, a+".provider_status = 'committed'", a+".visibility_state = 'active'",
			a+".operation_id is not null", a+".operation_id = "+p+".authority_operation_id",
			a+".authority_epoch = "+p+".authority_epoch", a+".authority_sequence = "+p+".authority_sequence",
			a+".scope_kind = '"+fence.scope+"'", a+".scope_digest = a."+fence.digest)
	}
	var missing []string
	for _, s := range required {
		if !strings.Contains(q, s) {
			missing = append(missing, s)
		}
	}
	for _, s := range []string{"max(", "now()", "current_timestamp", "offset "} {
		if strings.Contains(q, s) {
			missing = append(missing, "forbidden: "+s)
		}
	}
	return missing
}
func TestDesiredQueryGuardSensitivity(t *testing.T) {
	// Break caught: the emitted SQL silently stops requiring an exact active
	// fence/current artifact, while synthetic rows still claim CurrentValid.
	q := captureDesiredQuery(t).query
	fragments := []string{
		", false)::boolean AS pointer_groups_valid",
		", false)::boolean AS current_valid",
		", false)::boolean AS rotation_valid",
		"d.root_version = i.active_root_version",
		"d.metadata_version = i.active_metadata_version",
		"d.authority_epoch = a.authority_epoch",
		"r.publish_id = i.active_root_publish_id",
		"r.reserved_version = i.active_root_version",
		"m.publish_id = i.active_metadata_publish_id",
		"m.reserved_version = i.active_metadata_version",
		"r.authority_epoch = a.authority_epoch",
		"m.authority_epoch = a.authority_epoch",
		"h.authority_epoch = a.authority_epoch",
		"m.base_root_version = i.active_root_version",
		"s.failure_reason IS NULL",
		"s.status = 'active'",
		"s.terminal_at IS NOT NULL",
		"isfinite(s.terminal_at)",
		"s.signature_verified_at IS NOT NULL",
	}
	for _, p := range []string{"r", "m", "h"} {
		fragments = append(fragments, p+".status = 'active'", p+".terminal_at IS NOT NULL", "isfinite("+p+".terminal_at)", p+".failure_reason IS NULL")
	}
	for _, f := range []struct{ alias, owner, scope string }{
		{"df", "d", "node"}, {"rf", "r", "global_node_trust"}, {"mf", "m", "global_node_trust"}, {"hf", "h", "global_node_trust"},
	} {
		digest := "trust_scope_digest"
		if f.scope == "node" {
			digest = "node_scope_digest"
		}
		fragments = append(fragments,
			f.alias+".operation_id IS NOT NULL",
			f.alias+".operation_id = "+f.owner+".authority_operation_id",
			f.alias+".authority_epoch = "+f.owner+".authority_epoch",
			f.alias+".authority_sequence = "+f.owner+".authority_sequence",
			f.alias+".scope_kind = '"+f.scope+"'",
			f.alias+".scope_digest = a."+digest,
		)
	}
	for _, fragment := range fragments {
		t.Run(fragment, func(t *testing.T) {
			altered := strings.Replace(q, fragment, "REMOVED_GUARD", 1)
			if altered == q {
				t.Fatalf("fixture guard not present: %s", fragment)
			}
			if missing := desiredQueryMissingGuards(altered); len(missing) == 0 {
				t.Fatalf("removed guard not detected: %s", fragment)
			}
		})
	}
}

func TestDesiredQueryGuardShape(t *testing.T) {
	q := captureDesiredQuery(t).query
	if missing := desiredQueryMissingGuards(q); len(missing) != 0 {
		t.Fatalf("missing guards: %v", missing)
	}
	// Sensitivity only: removing either emitted guard must be detected.
	for _, s := range []string{"rf.effect_kind = 'root_publish'", "CASE WHEN c.current_valid AND c.active_root_version BETWEEN 1 AND 33"} {
		changed := strings.Replace(q, s, "REMOVED", 1)
		if changed == q || len(desiredQueryMissingGuards(changed)) == 0 {
			t.Fatalf("guard removal not detected: %s", s)
		}
	}
}
