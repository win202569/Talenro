package stateformat

import (
	"encoding/json"
	"sort"
	"time"
)

type metadataKeyWire struct {
	ID        string  `json:"key_id"`
	Algorithm string  `json:"algorithm"`
	PublicKey string  `json:"public_key"`
	Status    string  `json:"status"`
	Before    string  `json:"not_before"`
	After     string  `json:"not_after"`
	Revoked   *string `json:"revoked_at"`
	Reason    *string `json:"revocation_reason"`
	Incident  *string `json:"incident_id"`
}
type metadataWire struct {
	Schema   string            `json:"schema_version"`
	Epoch    string            `json:"control_plane_authority_epoch"`
	Sequence string            `json:"authority_sequence"`
	Root     string            `json:"root_version"`
	Version  string            `json:"metadata_version"`
	Issued   string            `json:"issued_at"`
	Until    string            `json:"valid_until"`
	Role     string            `json:"role"`
	Refresh  string            `json:"next_refresh_at"`
	Revoked  []string          `json:"revoked_key_ids"`
	Keys     []metadataKeyWire `json:"keys"`
}

var nullableString = &wireSchema{kind: 's', nullable: true}
var metadataKeySchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"key_id":            wireString,
	"algorithm":         wireString,
	"public_key":        wireString,
	"status":            wireString,
	"not_before":        wireString,
	"not_after":         wireString,
	"revoked_at":        nullableString,
	"revocation_reason": nullableString,
	"incident_id":       nullableString,
}}

var metadataSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"root_version":                  wireString,
	"metadata_version":              wireString,
	"issued_at":                     wireString,
	"valid_until":                   wireString,
	"role":                          wireString,
	"next_refresh_at":               wireString,
	"revoked_key_ids":               &wireSchema{kind: 'a', min: 0, max: 4096, element: wireString},
	"keys":                          &wireSchema{kind: 'a', min: 1, max: 64, element: metadataKeySchema},
}}

func validRevokedIDs(ids []string) bool {
	if len(ids) > 4096 {
		return false
	}
	last := ""
	for _, id := range ids {
		if _, e := parseDigest(id); e != nil || id <= last {
			return false
		}
		last = id
	}
	return true
}

type keyWindow struct{ start, end time.Time }

func decodeMetadataPayload(body []byte) (artifactBindings, MetadataMaterial, error) {
	if strictCanonical(body, 65536, metadataSchema) != nil {
		return artifactBindings{}, MetadataMaterial{}, ErrInvalidFormat
	}
	var w metadataWire
	if json.Unmarshal(body, &w) != nil {
		return artifactBindings{}, MetadataMaterial{}, ErrInvalidFormat
	}
	c := formatCheck{ok: true}
	b := artifactBindings{Kind: Metadata}
	b.Epoch = c.u(w.Epoch)
	b.Sequence = c.u(w.Sequence)
	b.RootVersion = c.u(w.Root)
	b.MetadataVersion = c.u(w.Version)
	b.IssuedAt = c.timestamp(w.Issued)
	b.ValidUntil = c.timestamp(w.Until)
	refresh := c.timestamp(w.Refresh)
	c.require(w.Schema == "node-state-trust-metadata.v1" && w.Role == "node_state" && validRevokedIDs(w.Revoked))
	c.require(validWindow(b.IssuedAt, b.ValidUntil, 7*24*time.Hour) && !refresh.Before(b.IssuedAt) && b.ValidUntil.Sub(refresh) >= 48*time.Hour)
	windows := make([]keyWindow, 0, len(w.Keys))
	last := ""
	for _, k := range w.Keys {
		c.require(validRootKey(rootKeyWire{ID: k.ID, Algorithm: k.Algorithm, PublicKey: k.PublicKey}) && k.ID > last)
		last = k.ID
		before, after := c.timestamp(k.Before), c.timestamp(k.After)
		c.require(before.Before(after))
		i := sort.SearchStrings(w.Revoked, k.ID)
		revoked := i < len(w.Revoked) && w.Revoked[i] == k.ID
		switch k.Status {
		case "active", "retiring":
			c.require(k.Revoked == nil && k.Reason == nil && k.Incident == nil && !revoked && before.Before(b.ValidUntil) && after.After(b.IssuedAt))
			if k.Status == "active" {
				windows = append(windows, keyWindow{before, after})
			}
		case "revoked":
			if k.Revoked == nil || k.Reason == nil {
				return artifactBindings{}, MetadataMaterial{}, ErrInvalidFormat
			}
			at := c.timestamp(*k.Revoked)
			c.require(!at.After(b.IssuedAt) && revoked)
			switch *k.Reason {
			case "scheduled":
				c.require(k.Incident == nil)
			case "security_incident":
				if k.Incident == nil {
					return artifactBindings{}, MetadataMaterial{}, ErrInvalidFormat
				}
				_, e := parseUUID(*k.Incident)
				c.require(e == nil)
			default:
				c.require(false)
			}
		default:
			c.require(false)
		}
	}
	sort.Slice(windows, func(i, j int) bool { return windows[i].start.Before(windows[j].start) })
	cursor := b.IssuedAt
	for _, w := range windows {
		if w.start.After(cursor) {
			c.require(false)
			break
		}
		if w.end.After(cursor) {
			cursor = w.end
		}
	}
	c.require(!cursor.Before(b.ValidUntil))
	if !c.ok {
		return artifactBindings{}, MetadataMaterial{}, ErrInvalidFormat
	}
	return b, metadataMaterialFrom(w), nil
}

func parseMetadataPayload(body []byte) (artifactBindings, error) {
	b, m, e := decodeMetadataPayload(body)
	m.Clear()
	return b, e
}
