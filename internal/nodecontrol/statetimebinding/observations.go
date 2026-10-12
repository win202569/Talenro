package statetimebinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
)

func bindObservations(ctx context.Context, in Input, policy statepolicybinding.Summary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	// Keep cross-package status enums closed and separate: no casts.
	switch in.Policy.Policy.Time.Status {
	case statepolicy.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Metadata.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Time.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Request.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Checkpoint.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	old := in.Policy.Policy.Old
	if old == nil {
		return ErrUnsupportedPolicyShape
	}
	metadata := in.Time.Metadata
	request := in.Time.Request
	checkpoint := in.Time.Checkpoint
	if old.StateID == (contracts.Digest{}) || metadata.StateID != old.StateID || checkpoint.StateID != old.StateID || metadata.NodeID != old.NodeID || request.NodeID != metadata.NodeID || checkpoint.NodeID != metadata.NodeID || metadata.Epoch != old.Epoch || request.Epoch != metadata.Epoch || checkpoint.Epoch != metadata.Epoch {
		return ErrBindingMismatch
	}
	for _, scope := range [2]struct {
		node     string
		epoch    uint64
		provider contracts.Digest
	}{
		{in.Policy.Binding.Main.Previous.Scope.NodeID, in.Policy.Binding.Main.Previous.Scope.AuthorityEpoch, in.Policy.Binding.Main.Previous.Scope.TimeProviderID},
		{in.Policy.Binding.Main.Proposed.Scope.NodeID, in.Policy.Binding.Main.Proposed.Scope.AuthorityEpoch, in.Policy.Binding.Main.Proposed.Scope.TimeProviderID},
	} {
		if metadata.NodeID.String() != scope.node || metadata.Epoch != scope.epoch || in.Time.Time.ProviderID != scope.provider {
			return ErrBindingMismatch
		}
	}
	clock := in.Time.Time
	policyClock := in.Policy.Policy.Time
	if clock.ProviderID != old.TimeProviderID || clock.ProviderID != policyClock.ProviderID || clock.ObservationID != policyClock.ObservationID || clock.Now != policyClock.Now || clock.Floor != policyClock.Floor || policy.Time.ProviderID != clock.ProviderID || policy.Time.ObservationID != clock.ObservationID || policy.Time.Now != clock.Now || policy.Time.ObservedFloor != clock.Floor || policy.Time.OldFloor != old.TimeFloor || policy.Time.MainFloor != in.Policy.Binding.Main.Previous.Scope.TimeFloor || clock.Floor.Before(old.TimeFloor) || clock.Now.Before(clock.Floor) {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
