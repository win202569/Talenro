package statecontract

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func rejectConflict(t *testing.T, in ModelInput) {
	t.Helper()
	m, e := EvaluateModel(context.Background(), in)
	if m != nil {
		m.Clear()
	}
	if m != nil || !errors.Is(e, ErrConflict) {
		t.Fatalf("want nil/Conflict; got model=%v error=%v", m != nil, e)
	}
}
func TestContractVisibleDesiredIdentities(t *testing.T) {
	in := pair()
	in.Previous.Desired.Mark.Version = 2
	in.Previous.Desired.Mark.Sequence = 2
	in.Proposed.Desired.Mark = in.Previous.Desired.Mark
	rejectConflict(t, in)
	in = pair()
	in.Previous.Desired.Mark = StreamMark{2, 2, ref(DesiredSnapshot, 41)}
	in.Proposed.Desired.Mark = StreamMark{3, 3, ref(DesiredSnapshot, 42)}
	in.Proposed.Supervisor.Verified.Desired = DesiredBinding{2, 2, d(99)}
	rejectConflict(t, in)
	in.Proposed.Supervisor.Verified.Desired = DesiredBinding{2, 2, d(41)}
	m, e := EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal("consistent cross-snapshot identity rejected", e)
	}
	m.Clear()
}
func dormant(x *MainStateContentModel) {
	x.Desired = DesiredSection{Header: SectionHeader{DomainDesired, NeverEstablished}}
	x.LKG = LKGSection{Header: SectionHeader{DomainLKG, NeverEstablished}}
	x.Supervisor.Applied = AppliedRecord{Coverage: NeverEstablished}
	x.Supervisor.Verified = x.Supervisor.Applied
}
func chainInput() ModelInput {
	in := pair()
	dormant(&in.Previous)
	dormant(&in.Proposed)
	in.Proposed.Root.Mark = StreamMark{4, 4, ref(RootPayload, 91)}
	in.Proposed.Metadata.RootVersion = 4
	in.Proposed.Metadata.RootPayload = d(91)
	in.Proposed.Metadata.Mark = StreamMark{2, 5, ref(Metadata, 70)}
	in.RootTransition = []RotationRef{{1, 2, 2, d(20), d(90), ref(RootRotation, 80)}, {2, 3, 3, d(90), d(20), ref(RootRotation, 82)}, {3, 4, 4, d(20), d(91), ref(RootRotation, 84)}}
	return in
}
func TestContractWholeChainIdentity(t *testing.T) {
	in := chainInput()
	rejectConflict(t, in)
	in.RootTransition[1].NewRootPayload = d(92)
	in.RootTransition[2].OldRootPayload = d(92)
	in.RootTransition[2].Material = in.RootTransition[0].Material
	rejectConflict(t, in)
	in.RootTransition[2].Material = ref(RootRotation, 84)
	m, e := EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
}
func TestContractHistoricalAuthorizers(t *testing.T) {
	in := pair()
	in.Previous.Recovery = recoveryFixture()
	in.Proposed.Recovery = recoveryFixture()
	in.Proposed.Metadata.Mark = StreamMark{2, 2, ref(Metadata, 70)}
	in.Proposed.Metadata.RevokedIDs = []contracts.Digest{d(2)}
	in.Proposed.Metadata.RevokedSetDigest = d(71)
	in.Proposed.Metadata.History = ref(MetadataHistory, 72)
	m, e := EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal("revocation blocked by retained snapshots", e)
	}
	if m.Summary().Outstanding[4] != RuntimeAuthorization {
		t.Fatal("history became authorization")
	}
	m.Clear()
	in = pair()
	in.Proposed.Resources.Mark = StreamMark{2, 2, ref(ResourceEnvelope, 73)}
	m, e = EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal("resource advance blocked", e)
	}
	m.Clear()
	in = pair()
	in.Proposed.Root.Mark = StreamMark{2, 2, ref(RootPayload, 90)}
	in.RootTransition = []RotationRef{{1, 2, 2, d(20), d(90), ref(RootRotation, 80)}}
	m, e = EvaluateModel(context.Background(), in)
	if e != nil {
		t.Fatal("root advance blocked by historical metadata", e)
	}
	m.Clear()
	in = pair()
	in.Proposed.Desired.MetadataPayload = d(99)
	rejectConflict(t, in)
	in = pair()
	in.Previous.Recovery = recoveryFixture()
	in.Proposed.Recovery = recoveryFixture()
	in.Proposed.Recovery.RootPayload = d(99)
	rejectConflict(t, in)
}
