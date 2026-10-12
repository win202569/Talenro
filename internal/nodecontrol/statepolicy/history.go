package statepolicy

import (
	"bytes"
	"crypto/sha256"
	"github.com/google/uuid"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"time"
)

func sortedIDs(ids []contracts.Digest) bool {
	for i, id := range ids {
		if id == (contracts.Digest{}) || (i > 0 && bytes.Compare(ids[i-1][:], id[:]) >= 0) {
			return false
		}
	}
	return true
}
func hasID(ids []contracts.Digest, id contracts.Digest) bool {
	i := sort.Search(len(ids), func(i int) bool { return bytes.Compare(ids[i][:], id[:]) >= 0 })
	return i < len(ids) && ids[i] == id
}
func sameKey(a, b stateformat.MetadataKeyMaterial) bool {
	return a.PublicKeyMaterial == b.PublicKeyMaterial && a.NotBefore.Equal(b.NotBefore) && a.NotAfter.Equal(b.NotAfter)
}
func validKeyRecord(k KeyRecord, ledger []contracts.Digest) bool {
	m := k.Material
	if sha256.Sum256(m.PublicKey[:]) != m.KeyID || !validInstant(m.NotBefore) || !validInstant(m.NotAfter) || !m.NotAfter.After(m.NotBefore) {
		return false
	}
	if k.DeadlinesKnown && (!validInstant(k.LastSnapshotUntil) || !validInstant(k.LastCompositeDeadline) || k.LastCompositeDeadline.Before(k.LastSnapshotUntil)) {
		return false
	}
	switch m.Status {
	case "active", "retiring":
		return !hasID(ledger, m.KeyID) && !m.HasRevocation && m.RevokedAt.IsZero() && m.RevocationReason == "" && !m.HasIncident && m.IncidentID == uuid.Nil && k.FirstRevokedAt.IsZero()
	case "revoked":
		return hasID(ledger, m.KeyID) && m.HasRevocation && validInstant(m.RevokedAt) && k.FirstRevokedAt.Equal(m.RevokedAt) && ((m.RevocationReason == "scheduled" && !m.HasIncident && m.IncidentID == uuid.Nil) || (m.RevocationReason == "security_incident" && m.HasIncident && m.IncidentID != uuid.Nil))
	default:
		return false
	}
}
func historyObservations(f *verifiedFacts, in ModelInput, known map[contracts.Digest]KeyRecord) (map[contracts.Digest]KeyHistoryObservation, error) {
	if len(in.History) > 4096 {
		return nil, fail(AssuranceMissing, UnsupportedRange)
	}
	items := append([]KeyHistoryObservation(nil), in.History...)
	defer clear(items)
	sort.Slice(items, func(i, j int) bool { return bytes.Compare(items[i].KeyID[:], items[j].KeyID[:]) < 0 })
	result := make(map[contracts.Digest]KeyHistoryObservation, len(items))
	base := contracts.Digest{}
	if in.Old != nil {
		base = in.Old.StateID
	}
	for i, h := range items {
		if i > 0 && h.KeyID == items[i-1].KeyID {
			return nil, fail(InvalidInput, Malformed)
		}
		if e := observationError(h.Status); e != nil {
			return nil, e
		}
		if h.ObservationID == (contracts.Digest{}) || h.BaseStateID != base || h.MetadataPayload != f.metadata.id.PayloadDigest {
			return nil, fail(AssuranceMissing, MissingHistory)
		}
		_, exists := known[h.KeyID]
		for _, k := range f.metadata.metadata.Keys {
			if k.KeyID == h.KeyID {
				exists = true
			}
		}
		if !exists {
			return nil, fail(AssuranceMissing, MissingHistory)
		}
		if h.DeadlinesKnown && (!validInstant(h.LastSnapshotUntil) || !validInstant(h.LastCompositeDeadline) || h.LastCompositeDeadline.Before(h.LastSnapshotUntil)) {
			return nil, fail(InvalidInput, InvalidTime)
		}
		if prior, ok := known[h.KeyID]; ok && prior.DeadlinesKnown && h.DeadlinesKnown && (h.LastSnapshotUntil.Before(prior.LastSnapshotUntil) || h.LastCompositeDeadline.Before(prior.LastCompositeDeadline)) {
			return nil, fail(SecurityFault, HistoryConflict)
		}
		result[h.KeyID] = h
	}
	return result, nil
}
func checkHistory(f *verifiedFacts, in ModelInput) (keys []KeyRecord, revoked []contracts.Digest, err error) {
	if f == nil {
		return nil, nil, fail(InvalidInput, Malformed)
	}
	next := f.metadata.metadata
	if len(next.RevokedIDs) >= 4096 || (in.Old != nil && (len(in.Old.RevokedIDs) >= 4096 || len(in.Old.Keys) > 4096)) {
		return nil, nil, fail(AssuranceMissing, UnsupportedRange)
	}
	if !sortedIDs(next.RevokedIDs) {
		return nil, nil, fail(InvalidInput, Malformed)
	}
	known := make(map[contracts.Digest]KeyRecord)
	var oldIDs []contracts.Digest
	if in.Old != nil {
		oldIDs = in.Old.RevokedIDs
		if !sortedIDs(oldIDs) {
			return nil, nil, fail(SecurityFault, HistoryConflict)
		}
		for _, id := range oldIDs {
			if !hasID(next.RevokedIDs, id) {
				return nil, nil, fail(SecurityFault, HistoryConflict)
			}
		}
		for i, k := range in.Old.Keys {
			if !validKeyRecord(k, oldIDs) || (i > 0 && bytes.Compare(in.Old.Keys[i-1].Material.KeyID[:], k.Material.KeyID[:]) >= 0) {
				return nil, nil, fail(SecurityFault, HistoryConflict)
			}
			known[k.Material.KeyID] = k
		}
	}
	observations, e := historyObservations(f, in, known)
	if e != nil {
		return nil, nil, e
	}
	// Reconcile observations before branching on presence in the new snapshot.
	// A removed tombstone still carries the same historical constraints.
	if in.Old != nil {
		for _, prior := range in.Old.Keys {
			h, ok := observations[prior.Material.KeyID]
			if !ok || !h.DeadlinesKnown {
				continue
			}
			if prior.Material.Status == "revoked" {
				if prior.DeadlinesKnown && (!prior.LastSnapshotUntil.Equal(h.LastSnapshotUntil) || !prior.LastCompositeDeadline.Equal(h.LastCompositeDeadline)) {
					return nil, nil, fail(SecurityFault, HistoryConflict)
				}
				if prior.Material.RevocationReason == "scheduled" && h.LastCompositeDeadline.After(prior.Material.RevokedAt) {
					return nil, nil, fail(SecurityFault, HistoryConflict)
				}
			}
			prior.DeadlinesKnown = true
			prior.LastSnapshotUntil = h.LastSnapshotUntil
			prior.LastCompositeDeadline = h.LastCompositeDeadline
			known[prior.Material.KeyID] = prior
		}
	}
	present := make(map[contracts.Digest]bool, len(next.Keys))
	for _, m := range next.Keys {
		if present[m.KeyID] {
			return nil, nil, fail(InvalidInput, Malformed)
		}
		present[m.KeyID] = true
		prior, exists := known[m.KeyID]
		h, observed := observations[m.KeyID]
		if exists && !sameKey(prior.Material, m) {
			return nil, nil, fail(SecurityFault, HistoryConflict)
		}
		if hasID(oldIDs, m.KeyID) && m.Status != "revoked" {
			return nil, nil, fail(SecurityFault, HistoryConflict)
		}
		record := prior
		record.Material = m
		switch m.Status {
		case "active", "retiring":
			if exists && ((prior.Material.Status == "retiring" && m.Status == "active") || prior.Material.Status == "revoked") {
				return nil, nil, fail(SecurityFault, HistoryConflict)
			}
		case "revoked":
			if !hasID(next.RevokedIDs, m.KeyID) || !m.HasRevocation || !validInstant(m.RevokedAt) || m.RevokedAt.After(f.metadata.bindings.IssuedAt) {
				return nil, nil, fail(InvalidInput, Malformed)
			}
			if exists && prior.Material.Status == "revoked" {
				if prior.Material != m {
					return nil, nil, fail(SecurityFault, HistoryConflict)
				}
			} else {
				if !observed {
					return nil, nil, fail(AssuranceMissing, MissingHistory)
				}
				if m.RevocationReason == "security_incident" {
					if !m.HasIncident || m.IncidentID == uuid.Nil || h.IncidentID != m.IncidentID || h.IncidentKind != "online_signer_equivocation" || !h.IncidentOpen {
						return nil, nil, fail(AssuranceMissing, MissingHistory)
					}
					return nil, nil, fail(SecurityFault, EmergencyRevocation)
				}
				if m.RevocationReason != "scheduled" || m.HasIncident || m.IncidentID != uuid.Nil {
					return nil, nil, fail(InvalidInput, Malformed)
				}
				if !h.DeadlinesKnown || ((!exists || prior.Material.Status != "retiring") && !h.RetiringKnown) {
					return nil, nil, fail(AssuranceMissing, MissingHistory)
				}
				if m.RevokedAt.Before(h.LastCompositeDeadline) {
					return nil, nil, fail(Rejected, HistoryConflict)
				}
				record.FirstRevokedAt = m.RevokedAt
			}
		default:
			return nil, nil, fail(InvalidInput, Malformed)
		}
		if observed && h.DeadlinesKnown {
			record.DeadlinesKnown = true
			record.LastSnapshotUntil = h.LastSnapshotUntil
			record.LastCompositeDeadline = h.LastCompositeDeadline
		}
		known[m.KeyID] = record
	}
	for _, id := range next.RevokedIDs {
		if !hasID(oldIDs, id) {
			k, ok := known[id]
			if !ok || k.Material.Status != "revoked" {
				return nil, nil, fail(AssuranceMissing, MissingHistory)
			}
		}
	}
	if len(known) > 4096 {
		return nil, nil, fail(AssuranceMissing, UnsupportedRange)
	}
	// Iterate old ordered records, not a map, to keep failure precedence stable.
	if in.Old != nil {
		for _, old := range in.Old.Keys {
			k := known[old.Material.KeyID]
			if k.Material.Status != "revoked" || present[k.Material.KeyID] {
				continue
			}
			if !k.DeadlinesKnown {
				return nil, nil, fail(AssuranceMissing, MissingHistory)
			}
			until, ok := checkedAdd(k.FirstRevokedAt, 30*24*time.Hour)
			if !ok || !validInstant(in.Time.Now) || in.Time.Now.Before(until) || !in.Time.Now.After(k.LastSnapshotUntil) {
				return nil, nil, fail(Rejected, HistoryConflict)
			}
		}
	}
	keys = make([]KeyRecord, 0, len(known))
	for _, k := range known {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i].Material.KeyID[:], keys[j].Material.KeyID[:]) < 0 })
	return keys, append([]contracts.Digest(nil), next.RevokedIDs...), nil
}
