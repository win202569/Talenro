package stateformat

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type rootKeyWire struct {
	ID        string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
}
type rootWire struct {
	Schema    string        `json:"schema_version"`
	Epoch     string        `json:"control_plane_authority_epoch"`
	Sequence  string        `json:"authority_sequence"`
	Version   string        `json:"root_version"`
	Threshold string        `json:"threshold"`
	Keys      []rootKeyWire `json:"keys"`
}
type rotationWire struct {
	Schema   string   `json:"schema_version"`
	Epoch    string   `json:"control_plane_authority_epoch"`
	Sequence string   `json:"authority_sequence"`
	Previous string   `json:"previous_root_version"`
	Root     rootWire `json:"new_root_set"`
	Issued   string   `json:"issued_at"`
	Until    string   `json:"valid_until"`
}

var rootKeySchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"key_id":     wireString,
	"algorithm":  wireString,
	"public_key": wireString,
}}

var rootSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"root_version":                  wireString,
	"threshold":                     wireString,
	"keys":                          &wireSchema{kind: 'a', min: 1, max: 5, element: rootKeySchema},
}}

var rotationSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"previous_root_version":         wireString,
	"new_root_set":                  rootSchema,
	"issued_at":                     wireString,
	"valid_until":                   wireString,
}}

func validRootKey(k rootKeyWire) bool {
	id, e := parseDigest(k.ID)
	if e != nil || k.Algorithm != "ed25519" {
		return false
	}
	key, e := parseKey(k.PublicKey)
	return e == nil && sha256.Sum256(key[:]) == id
}
func rootBindings(w rootWire) (artifactBindings, error) {
	c := formatCheck{ok: true}
	b := artifactBindings{Kind: InitialRoot}
	b.Epoch = c.u(w.Epoch)
	b.Sequence = c.u(w.Sequence)
	b.RootVersion = c.u(w.Version)
	threshold := c.u(w.Threshold)
	c.require(w.Schema == "node-state-root-set.v1" && len(w.Keys) >= 1 && len(w.Keys) <= 5 && threshold <= uint64(len(w.Keys)))
	c.require(b.RootVersion != 1 || len(w.Keys) == 1 && threshold == 1)
	last := ""
	for _, k := range w.Keys {
		c.require(validRootKey(k) && k.ID > last)
		last = k.ID
	}
	if !c.ok {
		return artifactBindings{}, ErrInvalidFormat
	}
	return b, nil
}
func decodeRootPayload(body []byte) (artifactBindings, RootMaterial, error) {
	if strictCanonical(body, 65536, rootSchema) != nil {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	var w rootWire
	if json.Unmarshal(body, &w) != nil {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	b, e := rootBindings(w)
	if e != nil {
		return artifactBindings{}, RootMaterial{}, e
	}
	return b, rootMaterialFrom(w), nil
}
func decodeRotationPayload(body []byte) (artifactBindings, RootMaterial, error) {
	if strictCanonical(body, 65536, rotationSchema) != nil {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	var w rotationWire
	if json.Unmarshal(body, &w) != nil {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	b, err := rootBindings(w.Root)
	if err != nil {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	c := formatCheck{ok: true}
	epoch, sequence, previous := c.u(w.Epoch), c.u(w.Sequence), c.u(w.Previous)
	c.require(w.Schema == "node-state-root-rotation.v1" && b.Epoch == epoch && b.Sequence == sequence && previous < 9223372036854775807 && b.RootVersion == previous+1)
	b.Kind = RootRotation
	b.PreviousRootVersion = previous
	b.IssuedAt = c.timestamp(w.Issued)
	b.ValidUntil = c.timestamp(w.Until)
	c.require(validWindow(b.IssuedAt, b.ValidUntil, 7*24*time.Hour))
	if !c.ok {
		return artifactBindings{}, RootMaterial{}, ErrInvalidFormat
	}
	return b, rootMaterialFrom(w.Root), nil
}

func parseRootPayload(body []byte) (artifactBindings, error) {
	b, m, e := decodeRootPayload(body)
	m.Clear()
	return b, e
}
func parseRotationPayload(body []byte) (artifactBindings, error) {
	b, m, e := decodeRotationPayload(body)
	m.Clear()
	return b, e
}

// Called only after the complete rotation passed the sole strict parser.
func embeddedRootSet(body []byte) (*ParsedRootSet, error) {
	var raw struct {
		Root json.RawMessage `json:"new_root_set"`
	}
	if json.Unmarshal(body, &raw) != nil {
		return nil, ErrInvalidFormat
	}
	defer clear(raw.Root)
	return ParseRootSet(raw.Root)
}
