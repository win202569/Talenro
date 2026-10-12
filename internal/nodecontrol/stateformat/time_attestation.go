package stateformat

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type timeAttestationWire struct {
	Schema     string `json:"schema_version"`
	Epoch      string `json:"control_plane_authority_epoch"`
	Checkpoint string `json:"node_authority_checkpoint_sequence"`
	Node       string `json:"node_id"`
	Nonce      string `json:"request_nonce"`
	Issued     string `json:"issued_at"`
	Until      string `json:"valid_until"`
}

var timeAttestationSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version": wireString, "control_plane_authority_epoch": wireString,
	"node_authority_checkpoint_sequence": wireString, "node_id": wireString,
	"request_nonce": wireString, "issued_at": wireString, "valid_until": wireString,
}}

func parseTimeAttestationPayload(body []byte) (artifactBindings, error) {
	if strictCanonical(body, 65536, timeAttestationSchema) != nil {
		return artifactBindings{}, ErrInvalidFormat
	}
	var w timeAttestationWire
	if json.Unmarshal(body, &w) != nil {
		return artifactBindings{}, ErrInvalidFormat
	}
	c := formatCheck{ok: true}
	b := artifactBindings{Kind: TimeAttestation}
	c.require(w.Schema == "node-time-attestation.v1")
	b.Epoch = c.u(w.Epoch)
	b.CheckpointSequence = c.u(w.Checkpoint)
	var e error
	b.NodeID, e = parseUUID(w.Node)
	c.require(e == nil)
	b.IssuedAt = c.timestamp(w.Issued)
	b.ValidUntil = c.timestamp(w.Until)
	c.require(validWindow(b.IssuedAt, b.ValidUntil, 5*time.Minute))
	if len(w.Nonce) != 43 {
		return artifactBindings{}, ErrInvalidFormat
	}
	n, e := base64.RawURLEncoding.Strict().Decode(b.RequestNonce[:], []byte(w.Nonce))
	c.require(e == nil && n == 32 && base64.RawURLEncoding.EncodeToString(b.RequestNonce[:]) == w.Nonce)
	if !c.ok {
		return artifactBindings{}, ErrInvalidFormat
	}
	return b, nil
}
