package stateformat

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

// Materials are detached, owned format facts, not verified authority.
// Do not copy an owner or use it concurrently with Clear. Borrowing is allowed.
type PublicKeyMaterial struct {
	KeyID     contracts.Digest
	PublicKey [32]byte
}
type RootMaterial struct {
	Version, Threshold uint64
	Keys               []PublicKeyMaterial
}
type MetadataKeyMaterial struct {
	PublicKeyMaterial
	Status              string
	NotBefore, NotAfter time.Time
	HasRevocation       bool
	RevokedAt           time.Time
	RevocationReason    string
	HasIncident         bool
	IncidentID          uuid.UUID
}
type MetadataMaterial struct {
	Version, RootVersion uint64
	Keys                 []MetadataKeyMaterial
	RevokedIDs           []contracts.Digest
}
type SignatureShare struct {
	KeyID     contracts.Digest
	Signature [64]byte
}
type SignatureMaterial struct{ Current, New, Metadata []SignatureShare }

func (m *RootMaterial) Clear() {
	if m != nil {
		clear(m.Keys)
		*m = RootMaterial{}
	}
}
func (m *MetadataMaterial) Clear() {
	if m != nil {
		clear(m.Keys)
		clear(m.RevokedIDs)
		*m = MetadataMaterial{}
	}
}
func (m *SignatureMaterial) Clear() {
	if m != nil {
		clear(m.Current)
		clear(m.New)
		clear(m.Metadata)
		*m = SignatureMaterial{}
	}
}
func (a *ParsedArtifact) RootMaterial() (RootMaterial, error) {
	if a == nil || (a.bindings.Kind != InitialRoot && a.bindings.Kind != RootRotation) {
		return RootMaterial{}, ErrInvalidFormat
	}
	m := a.root
	m.Keys = append([]PublicKeyMaterial(nil), m.Keys...)
	return m, nil
}
func (a *ParsedArtifact) MetadataMaterial() (MetadataMaterial, error) {
	if a == nil || a.bindings.Kind != Metadata {
		return MetadataMaterial{}, ErrInvalidFormat
	}
	m := a.metadata
	m.Keys = append([]MetadataKeyMaterial(nil), m.Keys...)
	m.RevokedIDs = append([]contracts.Digest(nil), m.RevokedIDs...)
	return m, nil
}
func (a *ParsedArtifact) SignatureMaterial() (SignatureMaterial, error) {
	if a == nil || (a.bindings.Kind != RootRotation && a.bindings.Kind != Metadata) {
		return SignatureMaterial{}, ErrInvalidFormat
	}
	m := a.signatures
	m.Current = append([]SignatureShare(nil), m.Current...)
	m.New = append([]SignatureShare(nil), m.New...)
	m.Metadata = append([]SignatureShare(nil), m.Metadata...)
	return m, nil
}

// These converters consume only DTOs already validated by the sole wire parser.
func rootMaterialFrom(w rootWire) (m RootMaterial) {
	done := false
	defer func() {
		if !done {
			m.Clear()
		}
	}()
	m.Version, _ = parseU(w.Version)
	m.Threshold, _ = parseU(w.Threshold)
	m.Keys = make([]PublicKeyMaterial, len(w.Keys))
	for i, k := range w.Keys {
		m.Keys[i].KeyID, _ = parseDigest(k.ID)
		m.Keys[i].PublicKey, _ = parseKey(k.PublicKey)
	}
	done = true
	return
}
func metadataMaterialFrom(w metadataWire) (m MetadataMaterial) {
	done := false
	defer func() {
		if !done {
			m.Clear()
		}
	}()
	m.Version, _ = parseU(w.Version)
	m.RootVersion, _ = parseU(w.Root)
	m.Keys = make([]MetadataKeyMaterial, len(w.Keys))
	m.RevokedIDs = make([]contracts.Digest, len(w.Revoked))
	for i, k := range w.Keys {
		v := &m.Keys[i]
		v.KeyID, _ = parseDigest(k.ID)
		v.PublicKey, _ = parseKey(k.PublicKey)
		v.Status = k.Status
		v.NotBefore, _ = parseTime(k.Before)
		v.NotAfter, _ = parseTime(k.After)
		if k.Revoked != nil {
			v.HasRevocation = true
			v.RevokedAt, _ = parseTime(*k.Revoked)
			v.RevocationReason = *k.Reason
		}
		if k.Incident != nil {
			v.HasIncident = true
			v.IncidentID, _ = parseUUID(*k.Incident)
		}
	}
	for i, id := range w.Revoked {
		m.RevokedIDs[i], _ = parseDigest(id)
	}
	done = true
	return
}
func sharesMaterialFrom(w []shareWire) (m []SignatureShare) {
	done := false
	defer func() {
		if !done {
			clear(m)
		}
	}()
	m = make([]SignatureShare, len(w))
	for i, s := range w {
		m[i].KeyID, _ = parseDigest(s.KeyID)
		m[i].Signature, _ = parseSignature(s.Signature)
	}
	done = true
	return
}
