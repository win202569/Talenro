package nodeleafbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"time"
)

func windowMatches(w statecontract.TimeWindow, start, end time.Time) bool {
	if w.Issued < -62135596800 || w.Issued > 253402300799 ||
		w.Until < -62135596800 || w.Until > 253402300799 || w.Issued >= w.Until {
		return false
	}
	return time.Unix(w.Issued, 0).UTC().Equal(start) && time.Unix(w.Until, 0).UTC().Equal(end)
}
func bindProjection(ctx context.Context, in Input, evidence nodecerttime.Summary) (Projection, error) {
	if e := stageError(ctx, nil); e != nil {
		return Projection{}, e
	}
	main := in.Time.Policy.Binding.Main.Proposed
	identity := main.Identity
	leaf := evidence.Certificate.Leaf.Leaf
	if main.Scope.NodeID != leaf.NodeID.String() ||
		identity.Header.Domain != statecontract.DomainIdentity ||
		identity.Header.Coverage != statecontract.Present ||
		identity.Leaf.Kind != statecontract.CertificateDER ||
		identity.Leaf.Payload != leaf.LeafDERDigest ||
		identity.Leaf.Envelope != (contracts.Digest{}) ||
		identity.PublicKey != leaf.SPKIDERDigest ||
		!windowMatches(identity.Window, leaf.NotBefore, leaf.NotAfter) {
		return Projection{}, stageError(ctx, ErrBindingMismatch)
	}
	if e := stageError(ctx, nil); e != nil {
		return Projection{}, e
	}
	return Projection{NodeID: main.Scope.NodeID, Leaf: identity.Leaf.Payload, PublicKey: identity.PublicKey, Window: identity.Window}, nil
}
