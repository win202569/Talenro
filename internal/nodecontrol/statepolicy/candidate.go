package statepolicy

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/stateverify"
	"time"
)

// Candidate is an owned conditional model, never production authorization.
type Candidate struct {
	ready, noChange bool
	proposed, old   *StateModel
	deployment      DeploymentObservation
	time            TimeObservation
	history         []KeyHistoryObservation
	identities      []stateverify.ArtifactIdentity
	deadline        time.Time
}

func cloneState(s *StateModel) *StateModel {
	if s == nil {
		return nil
	}
	c := *s
	c.Roots = nil
	c.Keys = nil
	c.RevokedIDs = nil
	done := false
	defer func() {
		if !done {
			c.Clear()
		}
	}()
	if s.Roots != nil {
		c.Roots = make([]RootRecord, len(s.Roots))
	}
	for i, r := range s.Roots {
		c.Roots[i].Mark = r.Mark
		c.Roots[i].Root.Version = r.Root.Version
		c.Roots[i].Root.Threshold = r.Root.Threshold
		if r.Root.Keys != nil {
			c.Roots[i].Root.Keys = make([]stateformat.PublicKeyMaterial, len(r.Root.Keys))
			copy(c.Roots[i].Root.Keys, r.Root.Keys)
		}
	}
	if s.Keys != nil {
		c.Keys = make([]KeyRecord, len(s.Keys))
		copy(c.Keys, s.Keys)
	}
	if s.RevokedIDs != nil {
		c.RevokedIDs = make([]contracts.Digest, len(s.RevokedIDs))
		copy(c.RevokedIDs, s.RevokedIDs)
	}
	done = true
	return &c
}
func sameBusiness(a, b *StateModel) bool {
	return a != nil && b != nil && a.NodeID == b.NodeID && a.Epoch == b.Epoch && a.AnchorPayload == b.AnchorPayload && a.Metadata == b.Metadata && a.Desired == b.Desired && reflect.DeepEqual(a.Roots, b.Roots) && reflect.DeepEqual(a.Keys, b.Keys) && equalIDs(a.RevokedIDs, b.RevokedIDs)
}
func equalIDs(a, b []contracts.Digest) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (c *Candidate) NoChange() bool { return c != nil && c.ready && c.noChange }
func (c *Candidate) State() *StateModel {
	if c == nil || !c.ready {
		return nil
	}
	return cloneState(c.proposed)
}
func (c *Candidate) ValidUntil() time.Time {
	if c == nil || !c.ready {
		return time.Time{}
	}
	return c.deadline
}
func (c *Candidate) Clear() {
	if c == nil {
		return
	}
	c.proposed.Clear()
	c.old.Clear()
	clear(c.history)
	clear(c.identities)
	*c = Candidate{}
}
func RecheckModel(ctx context.Context, c *Candidate, in ModelInput) error {
	if e := contextError(ctx); e != nil {
		return e
	}
	if c == nil || !c.ready || !reflect.DeepEqual(c.old, in.Old) {
		return fail(Rejected, StaleCandidate)
	}
	if in.Time.ProviderID != c.time.ProviderID || in.Time.Now.Before(c.time.Now) || in.Time.Floor.Before(c.time.Floor) || !in.Time.Now.Before(c.deadline) {
		return fail(Rejected, StaleCandidate)
	}
	fresh, e := EvaluateModel(ctx, in)
	if e != nil {
		return e
	}
	defer fresh.Clear()
	if fresh.deployment != c.deployment || !reflect.DeepEqual(fresh.identities, c.identities) || !reflect.DeepEqual(fresh.history, c.history) || !sameBusiness(fresh.proposed, c.proposed) || !fresh.deadline.Equal(c.deadline) {
		return fail(Rejected, StaleCandidate)
	}
	return contextError(ctx)
}
