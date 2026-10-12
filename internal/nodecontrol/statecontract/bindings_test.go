package statecontract

import "testing"

func TestContractBindings(t *testing.T) {
	if e := validateBindings(fixture()); e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(*MainStateContentModel){
		func(x *MainStateContentModel) { x.Checkpoint.AuthorityEpoch++ },
		func(x *MainStateContentModel) { x.Metadata.RootPayload = d(99) },
		func(x *MainStateContentModel) { x.Desired.ResourceVersion++ },
		func(x *MainStateContentModel) { x.Desired.ResourcePayload = d(99) },
		func(x *MainStateContentModel) { x.Desired.RootPayload = d(99) },
		func(x *MainStateContentModel) { x.Desired.SigningKeyID = d(99) },
		func(x *MainStateContentModel) { x.Recovery = recoveryFixture(); x.Recovery.IdentityEpoch++ },
		func(x *MainStateContentModel) { x.Supervisor.Guard.CounterID = d(99) },
		func(x *MainStateContentModel) { x.ServerTrust.Agent.AuthorityEpoch++ },
	} {
		x := fixture()
		edit(&x)
		if validateBindings(x) == nil {
			t.Fatal("mismatched binding accepted")
		}
	}
}
func TestContractAppliedAndLKG(t *testing.T) {
	for _, edit := range []func(*MainStateContentModel){
		func(x *MainStateContentModel) {
			x.Desired = DesiredSection{Header: SectionHeader{DomainDesired, NeverEstablished}}
		},
		func(x *MainStateContentModel) { x.Supervisor.Applied.Desired.Generation++ },
		func(x *MainStateContentModel) { x.Supervisor.Verified.Desired.Payload = d(99) },
		func(x *MainStateContentModel) { x.LKG.Entries[0].Desired.Payload = d(99) },
		func(x *MainStateContentModel) { x.LKG.HighestAccepted.Generation = 0 },
	} {
		x := fixture()
		edit(&x)
		if validateBindings(x) == nil {
			t.Fatal("inconsistent applied accepted")
		}
	}
	x := fixture()
	x.LKG.Entries = nil
	if e := validateBindings(x); e != nil {
		t.Fatal(e)
	}
	x = fixture()
	x.LKG.Entries[0].RootPayload = d(90)
	if e := validateBindings(x); e != nil {
		t.Fatal("historical authorizer rejected", e)
	}
	x = fixture()
	x.Desired = DesiredSection{Header: SectionHeader{DomainDesired, NeverEstablished}}
	x.Supervisor.Applied = AppliedRecord{Coverage: NeverEstablished}
	x.Supervisor.Verified = x.Supervisor.Applied
	x.LKG = LKGSection{Header: SectionHeader{DomainLKG, NeverEstablished}}
	if e := validateBindings(x); e != nil {
		t.Fatal(e)
	}
}
func TestContractIndependentCounters(t *testing.T) {
	x := fixture()
	x.Root.Mark.Sequence = 4
	x.Metadata.Mark.Sequence = 8
	x.Desired.Mark.Sequence = 20
	x.Supervisor.Applied.Desired.Sequence = 20
	x.Supervisor.Verified.Desired.Sequence = 20
	x.LKG.HighestAccepted.Sequence = 20
	x.LKG.Entries[0].Desired.Sequence = 20
	if e := validateBindings(x); e != nil {
		t.Fatal(e)
	}
	x.Identity.Serial = d(1)
	if e := validateShape(x); e != nil {
		t.Fatal(e)
	}
}
