package statecontract

import "talenro.local/platform/internal/nodecontrol/contracts"

func contains(a []contracts.Digest, d contracts.Digest) bool {
	for _, v := range a {
		if v == d {
			return true
		}
	}
	return false
}
func desiredOf(m StreamMark) DesiredBinding {
	return DesiredBinding{m.Version, m.Sequence, m.Material.Payload}
}
func atMost(a, b DesiredBinding) bool {
	return a.Generation <= b.Generation && a.Sequence <= b.Sequence && (a.Generation != b.Generation || a == b)
}
func validateBindings(x MainStateContentModel) error {
	s := x.Scope
	if x.Checkpoint.NodeID != s.NodeID || x.Checkpoint.AuthorityEpoch != s.AuthorityEpoch || x.Resources.NodeID != s.NodeID || x.Resources.AuthorityEpoch != s.AuthorityEpoch || x.HostPolicy.NodeID != s.NodeID || x.HostPolicy.AuthorityEpoch != s.AuthorityEpoch || x.ServerTrust.Bootstrap.AuthorityEpoch != s.AuthorityEpoch || x.ServerTrust.Agent.AuthorityEpoch != s.AuthorityEpoch || x.Supervisor.Guard != s.Supervisor {
		return ErrBindingMismatch
	}
	if x.Metadata.RootVersion > x.Root.Mark.Version || (x.Metadata.RootVersion == x.Root.Mark.Version) != (x.Metadata.RootPayload == x.Root.Mark.Material.Payload) {
		return ErrBindingMismatch
	}
	auth := func(root, meta, key contracts.Digest) bool {
		// An older accepted snapshot keeps its original authorizers. Current
		// revocation is an operation-authorization condition, not a history eraser.
		// Unknown historical materials still owe L1/L2; only visible links can
		// be checked here.
		return meta != x.Metadata.Mark.Material.Payload || (root == x.Metadata.RootPayload && contains(x.Metadata.KeyIDs, key))
	}
	if d := x.Desired; d.Header.Coverage == Present {
		if d.ResourceVersion > x.Resources.Mark.Version || (d.ResourceVersion == x.Resources.Mark.Version) != (d.ResourcePayload == x.Resources.Mark.Material.Payload) || !auth(d.RootPayload, d.MetadataPayload, d.SigningKeyID) {
			return ErrBindingMismatch
		}
	}
	if r := x.Recovery; r.Header.Coverage == Present {
		if r.IdentityEpoch != x.Identity.IdentityEpoch || !auth(r.RootPayload, r.MetadataPayload, r.SigningKeyID) {
			return ErrBindingMismatch
		}
	}
	sup := x.Supervisor
	if x.Desired.Header.Coverage == NeverEstablished {
		if sup.Verified.Coverage != NeverEstablished || sup.Applied.Coverage != NeverEstablished || x.LKG.Header.Coverage != NeverEstablished || sup.Transition.Coverage != NeverEstablished {
			return ErrBindingMismatch
		}
	}
	if sup.Verified.Coverage == Present && !atMost(sup.Verified.Desired, desiredOf(x.Desired.Mark)) {
		return ErrBindingMismatch
	}
	if sup.Applied.Coverage == Present && (sup.Verified.Coverage != Present || !atMost(sup.Applied.Desired, sup.Verified.Desired)) {
		return ErrBindingMismatch
	}
	tr := sup.Transition
	if tr.Coverage == Present && (sup.Applied.Coverage != Present || sup.Verified.Coverage != Present || tr.PriorApplied != sup.Applied.Desired || tr.CandidateVerified != sup.Verified.Desired) {
		return ErrBindingMismatch
	}
	if x.LKG.Header.Coverage == Present {
		if !atMost(x.LKG.HighestAccepted, desiredOf(x.Desired.Mark)) {
			return ErrBindingMismatch
		}
		for _, e := range x.LKG.Entries {
			if sup.Applied.Coverage != Present || e.Desired != sup.Applied.Desired || !atMost(e.Desired, x.LKG.HighestAccepted) {
				return ErrBindingMismatch
			}
		}
	}
	return nil
}
