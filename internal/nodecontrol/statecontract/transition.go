package statecontract

import "talenro.local/platform/internal/nodecontrol/contracts"

// visibleIdentities rejects contradictions across references, not merely
// corresponding fields. This is not reconstruction of missing materials.
func visibleIdentities(in ModelInput) error {
	var refs []DesiredBinding
	for _, x := range [2]MainStateContentModel{in.Previous, in.Proposed} {
		if x.Desired.Header.Coverage == Present {
			refs = append(refs, desiredOf(x.Desired.Mark))
		}
		for _, r := range [2]AppliedRecord{x.Supervisor.Verified, x.Supervisor.Applied} {
			if r.Coverage == Present {
				refs = append(refs, r.Desired)
			}
		}
		if x.Supervisor.Transition.Coverage == Present {
			refs = append(refs, x.Supervisor.Transition.PriorApplied, x.Supervisor.Transition.CandidateVerified)
		}
		if x.LKG.Header.Coverage == Present {
			refs = append(refs, x.LKG.HighestAccepted)
			for _, e := range x.LKG.Entries {
				refs = append(refs, e.Desired)
			}
		}
	}
	for i, a := range refs {
		for _, b := range refs[:i] {
			if a.Generation == b.Generation || a.Sequence == b.Sequence || a.Payload == b.Payload {
				if a != b {
					return ErrConflict
				}
			} else if (a.Generation < b.Generation) != (a.Sequence < b.Sequence) {
				return ErrConflict
			}
		}
	}
	a, b := in.Previous.Desired, in.Proposed.Desired
	if a.Header.Coverage == Present && b.Header.Coverage == Present && a.Mark.Material.Payload == b.Mark.Material.Payload {
		if a.RootPayload != b.RootPayload || a.MetadataPayload != b.MetadataPayload || a.SigningKeyID != b.SigningKeyID || a.ResourceVersion != b.ResourceVersion || a.ResourcePayload != b.ResourcePayload || a.InventoryVersion != b.InventoryVersion {
			return ErrConflict
		}
	}
	r, s := in.Previous.Recovery, in.Proposed.Recovery
	if r.Header.Coverage == Present && s.Header.Coverage == Present && r.Mark.Material.Payload == s.Mark.Material.Payload {
		if r.RootPayload != s.RootPayload || r.MetadataPayload != s.MetadataPayload || r.SigningKeyID != s.SigningKeyID {
			return ErrConflict
		}
	}
	return nil
}

