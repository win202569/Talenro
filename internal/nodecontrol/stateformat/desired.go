package stateformat

import (
	"encoding/json"
	"time"
)

type desiredWire struct {
	Schema          string        `json:"schema_version"`
	Epoch           string        `json:"control_plane_authority_epoch"`
	Sequence        string        `json:"authority_sequence"`
	NodeID          string        `json:"node_id"`
	Generation      string        `json:"generation"`
	Inventory       string        `json:"inventory_version"`
	ResourceVersion string        `json:"node_resource_envelope_version"`
	ResourceDigest  string        `json:"node_resource_envelope_digest"`
	Issued          string        `json:"issued_at"`
	Until           string        `json:"valid_until"`
	Agent           string        `json:"minimum_agent_version"`
	Supervisor      string        `json:"minimum_supervisor_version"`
	Reason          string        `json:"reason"`
	Processes       []processWire `json:"processes"`
}
type processWire struct {
	Slot            string       `json:"slot_id"`
	Adapter         string       `json:"adapter"`
	Lifecycle       string       `json:"lifecycle"`
	Release         string       `json:"release_id"`
	Required        bool         `json:"required"`
	Test            string       `json:"test_profile"`
	CapacityProfile string       `json:"capacity_profile_id"`
	CapacityVersion string       `json:"capacity_profile_version"`
	Metrics         []string     `json:"required_metrics"`
	Capacity        capacityWire `json:"capacity_limits"`
	Probe           string       `json:"probe_profile"`
	Restart         string       `json:"restart_profile"`
}
type capacityWire struct {
	Egress      string `json:"egress_limit_bps"`
	Connections string `json:"connection_limit"`
	Handshakes  string `json:"handshake_limit_per_second"`
	CPUQuota    string `json:"cpu_quota_millicores"`
	CPULimit    string `json:"cpu_limit_basis_points"`
	Memory      string `json:"memory_limit_bytes"`
	Tasks       string `json:"task_limit"`
	FD          string `json:"file_descriptor_limit"`
	Queue       string `json:"queue_limit"`
	Loss        string `json:"packet_loss_limit_basis_points"`
}

var capacitySchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"egress_limit_bps":               wireString,
	"connection_limit":               wireString,
	"handshake_limit_per_second":     wireString,
	"cpu_quota_millicores":           wireString,
	"cpu_limit_basis_points":         wireString,
	"memory_limit_bytes":             wireString,
	"task_limit":                     wireString,
	"file_descriptor_limit":          wireString,
	"queue_limit":                    wireString,
	"packet_loss_limit_basis_points": wireString,
}}

var processSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"slot_id":                  wireString,
	"adapter":                  wireString,
	"lifecycle":                wireString,
	"release_id":               wireString,
	"required":                 &wireSchema{kind: 'b'},
	"test_profile":             wireString,
	"capacity_profile_id":      wireString,
	"capacity_profile_version": wireString,
	"required_metrics":         &wireSchema{kind: 'a', min: 1, max: 64, element: wireString},
	"capacity_limits":          capacitySchema,
	"probe_profile":            wireString,
	"restart_profile":          wireString,
}}

var desiredSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                 wireString,
	"control_plane_authority_epoch":  wireString,
	"authority_sequence":             wireString,
	"node_id":                        wireString,
	"generation":                     wireString,
	"inventory_version":              wireString,
	"node_resource_envelope_version": wireString,
	"node_resource_envelope_digest":  wireString,
	"issued_at":                      wireString,
	"valid_until":                    wireString,
	"minimum_agent_version":          wireString,
	"minimum_supervisor_version":     wireString,
	"reason":                         wireString,
	"processes":                      &wireSchema{kind: 'a', min: 0, max: 8, element: processSchema},
}}

func parseDesiredPayload(body []byte) (artifactBindings, error) {
	b, slots, e := decodeDesiredPayload(body)
	clear(slots)
	return b, e
}
func decodeDesiredPayload(body []byte) (artifactBindings, []string, error) {
	if strictCanonical(body, 65536, desiredSchema) != nil {
		return artifactBindings{}, nil, ErrInvalidFormat
	}
	var w desiredWire
	if json.Unmarshal(body, &w) != nil {
		return artifactBindings{}, nil, ErrInvalidFormat
	}
	c := formatCheck{ok: true}
	b := artifactBindings{Kind: Desired}
	b.Epoch = c.u(w.Epoch)
	b.Sequence = c.u(w.Sequence)
	b.Generation = c.u(w.Generation)
	b.InventoryVersion = c.u(w.Inventory)
	b.ResourceEnvelopeVersion = c.u(w.ResourceVersion)
	var err error
	b.NodeID, err = parseUUID(w.NodeID)
	c.require(err == nil)
	b.ResourceEnvelopeDigest, err = parseDigest(w.ResourceDigest)
	c.require(err == nil)
	b.IssuedAt = c.timestamp(w.Issued)
	b.ValidUntil = c.timestamp(w.Until)
	c.require(w.Schema == "node-desired-state.v1" && validBuild(w.Agent) && validBuild(w.Supervisor) && oneOf(w.Reason, "initial", "operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize"))
	c.require(validWindow(b.IssuedAt, b.ValidUntil, 24*time.Hour))
	previous := ""
	for _, p := range w.Processes {
		c.require(validID(p.Slot, 64) && p.Slot > previous)
		previous = p.Slot
		c.require(oneOf(p.Adapter, "fixture", "xray", "sing_box") && oneOf(p.Lifecycle, "running", "stopped", "draining"))
		for _, id := range []string{p.Release, p.Test, p.CapacityProfile, p.Probe, p.Restart} {
			c.require(validID(id, 128))
		}
		c.u(p.CapacityVersion)
		last := ""
		for _, metric := range p.Metrics {
			c.require(validMetric(metric) && metric > last)
			last = metric
		}
		c.bounded(p.Capacity.Egress, 1000000, 1000000000000)
		c.bounded(p.Capacity.Connections, 1, 10000000)
		c.bounded(p.Capacity.Handshakes, 1, 1000000)
		c.bounded(p.Capacity.CPUQuota, 100, 64000)
		c.bounded(p.Capacity.CPULimit, 1, 10000)
		c.bounded(p.Capacity.Memory, 67108864, 1099511627776)
		c.bounded(p.Capacity.Tasks, 32, 4096)
		c.bounded(p.Capacity.FD, 64, 1000000)
		c.bounded(p.Capacity.Queue, 1, 1000000)
		c.bounded(p.Capacity.Loss, 1, 10000)
	}
	if !c.ok {
		return artifactBindings{}, nil, ErrInvalidFormat
	}
	var slots []string
	for _, p := range w.Processes {
		slots = append(slots, p.Slot)
	}
	return b, slots, nil
}
func validMetric(s string) bool {
	if len(s) < 1 || len(s) > 64 || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !(s[i] >= 'a' && s[i] <= 'z' || s[i] >= '0' && s[i] <= '9' || s[i] == '_') {
			return false
		}
	}
	return true
}
