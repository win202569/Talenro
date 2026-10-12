package statepolicy

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

// EvaluateModel never calls observation providers or performs a state commit.
// Every positive result is conditional on the explicitly supplied model inputs.
func EvaluateModel(ctx context.Context, in ModelInput) (*Candidate, error) {
	if e := contextError(ctx); e != nil {
		return nil, e
	}
	if len(in.History) > 4096 {
		return nil, fail(AssuranceMissing, UnsupportedRange)
	}
	if in.Old != nil && (len(in.Old.Roots) > 33 || len(in.Old.Keys) > 4096 || len(in.Old.RevokedIDs) > 4096) {
		return nil, fail(AssuranceMissing, UnsupportedRange)
	}
	f, e := loadFacts(ctx, in.Batch)
	if e != nil {
		return nil, e
	}
	defer f.Clear()
	accepted, _, e := checkTransition(f, in)
	if e != nil {
		return nil, e
	}
	deadline, e := checkTime(f, in.Time, in.Old, accepted)
	if e != nil {
		return nil, e
	}
	keys, revoked, e := checkHistory(f, in)
	if e != nil {
		return nil, e
	}
	c := &Candidate{deadline: deadline, deployment: in.Deployment, time: in.Time}
	done := false
	defer func() {
		if !done {
			c.Clear()
		}
	}()
	// Take ownership immediately, before cloning further data.
	c.proposed = &StateModel{Keys: keys, RevokedIDs: revoked}
	s := c.proposed
	if in.Old != nil {
		s.StateID = in.Old.StateID
		s.CounterID = in.Old.CounterID
		s.StateEpoch = in.Old.StateEpoch
		s.TimeFloor = in.Old.TimeFloor
	} else {
		s.TimeFloor = in.Time.Floor
	}
	s.Status = Unknown // A proposal is not an observed committed state.
	s.NodeID = f.desired.bindings.NodeID
	s.Epoch = f.anchor.bindings.Epoch
	s.AnchorPayload = f.anchor.id.PayloadDigest
	s.DeploymentSourceID = in.Deployment.SourceID
	s.TimeProviderID = in.Time.ProviderID
	s.Metadata = HighWater{f.metadata.bindings.MetadataVersion, f.metadata.bindings.Sequence, f.metadata.id.PayloadDigest}
	s.Desired = HighWater{f.desired.bindings.Generation, f.desired.bindings.Sequence, f.desired.id.PayloadDigest}
	s.Roots = make([]RootRecord, len(f.rotations)+1)
	for i := range s.Roots {
		a := &f.anchor
		if i > 0 {
			a = &f.rotations[i-1]
		}
		r := &s.Roots[i]
		r.Mark = HighWater{a.root.Version, a.bindings.Sequence, a.id.PayloadDigest}
		r.Root.Version = a.root.Version
		r.Root.Threshold = a.root.Threshold
		r.Root.Keys = append([]stateformat.PublicKeyMaterial(nil), a.root.Keys...)
	}
	c.old = cloneState(in.Old)
	c.history = append([]KeyHistoryObservation(nil), in.History...)
	c.identities = make([]stateverify.ArtifactIdentity, 0, len(f.rotations)+3)
	c.identities = append(c.identities, f.anchor.id)
	for _, r := range f.rotations {
		c.identities = append(c.identities, r.id)
	}
	c.identities = append(c.identities, f.metadata.id, f.desired.id)
	c.noChange = sameBusiness(in.Old, s)
	if e = contextError(ctx); e != nil {
		return nil, e
	}
	c.ready = true
	done = true
	return c, nil
}
