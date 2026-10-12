package nodecerttime

import "context"

func bindSnapshots(ctx context.Context, in Input, s Summary) error {
	if e := stageError(ctx, nil); e != nil {
		return e
	}
	node := s.Time.Time.NodeID
	if s.Certificate.Leaf.Leaf.NodeID != in.Certificate.Leaf.NodeID || in.Certificate.Leaf.NodeID != node ||
		node != in.Time.Time.Metadata.NodeID || node != in.Time.Time.Request.NodeID || node != in.Time.Time.Checkpoint.NodeID ||
		node.String() != in.Time.Policy.Binding.Main.Previous.Scope.NodeID || node.String() != in.Time.Policy.Binding.Main.Proposed.Scope.NodeID {
		return stageError(ctx, ErrBindingMismatch)
	}
	tuple := s.Time.Policy.Time
	attested, policy := in.Time.Time.Time, in.Time.Policy.Policy.Time
	if tuple.ProviderID != attested.ProviderID || tuple.ProviderID != policy.ProviderID ||
		tuple.ObservationID != attested.ObservationID || tuple.ObservationID != policy.ObservationID ||
		tuple.Now != attested.Now || tuple.Now != policy.Now ||
		tuple.ObservedFloor != attested.Floor || tuple.ObservedFloor != policy.Floor {
		return stageError(ctx, ErrBindingMismatch)
	}
	return stageError(ctx, nil)
}
