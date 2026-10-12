package statecontract

import (
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func TestContractCoverage(t *testing.T) {
	if err := validateShape(fixture()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Checkpoint", "Root", "Metadata", "Identity", "ServerTrust", "Releases", "Resources", "HostPolicy", "Desired", "Recovery", "Supervisor", "LKG"} {
		t.Run(name, func(t *testing.T) {
			x := fixture()
			v := reflect.ValueOf(&x).Elem().FieldByName(name).FieldByName("Header")
			v.FieldByName("Coverage").SetUint(0)
			if !errors.Is(validateShape(x), ErrIncompleteCoverage) {
				t.Fatal("unknown accepted")
			}
			v.FieldByName("Coverage").SetUint(uint64(Present))
			v.FieldByName("Domain").SetUint(0)
			if !errors.Is(validateShape(x), ErrBindingMismatch) {
				t.Fatal("wrong tag accepted")
			}
		})
	}
	x := fixture()
	x.Desired = DesiredSection{Header: SectionHeader{DomainDesired, NeverEstablished}}
	if err := validateShape(x); err != nil {
		t.Fatal(err)
	}
	x.Desired.InventoryVersion = 1
	if validateShape(x) == nil {
		t.Fatal("residual accepted")
	}
	if validateShape(MainStateContentModel{}) == nil {
		t.Fatal("zero accepted")
	}
	for _, name := range []string{"Checkpoint", "Root", "Metadata", "Identity", "ServerTrust", "Releases", "Resources", "HostPolicy", "Supervisor"} {
		x := fixture()
		reflect.ValueOf(&x).Elem().FieldByName(name).FieldByName("Header").FieldByName("Coverage").SetUint(uint64(NeverEstablished))
		if !errors.Is(validateShape(x), ErrIncompleteCoverage) {
			t.Fatal("required domain absent", name)
		}
	}
}
func TestContractShapeBounds(t *testing.T) {
	cases := []struct {
		name string
		edit func(*MainStateContentModel)
	}{
		{"root empty", func(x *MainStateContentModel) { x.Root.KeyIDs = nil }},
		{"threshold", func(x *MainStateContentModel) { x.Root.Threshold = 2 }},
		{"zero threshold", func(x *MainStateContentModel) { x.Root.Threshold = 0 }},
		{"zero digest", func(x *MainStateContentModel) { x.Scope.DeploymentID = contracts.Digest{} }},
		{"uuid", func(x *MainStateContentModel) { x.Scope.NodeID = "12345678-1234-4234-8234-123456789ABC" }},
		{"non RFC4122 uuid", func(x *MainStateContentModel) { x.Scope.NodeID = "12345678-1234-4234-7234-123456789abc" }},
		{"epoch", func(x *MainStateContentModel) { x.Scope.AuthorityEpoch = 1 << 63 }},
		{"zero epoch", func(x *MainStateContentModel) { x.Scope.AuthorityEpoch = 0 }},
		{"wrong kind", func(x *MainStateContentModel) { x.Desired.Mark.Material.Kind = RecoverySnapshot }},
		{"missing envelope", func(x *MainStateContentModel) { x.Metadata.Mark.Material.Envelope = contracts.Digest{} }},
		{"extra envelope", func(x *MainStateContentModel) { x.Identity.Leaf.Envelope = d(3) }},
		{"time", func(x *MainStateContentModel) { x.Identity.Window = TimeWindow{10, 10} }},
		{"time range", func(x *MainStateContentModel) { x.Identity.LineageUntil = 253402300800 }},
		{"domain", func(x *MainStateContentModel) { x.ServerTrust.Agent.TrustDomain = "127.0.0.1" }},
		{"slot", func(x *MainStateContentModel) { x.Desired.SlotIDs = []string{"bad/slot"} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			x := fixture()
			c.edit(&x)
			if validateShape(x) == nil {
				t.Fatal("invalid shape accepted")
			}
		})
	}
	for _, tc := range []struct {
		max int
		set func(*MainStateContentModel, []contracts.Digest)
	}{{5, func(x *MainStateContentModel, v []contracts.Digest) { x.Root.KeyIDs = v }}, {64, func(x *MainStateContentModel, v []contracts.Digest) { x.Metadata.KeyIDs = v }}, {4096, func(x *MainStateContentModel, v []contracts.Digest) { x.Metadata.RevokedIDs = v }}, {128, func(x *MainStateContentModel, v []contracts.Digest) { x.ServerTrust.Agent.Deauthorized = v }}} {
		values := make([]contracts.Digest, tc.max+1)
		for i := range values {
			values[i][0] = byte((i + 1) >> 8)
			values[i][1] = byte(i + 1)
		}
		x := fixture()
		tc.set(&x, values[:tc.max])
		if err := validateShape(x); err != nil {
			t.Fatal(err)
		}
		tc.set(&x, values)
		if validateShape(x) == nil {
			t.Fatal("overflow accepted")
		}
		tc.set(&x, []contracts.Digest{d(2), d(1)})
		if validateShape(x) == nil {
			t.Fatal("unordered accepted")
		}
		tc.set(&x, []contracts.Digest{d(1), d(1)})
		if validateShape(x) == nil {
			t.Fatal("duplicate accepted")
		}
	}
	for _, n := range []int{8, 9} {
		x := fixture()
		x.Desired.SlotIDs = nil
		x.LKG.Entries = nil
		for i := 0; i < n; i++ {
			s := string(rune('a' + i))
			x.Desired.SlotIDs = append(x.Desired.SlotIDs, s)
			e := fixture().LKG.Entries[0]
			e.SlotID = s
			x.LKG.Entries = append(x.LKG.Entries, e)
		}
		if (validateShape(x) == nil) != (n == 8) {
			t.Fatal("slot limit", n)
		}
	}
	for _, n := range []int{64, 65} {
		x := fixture()
		x.Recovery = recoveryFixture()
		for i := 0; i < n; i++ {
			s := faultID(i)
			x.Recovery.Local = append(x.Recovery.Local, LocalFaultBinding{s, x.Scope.NodeID})
		}
		if (validateShape(x) == nil) != (n == 64) {
			t.Fatal("local limit", n)
		}
	}
	for _, n := range []int{16, 17} {
		x := fixture()
		x.Recovery = recoveryFixture()
		for i := 0; i < n; i++ {
			s := faultID(i)
			x.Recovery.Supervisor = append(x.Recovery.Supervisor, SupervisorFaultBinding{s, x.Scope.NodeID, d(1)})
		}
		if (validateShape(x) == nil) != (n == 16) {
			t.Fatal("supervisor limit", n)
		}
	}
}
func TestContractGuardSeparation(t *testing.T) {
	for a := 0; a < 3; a++ {
		for b := a + 1; b < 3; b++ {
			for _, key := range []bool{false, true} {
				x := fixture()
				g := []*GuardBinding{&x.Scope.Main, &x.Scope.Supervisor, &x.Scope.Latch}
				if key {
					g[b].SealKeyID = g[a].SealKeyID
				} else {
					g[b].CounterID = g[a].CounterID
				}
				if validateShape(x) == nil {
					t.Fatal("guard alias accepted")
				}
			}
		}
	}
}

// Every persisted time field shares the documented range, including expired
// history. These are coverage assertions for already-correct shape checks.
func TestContractEveryTimeFieldBounds(t *testing.T) {
	fields := []func(*MainStateContentModel) *int64{
		func(x *MainStateContentModel) *int64 { return &x.Scope.TimeFloor.UnixSeconds },
		func(x *MainStateContentModel) *int64 { return &x.Metadata.Window.Issued },
		func(x *MainStateContentModel) *int64 { return &x.Metadata.Window.Until },
		func(x *MainStateContentModel) *int64 { return &x.Identity.LineageUntil },
		func(x *MainStateContentModel) *int64 { return &x.Identity.Window.Issued },
		func(x *MainStateContentModel) *int64 { return &x.Identity.Window.Until },
		func(x *MainStateContentModel) *int64 { return &x.Desired.Window.Issued },
		func(x *MainStateContentModel) *int64 { return &x.Desired.Window.Until },
		func(x *MainStateContentModel) *int64 { return &x.Recovery.Window.Issued },
		func(x *MainStateContentModel) *int64 { return &x.Recovery.Window.Until },
		func(x *MainStateContentModel) *int64 { return &x.LKG.Entries[0].Deadline },
	}
	for i, field := range fields {
		for _, n := range []int64{-62135596801, 253402300800} {
			x := fixture()
			x.Scope.TimeFloor.Present = true
			x.Recovery = recoveryFixture()
			*field(&x) = n
			if !errors.Is(validateShape(x), ErrInvalidInput) {
				t.Fatalf("field %d accepted out-of-range time", i)
			}
		}
	}
	x := fixture()
	x.Recovery = recoveryFixture()
	w := TimeWindow{-62135596800, 253402300799}
	x.Metadata.Window = w
	x.Identity.Window = w
	x.Desired.Window = w
	x.Recovery.Window = w
	x.Identity.LineageUntil = w.Until
	x.LKG.Entries[0].Deadline = w.Issued
	x.Scope.TimeFloor = OptionalTime{true, w.Issued}
	if e := validateShape(x); e != nil {
		t.Fatal("inclusive time endpoints", e)
	}
}
