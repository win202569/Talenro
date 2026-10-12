package statecontract

import (
	"bytes"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const maxVersion uint64 = 1<<63 - 1

func version(n uint64) bool          { return n > 0 && n <= maxVersion }
func digest(d contracts.Digest) bool { return d != (contracts.Digest{}) }
func timestamp(n int64) bool         { return n >= -62135596800 && n <= 253402300799 }
func window(w TimeWindow) bool {
	return timestamp(w.Issued) && timestamp(w.Until) && w.Issued < w.Until
}
func node(s string) bool {
	u, e := uuid.Parse(s)
	return e == nil && u != uuid.Nil && u.Variant() == uuid.RFC4122 && u.String() == s
}
func artifact(a ArtifactRef, k ArtifactKind) bool {
	if a.Kind != k || !digest(a.Payload) {
		return false
	}
	signed := false
	switch k {
	case RootRotation, Metadata, TrustBundle, ResourceEnvelope, HostMemoryPolicy, DesiredSnapshot, RecoverySnapshot:
		signed = true
	}
	return digest(a.Envelope) == signed
}
func stream(m StreamMark, k ArtifactKind) bool {
	return version(m.Version) && version(m.Sequence) && artifact(m.Material, k)
}
func digests(a []contracts.Digest, min, max int) bool {
	if len(a) < min || len(a) > max {
		return false
	}
	for i, v := range a {
		if !digest(v) || (i > 0 && bytes.Compare(a[i-1][:], v[:]) >= 0) {
			return false
		}
	}
	return true
}
func slot(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func binding(b DesiredBinding) bool {
	return version(b.Generation) && version(b.Sequence) && digest(b.Payload)
}
func record(r AppliedRecord) bool {
	return r.Coverage == NeverEstablished && r.Desired == (DesiredBinding{}) || r.Coverage == Present && binding(r.Desired)
}
func domain(s string) bool {
	if len(s) < 1 || len(s) > 253 {
		return false
	}
	labels := strings.Split(s, ".")
	ip := len(labels) == 4
	for _, p := range labels {
		if len(p) < 1 || len(p) > 63 || p[0] == '-' || p[len(p)-1] == '-' {
			return false
		}
		n := 0
		numeric := true
		for _, c := range p {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
			if c < '0' || c > '9' {
				numeric = false
			} else if n <= 255 {
				n = n*10 + int(c-'0')
			}
		}
		ip = ip && numeric && n <= 255
	}
	return !ip
}
func bundle(b BundleRef, p BundlePurpose) bool {
	return b.Purpose == p && domain(b.TrustDomain) && version(b.AuthorityEpoch) && stream(b.Mark, TrustBundle) && digest(b.DeploymentKeyID) && digests(b.Deauthorized, 0, 128) && digest(b.DeauthorizedSetDigest)
}
func headers(x MainStateContentModel) [12]SectionHeader {
	return [12]SectionHeader{x.Checkpoint.Header, x.Root.Header, x.Metadata.Header, x.Identity.Header, x.ServerTrust.Header, x.Releases.Header, x.Resources.Header, x.HostPolicy.Header, x.Desired.Header, x.Recovery.Header, x.Supervisor.Header, x.LKG.Header}
}
func validateShape(x MainStateContentModel) error {
	if x.ModelVersion != 1 {
		return ErrInvalidInput
	}
	for i, h := range headers(x) {
		d := Domain(i + 1)
		if h.Domain != d {
			return ErrBindingMismatch
		}
		if h.Coverage != Present && !(h.Coverage == NeverEstablished && (d == DomainDesired || d == DomainRecovery || d == DomainLKG)) {
			return ErrIncompleteCoverage
		}
	}
	// All bounded collections are checked before traversing or copying any of them.
	if len(x.Root.KeyIDs) > 5 || len(x.Metadata.KeyIDs) > 64 || len(x.Metadata.RevokedIDs) > 4096 || len(x.ServerTrust.Bootstrap.Deauthorized) > 128 || len(x.ServerTrust.Agent.Deauthorized) > 128 || len(x.Desired.SlotIDs) > 8 || len(x.LKG.Entries) > 8 || len(x.Recovery.Local) > 64 || len(x.Recovery.Supervisor) > 16 {
		return ErrInvalidInput
	}
	s := x.Scope
	if !node(s.NodeID) || !version(s.AuthorityEpoch) || !digest(s.DeploymentID) || !digest(s.TrustAnchorID) || !digest(s.DeploymentKeySetID) || !digest(s.TimeProviderID) || (!s.TimeFloor.Present && s.TimeFloor.UnixSeconds != 0) || (s.TimeFloor.Present && !timestamp(s.TimeFloor.UnixSeconds)) {
		return ErrInvalidInput
	}
	guards := [3]GuardBinding{s.Main, s.Supervisor, s.Latch}
	for i, g := range guards {
		if !digest(g.CounterID) || !digest(g.SealKeyID) {
			return ErrInvalidInput
		}
		for j := 0; j < i; j++ {
			if guards[j].CounterID == g.CounterID || guards[j].SealKeyID == g.SealKeyID {
				return ErrBindingMismatch
			}
		}
	}
	c := x.Checkpoint
	if !node(c.NodeID) || !version(c.AuthorityEpoch) || !version(c.Sequence) {
		return ErrInvalidInput
	}
	r := x.Root
	if !stream(r.Mark, RootPayload) || !digests(r.KeyIDs, 1, 5) || r.Threshold == 0 || int(r.Threshold) > len(r.KeyIDs) {
		return ErrInvalidInput
	}
	m := x.Metadata
	if !stream(m.Mark, Metadata) || !version(m.RootVersion) || !digest(m.RootPayload) || !digests(m.KeyIDs, 1, 64) || !digests(m.RevokedIDs, 0, 4096) || !digest(m.RevokedSetDigest) || !artifact(m.History, MetadataHistory) || !window(m.Window) {
		return ErrInvalidInput
	}
	id := x.Identity
	if !version(id.IdentityEpoch) || !node(id.LineageID) || !timestamp(id.LineageUntil) || !digest(id.IssuerID) || !digest(id.Serial) || !digest(id.PublicKey) || !digest(id.KeyHandle) || !artifact(id.Leaf, CertificateDER) || !artifact(id.History, IdentityHistory) || !window(id.Window) {
		return ErrInvalidInput
	}
	if !bundle(x.ServerTrust.Bootstrap, BootstrapServer) || !bundle(x.ServerTrust.Agent, AgentServer) {
		return ErrInvalidInput
	}
	rel := x.Releases
	if !digest(rel.BuildID) || !version(rel.Manifest.Version) || !artifact(rel.Manifest.Material, ApprovedManifest) || !version(rel.Map.Version) || !artifact(rel.Map.Material, InstalledMap) {
		return ErrInvalidInput
	}
	re := x.Resources
	hp := x.HostPolicy
	if !stream(re.Mark, ResourceEnvelope) || !node(re.NodeID) || !version(re.AuthorityEpoch) || !digest(re.DeploymentKeyID) || !stream(hp.Mark, HostMemoryPolicy) || !node(hp.NodeID) || !version(hp.AuthorityEpoch) || !digest(hp.DeploymentKeyID) {
		return ErrInvalidInput
	}
	des := x.Desired
	if des.Header.Coverage == NeverEstablished {
		des.SlotIDs = nil
		if !reflect.DeepEqual(des, DesiredSection{Header: x.Desired.Header}) || len(x.Desired.SlotIDs) != 0 {
			return ErrIncompleteCoverage
		}
	} else {
		if !stream(des.Mark, DesiredSnapshot) || !version(des.InventoryVersion) || !version(des.ResourceVersion) || !digest(des.ResourcePayload) || !digest(des.RootPayload) || !digest(des.MetadataPayload) || !digest(des.SigningKeyID) || !window(des.Window) {
			return ErrInvalidInput
		}
		for i, s := range des.SlotIDs {
			if !slot(s) || (i > 0 && des.SlotIDs[i-1] >= s) {
				return ErrInvalidInput
			}
		}
	}
	rec := x.Recovery
	if rec.Header.Coverage == NeverEstablished {
		rec.Local = nil
		rec.Supervisor = nil
		if !reflect.DeepEqual(rec, RecoverySection{Header: x.Recovery.Header}) || len(x.Recovery.Local) != 0 || len(x.Recovery.Supervisor) != 0 {
			return ErrIncompleteCoverage
		}
	} else {
		if !stream(rec.Mark, RecoverySnapshot) || !version(rec.IdentityEpoch) || !node(rec.RecoveryID) || !version(rec.SessionVersion) || !artifact(rec.Session, RecoverySession) || !digest(rec.RootPayload) || !digest(rec.MetadataPayload) || !digest(rec.SigningKeyID) || !window(rec.Window) {
			return ErrInvalidInput
		}
		for i, f := range rec.Local {
			if !node(f.FaultID) || !node(f.IncidentID) || (i > 0 && rec.Local[i-1].FaultID >= f.FaultID) {
				return ErrInvalidInput
			}
		}
		for i, f := range rec.Supervisor {
			if !node(f.FaultID) || !node(f.IncidentID) || !digest(f.Evidence) || (i > 0 && rec.Supervisor[i-1].FaultID >= f.FaultID) {
				return ErrInvalidInput
			}
		}
	}
	sup := x.Supervisor
	for _, r := range [2]AppliedRecord{sup.Verified, sup.Applied} {
		if r.Coverage != Present && r.Coverage != NeverEstablished {
			return ErrIncompleteCoverage
		}
		if r.Coverage == NeverEstablished && r.Desired != (DesiredBinding{}) {
			return ErrIncompleteCoverage
		}
	}
	if !digest(sup.Identity) || !digest(sup.Guard.CounterID) || !digest(sup.Guard.SealKeyID) || !artifact(sup.State, SupervisorState) || !record(sup.Verified) || !record(sup.Applied) {
		return ErrInvalidInput
	}
	tr := sup.Transition
	if tr.Coverage != Present && tr.Coverage != NeverEstablished {
		return ErrIncompleteCoverage
	}
	if tr.Coverage == NeverEstablished {
		if tr != (TransitionRef{Coverage: NeverEstablished}) {
			return ErrIncompleteCoverage
		}
	} else if !artifact(tr.Material, TransitionCapsule) || !binding(tr.PriorApplied) || !binding(tr.CandidateVerified) {
		return ErrInvalidInput
	}
	l := x.LKG
	if l.Header.Coverage == NeverEstablished {
		l.Entries = nil
		if !reflect.DeepEqual(l, LKGSection{Header: x.LKG.Header}) || len(x.LKG.Entries) != 0 {
			return ErrIncompleteCoverage
		}
	} else {
		if !binding(l.HighestAccepted) || !artifact(l.History, LKGHistory) {
			return ErrInvalidInput
		}
		for i, e := range l.Entries {
			if !slot(e.SlotID) || (i > 0 && l.Entries[i-1].SlotID >= e.SlotID) || !binding(e.Desired) || !artifact(e.Process, ProcessSpec) || !artifact(e.Config, ConfigMaterial) || !artifact(e.Release, ReleaseMaterial) || !digest(e.RootPayload) || !digest(e.MetadataPayload) || !digest(e.SigningKeyID) || !timestamp(e.Deadline) {
				return ErrInvalidInput
			}
		}
	}
	return nil
}
