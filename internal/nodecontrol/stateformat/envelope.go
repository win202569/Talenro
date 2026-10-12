package stateformat

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type shareWire struct {
	KeyID     string `json:"key_id"`
	Signature string `json:"signature"`
}

type desiredEnvelopeWire struct {
	Schema    string `json:"schema_version"`
	Payload   string `json:"payload_b64url"`
	Digest    string `json:"payload_sha256"`
	KeyID     string `json:"key_id"`
	Signature string `json:"signature"`
}

type envelopeWire struct {
	Schema        string      `json:"schema_version"`
	Payload       string      `json:"payload_b64url"`
	Digest        string      `json:"payload_sha256"`
	Epoch         string      `json:"control_plane_authority_epoch"`
	Sequence      string      `json:"authority_sequence"`
	Root          string      `json:"root_version"`
	KeyID         string      `json:"key_id"`
	Signature     string      `json:"signature"`
	Previous      string      `json:"previous_root_version"`
	CurrentShares []shareWire `json:"current_signatures"`
	NewShares     []shareWire `json:"new_signatures"`
	Shares        []shareWire `json:"signatures"`
}

var shareSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"key_id":    wireString,
	"signature": wireString,
}}

var desiredEnvelopeSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version": wireString,
	"payload_b64url": wireString,
	"payload_sha256": wireString,
	"key_id":         wireString,
	"signature":      wireString,
}}

var initialEnvelopeSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"payload_b64url":                wireString,
	"payload_sha256":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"root_version":                  wireString,
}}

var rotationEnvelopeSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"payload_b64url":                wireString,
	"payload_sha256":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"root_version":                  wireString,
	"previous_root_version":         wireString,
	"current_signatures":            &wireSchema{kind: 'a', min: 0, max: 64, element: shareSchema},
	"new_signatures":                &wireSchema{kind: 'a', min: 0, max: 64, element: shareSchema},
}}

var metadataEnvelopeSchema = &wireSchema{kind: 'o', fields: map[string]*wireSchema{
	"schema_version":                wireString,
	"payload_b64url":                wireString,
	"payload_sha256":                wireString,
	"control_plane_authority_epoch": wireString,
	"authority_sequence":            wireString,
	"root_version":                  wireString,
	"signatures":                    &wireSchema{kind: 'a', min: 0, max: 64, element: shareSchema},
}}

