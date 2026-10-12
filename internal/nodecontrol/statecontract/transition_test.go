package statecontract

import (
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func pair() ModelInput { return ModelInput{Previous: fixture(), Proposed: fixture()} }
func TestContractCoverageRegression(t *testing.T) {
	for _, name := range []string{"desired", "recovery", "lkg", "verified", "applied"} {
		in := pair()
		switch name {
		case "desired":
			in.Proposed.Desired.Header.Coverage = NeverEstablished
		case "recovery":
			in.Previous.Recovery = recoveryFixture()
		case "lkg":
			in.Proposed.LKG.Header.Coverage = NeverEstablished
		case "verified":
			in.Proposed.Supervisor.Verified.Coverage = NeverEstablished
		case "applied":
			in.Proposed.Supervisor.Applied.Coverage = NeverEstablished
		}
		if !errors.Is(validateTransition(in), ErrRollback) {
			t.Fatal(name)
		}
	}
}
func TestContractLKGHistory(t *testing.T) {
	in := pair()
	in.Proposed.LKG.Entries = nil
	if e := validateTransition(in); e != nil {
		t.Fatal(e)
	}
	in.Previous.Metadata.RevokedIDs = []contracts.Digest{d(7)}
	if !errors.Is(validateTransition(in), ErrRollback) {
		t.Fatal("revocation deletion accepted")
	}
	in = pair()
	in.Previous.ServerTrust.Agent.Deauthorized = []contracts.Digest{d(7)}
	if !errors.Is(validateTransition(in), ErrRollback) {
		t.Fatal("CA deletion accepted")
	}
	in = pair()
	in.Proposed.Metadata.RevokedIDs = []contracts.Digest{d(7)}
	if e := validateTransition(in); e != nil {
		t.Fatal(e)
	}
}
func marks(x *MainStateContentModel) []*StreamMark {
	return []*StreamMark{&x.Root.Mark, &x.Metadata.Mark, &x.Resources.Mark, &x.HostPolicy.Mark, &x.ServerTrust.Bootstrap.Mark, &x.ServerTrust.Agent.Mark, &x.Desired.Mark, &x.Recovery.Mark}
}
func TestContractForks(t *testing.T) {
	for i := 0; i < 8; i++ {
		for mode := 0; mode < 4; mode++ {
			in := pair()
			in.Previous.Recovery = recoveryFixture()
			in.Proposed.Recovery = recoveryFixture()
			a, b := marks(&in.Previous)[i], marks(&in.Proposed)[i]
			a.Version = 2
			a.Sequence = 2
			b.Version = 2
			b.Sequence = 2
			switch mode {
			case 0:
				b.Version = 1
			case 1:
				b.Sequence = 1
			case 2:
				b.Material.Payload = d(99)
			case 3:
				b.Version = 3
				b.Sequence = 3
			}
			if validateTransition(in) == nil {
				t.Fatalf("mark %d mode %d", i, mode)
			}
		}
	}
	for _, which := range []bool{false, true} {
		for mode := 0; mode < 3; mode++ {
			in := pair()
			a, b := &in.Previous.Releases.Manifest, &in.Proposed.Releases.Manifest
			if which {
				a, b = &in.Previous.Releases.Map, &in.Proposed.Releases.Map
			}
			a.Version = 2
			b.Version = 2
			switch mode {
			case 0:
				b.Version = 1
			case 1:
				b.Material.Payload = d(99)
			case 2:
				b.Version = 3
			}
			if validateTransition(in) == nil {
				t.Fatal("version mark accepted")
			}
		}
	}
	in := pair()
	in.Proposed.Metadata.Mark.Material.Envelope = d(99)
	if e := validateTransition(in); e != nil {
		t.Fatal("envelope renewal treated as fork", e)
	}
	in = pair()
	in.Proposed.Identity.Serial = d(99)
	if !errors.Is(validateTransition(in), ErrConflict) {
		t.Fatal("same leaf conflicting serial")
	}
}
func rootPair() ModelInput {
	in := pair()
	in.Previous.Root.Mark.Version = 100
	in.Proposed.Root.Mark = StreamMark{101, 2, ref(RootPayload, 90)}
	in.RootTransition = []RotationRef{{100, 101, 2, d(20), d(90), ref(RootRotation, 80)}}
	return in
}
func TestContractRootTransition(t *testing.T) {
	if e := validateTransition(rootPair()); e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(*ModelInput){func(x *ModelInput) { x.RootTransition = nil }, func(x *ModelInput) { x.RootTransition[0].NewVersion++ }, func(x *ModelInput) { x.RootTransition[0].OldRootPayload = d(99) }, func(x *ModelInput) { x.RootTransition[0].Sequence = 1 }, func(x *ModelInput) { x.RootTransition = make([]RotationRef, 33) }, func(x *ModelInput) { x.RootTransition[0].NewVersion = 1 << 63 }} {
		in := rootPair()
		edit(&in)
		if validateTransition(in) == nil {
			t.Fatal("bad chain accepted")
		}
	}
	in := pair()
	in.RootTransition = rootPair().RootTransition
	if validateTransition(in) == nil {
		t.Fatal("unchanged root chain accepted")
	}
}
func TestContractTransitionScope(t *testing.T) {
	for _, edit := range []func(*MainStateContentModel){func(x *MainStateContentModel) { x.Scope.AuthorityEpoch++ }, func(x *MainStateContentModel) { x.Scope.DeploymentID = d(99) }, func(x *MainStateContentModel) { x.Scope.Main.CounterID = d(99) }, func(x *MainStateContentModel) { x.Scope.TimeProviderID = d(99) }, func(x *MainStateContentModel) { x.Identity.IdentityEpoch++ }, func(x *MainStateContentModel) { x.ServerTrust.Agent.TrustDomain = "other.test" }} {
		in := pair()
		edit(&in.Proposed)
		if !errors.Is(validateTransition(in), ErrBindingMismatch) {
			t.Fatal("scope replacement accepted")
		}
	}
	in := pair()
	in.Previous.Scope.TimeFloor = OptionalTime{true, 10}
	if !errors.Is(validateTransition(in), ErrRollback) {
		t.Fatal("floor deletion")
	}
	in.Proposed.Scope.TimeFloor = OptionalTime{true, 9}
	if !errors.Is(validateTransition(in), ErrRollback) {
		t.Fatal("floor lowering")
	}
	in = pair()
	in.Proposed.Identity.Serial = d(1)
	in.Proposed.Identity.Leaf.Payload = d(99)
	if e := validateTransition(in); e != nil {
		t.Fatal("serial treated as counter", e)
	}
	in = pair()
	in.Previous.Recovery = recoveryFixture()
	in.Proposed.Recovery = recoveryFixture()
	in.Previous.Recovery.Mark.Version = 2
	in.Proposed.Recovery.RecoveryID = faultID(3)
	if !errors.Is(validateTransition(in), ErrRollback) {
		t.Fatal("recovery reset accepted")
	}
}
