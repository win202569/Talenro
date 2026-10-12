package stateformat

import (
	"bytes"
	"crypto/sha256"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// RootSetBindings describes format coordinates, never an initial trust anchor.
type RootSetBindings struct{ Epoch, Sequence, Version uint64 }

// ParsedRootSet owns format facts. Do not shallow-copy or access concurrently with Clear.
type ParsedRootSet struct {
	bindings RootSetBindings
	payload  []byte
	digest   contracts.Digest
	material RootMaterial
}

func ParseRootSet(payload []byte) (*ParsedRootSet, error) {
	b, m, e := decodeRootPayload(payload)
	if e != nil {
		return nil, e
	}
	return &ParsedRootSet{bindings: RootSetBindings{b.Epoch, b.Sequence, b.RootVersion}, payload: bytes.Clone(payload), digest: sha256.Sum256(payload), material: m}, nil
}
func (r *ParsedRootSet) Bindings() RootSetBindings {
	if r == nil {
		return RootSetBindings{}
	}
	return r.bindings
}
func (r *ParsedRootSet) Payload() []byte {
	if r == nil {
		return nil
	}
	return bytes.Clone(r.payload)
}
func (r *ParsedRootSet) PayloadDigest() contracts.Digest {
	if r == nil {
		return contracts.Digest{}
	}
	return r.digest
}
func (r *ParsedRootSet) RootMaterial() (RootMaterial, error) {
	if r == nil || r.bindings.Version == 0 {
		return RootMaterial{}, ErrInvalidFormat
	}
	m := r.material
	m.Keys = append([]PublicKeyMaterial(nil), m.Keys...)
	return m, nil
}
func (r *ParsedRootSet) Clear() {
	if r != nil {
		clear(r.payload)
		r.material.Clear()
		*r = ParsedRootSet{}
	}
}
func (a *ParsedArtifact) EmbeddedRootSet() (*ParsedRootSet, error) {
	if a == nil || a.bindings.Kind != RootRotation || a.embeddedRoot == nil {
		return nil, ErrInvalidFormat
	}
	return ParseRootSet(a.embeddedRoot.payload)
}
func (a *ParsedArtifact) DesiredSlotIDs() ([]string, error) {
	if a == nil || a.bindings.Kind != Desired {
		return nil, ErrInvalidFormat
	}
	return append([]string(nil), a.slots...), nil
}