func compareStream(a, b StreamMark) error {
	if b.Version < a.Version || b.Sequence < a.Sequence {
		return ErrRollback
	}
	if a.Version == b.Version {
		if a.Sequence != b.Sequence || a.Material.Payload != b.Material.Payload {
			return ErrConflict
		}
	} else if b.Sequence <= a.Sequence || a.Material.Payload == b.Material.Payload {
		return ErrConflict
	}
	return nil
}
func compareVersion(a, b VersionMark) error {
	if b.Version < a.Version {
		return ErrRollback
	}
	if (a.Version == b.Version) != (a.Material.Payload == b.Material.Payload) {
		return ErrConflict
	}
	return nil
}
func compareBinding(a, b DesiredBinding) error {
	return compareStream(StreamMark{a.Generation, a.Sequence, ArtifactRef{Payload: a.Payload}}, StreamMark{b.Generation, b.Sequence, ArtifactRef{Payload: b.Payload}})
}
func subset(a, b []contracts.Digest) bool {
	for _, v := range a {
		if !contains(b, v) {
			return false
		}
	}
	return true
}
func compareRecord(a, b AppliedRecord) error {
	if a.Coverage == Present {
		if b.Coverage != Present {
			return ErrRollback
		}
		return compareBinding(a.Desired, b.Desired)
	}
	return nil
}
func validateTransition(in ModelInput) error {
	a, b := in.Previous, in.Proposed
	as, bs := a.Scope, b.Scope
	as.TimeFloor = OptionalTime{}
	bs.TimeFloor = OptionalTime{}
	if as != bs || a.Identity.IdentityEpoch != b.Identity.IdentityEpoch || a.ServerTrust.Agent.TrustDomain != b.ServerTrust.Agent.TrustDomain || a.ServerTrust.Bootstrap.TrustDomain != b.ServerTrust.Bootstrap.TrustDomain {
		return ErrBindingMismatch
	}
	if a.Scope.TimeFloor.Present && (!b.Scope.TimeFloor.Present || b.Scope.TimeFloor.UnixSeconds < a.Scope.TimeFloor.UnixSeconds) {
		return ErrRollback
	}
	for i, h := range headers(a) {
		if h.Coverage == Present && headers(b)[i].Coverage != Present {
			return ErrRollback
		}
	}
	if b.Checkpoint.Sequence < a.Checkpoint.Sequence {
		return ErrRollback
	}
	pairs := [][2]StreamMark{{a.Root.Mark, b.Root.Mark}, {a.Metadata.Mark, b.Metadata.Mark}, {a.Resources.Mark, b.Resources.Mark}, {a.HostPolicy.Mark, b.HostPolicy.Mark}, {a.ServerTrust.Bootstrap.Mark, b.ServerTrust.Bootstrap.Mark}, {a.ServerTrust.Agent.Mark, b.ServerTrust.Agent.Mark}}
	if a.Desired.Header.Coverage == Present {
		pairs = append(pairs, [2]StreamMark{a.Desired.Mark, b.Desired.Mark})
	}
	if a.Recovery.Header.Coverage == Present {
		pairs = append(pairs, [2]StreamMark{a.Recovery.Mark, b.Recovery.Mark})
	}
	for _, p := range pairs {
		if e := compareStream(p[0], p[1]); e != nil {
			return e
		}
	}
	for _, p := range [][2]VersionMark{{a.Releases.Manifest, b.Releases.Manifest}, {a.Releases.Map, b.Releases.Map}} {
		if e := compareVersion(p[0], p[1]); e != nil {
			return e
		}
	}
	if !subset(a.Metadata.RevokedIDs, b.Metadata.RevokedIDs) || !subset(a.ServerTrust.Agent.Deauthorized, b.ServerTrust.Agent.Deauthorized) || !subset(a.ServerTrust.Bootstrap.Deauthorized, b.ServerTrust.Bootstrap.Deauthorized) {
		return ErrRollback
	}
	if a.Identity.Leaf.Payload == b.Identity.Leaf.Payload && (a.Identity.IssuerID != b.Identity.IssuerID || a.Identity.Serial != b.Identity.Serial || a.Identity.PublicKey != b.Identity.PublicKey || a.Identity.Window != b.Identity.Window) {
		return ErrConflict
	}
	if e := compareRecord(a.Supervisor.Verified, b.Supervisor.Verified); e != nil {
		return e
	}
	if e := compareRecord(a.Supervisor.Applied, b.Supervisor.Applied); e != nil {
		return e
	}
	if a.LKG.Header.Coverage == Present {
		if e := compareBinding(a.LKG.HighestAccepted, b.LKG.HighestAccepted); e != nil {
			return e
		}
	}
	if len(in.RootTransition) > 32 {
		return ErrInvalidInput
	}
	if a.Root.Mark.Version == b.Root.Mark.Version {
		if len(in.RootTransition) != 0 {
			return ErrConflict
		}
		return nil
	}
	if len(in.RootTransition) == 0 {
		return ErrBindingMismatch
	}
	prev, seq, payload := a.Root.Mark.Version, a.Root.Mark.Sequence, a.Root.Mark.Material.Payload
	seenRoots := []contracts.Digest{payload}
	var seenRotations []contracts.Digest
	for _, r := range in.RootTransition {
		if !version(r.PreviousVersion) || !version(r.NewVersion) || !version(r.Sequence) || !digest(r.OldRootPayload) || !digest(r.NewRootPayload) || !artifact(r.Material, RootRotation) {
			return ErrInvalidInput
		}
		if contains(seenRoots, r.NewRootPayload) || contains(seenRotations, r.Material.Payload) {
			return ErrConflict
		}
		if prev == maxVersion || r.PreviousVersion != prev || r.NewVersion != prev+1 || r.Sequence <= seq || r.OldRootPayload != payload || r.NewRootPayload == payload {
			return ErrBindingMismatch
		}
		prev, seq, payload = r.NewVersion, r.Sequence, r.NewRootPayload
		seenRoots = append(seenRoots, payload)
		seenRotations = append(seenRotations, r.Material.Payload)
	}
	if prev != b.Root.Mark.Version || seq != b.Root.Mark.Sequence || payload != b.Root.Mark.Material.Payload {
		return ErrBindingMismatch
	}
	return nil
}