func decodePayload(s string) ([]byte, error) {
	if len(s) == 0 || len(s) > 87382 {
		return nil, ErrInvalidFormat
	}
	for i := range s {
		if !(asciiAlnum(s[i]) || s[i] == '_' || s[i] == '-') {
			return nil, ErrInvalidFormat
		}
	}
	b, e := base64.RawURLEncoding.Strict().DecodeString(s)
	if e != nil || len(b) == 0 || len(b) > 65536 || base64.RawURLEncoding.EncodeToString(b) != s {
		clear(b)
		return nil, ErrInvalidFormat
	}
	return b, nil
}
func validShares(shares []shareWire) bool {
	if len(shares) > 64 {
		return false
	}
	last := ""
	for _, s := range shares {
		if _, e := parseDigest(s.KeyID); e != nil || s.KeyID <= last {
			return false
		}
		if _, e := parseSignature(s.Signature); e != nil {
			return false
		}
		last = s.KeyID
	}
	return true
}
func ParseEnvelope(kind Kind, envelope []byte) (result *ParsedArtifact, err error) {
	var schema *wireSchema
	var version string
	switch kind {
	case TimeAttestation:
		schema = desiredEnvelopeSchema
		version = "talenro-node-time-attestation/v1"
	case Desired:
		schema = desiredEnvelopeSchema
		version = "talenro-node-desired-state/v1"
	case InitialRoot:
		schema = initialEnvelopeSchema
		version = "talenro-node-state-root-initial/v1"
	case RootRotation:
		schema = rotationEnvelopeSchema
		version = "talenro-node-state-root-rotation/v1"
	case Metadata:
		schema = metadataEnvelopeSchema
		version = "talenro-node-state-metadata/v1"
	default:
		return nil, ErrInvalidFormat
	}
	if strictCanonical(envelope, 262144, schema) != nil {
		return nil, ErrInvalidFormat
	}
	var w envelopeWire
	if json.Unmarshal(envelope, &w) != nil || w.Schema != version {
		return nil, ErrInvalidFormat
	}
	payload, e := decodePayload(w.Payload)
	if e != nil {
		return nil, ErrInvalidFormat
	}
	// Until the successful ownership transfer, clear the only mutable inner copy
	// on every return or panic. Immutable decoder strings cannot be erased here.
	defer func() {
		if result == nil {
			clear(payload)
		}
	}()
	digest, e := parseDigest(w.Digest)
	if e != nil || sha256.Sum256(payload) != digest {
		return nil, ErrInvalidFormat
	}
	a := &ParsedArtifact{payload: payload}
	defer func() {
		if result == nil {
			a.Clear()
		}
	}()
	var b artifactBindings
	switch kind {
	case TimeAttestation:
		b, e = parseTimeAttestationPayload(payload)
	case Desired:
		b, a.slots, e = decodeDesiredPayload(payload)
	case InitialRoot:
		b, a.root, e = decodeRootPayload(payload)
	case RootRotation:
		b, a.root, e = decodeRotationPayload(payload)
		if e == nil {
			a.embeddedRoot, e = embeddedRootSet(payload)
		}
	case Metadata:
		b, a.metadata, e = decodeMetadataPayload(payload)
	}
	if e != nil {
		return nil, ErrInvalidFormat
	}
	c := formatCheck{ok: true}
	if kind == Desired || kind == TimeAttestation {
		b.SigningKeyID, e = parseDigest(w.KeyID)
		c.require(e == nil)
		b.Signature, e = parseSignature(w.Signature)
		c.require(e == nil)
	} else {
		epoch, sequence, root := c.u(w.Epoch), c.u(w.Sequence), c.u(w.Root)
		c.require(b.Epoch == epoch && b.Sequence == sequence && b.RootVersion == root)
		switch kind {
		case InitialRoot:
			c.require(root == 1)
		case RootRotation:
			previous := c.u(w.Previous)
			c.require(b.PreviousRootVersion == previous && validShares(w.CurrentShares) && validShares(w.NewShares))
		case Metadata:
			c.require(validShares(w.Shares))
		}
	}
	if !c.ok {
		return nil, ErrInvalidFormat
	}
	if kind == RootRotation {
		a.signatures.Current = sharesMaterialFrom(w.CurrentShares)
		a.signatures.New = sharesMaterialFrom(w.NewShares)
	}
	if kind == Metadata {
		a.signatures.Metadata = sharesMaterialFrom(w.Shares)
	}
	a.envelope = bytes.Clone(envelope)
	a.bindings = b
	a.payloadDigest = digest
	a.envelopeDigest = sha256.Sum256(envelope)
	return a, nil
}
func BuildDesiredEnvelope(payload []byte, digest, keyID contracts.Digest, signature [64]byte) (*ParsedArtifact, error) {
	if _, e := parseDesiredPayload(payload); e != nil || sha256.Sum256(payload) != digest || keyID == (contracts.Digest{}) || signature == ([64]byte{}) {
		return nil, ErrInvalidFormat
	}
	w := desiredEnvelopeWire{Schema: "talenro-node-desired-state/v1", Payload: base64.RawURLEncoding.EncodeToString(payload), Digest: hex.EncodeToString(digest[:]), KeyID: hex.EncodeToString(keyID[:]), Signature: base64.RawURLEncoding.EncodeToString(signature[:])}
	raw, e := json.Marshal(w)
	if e != nil {
		return nil, ErrInvalidFormat
	}
	defer clear(raw)
	canonical, e := jcs.Transform(raw)
	if e != nil {
		return nil, ErrInvalidFormat
	}
	defer clear(canonical)
	return ParseEnvelope(Desired, canonical)
}
