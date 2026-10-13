package nodeissuermapping

import (
	"context"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	"testing"
	"time"
)

func assertDeclaredOnly(t testing.TB, in Input) Summary {
	t.Helper()
	wantLower := assertIssuerMappingLower(t, in)
	m, e := CheckDeclared(context.Background(), in)
	if m != nil {
		defer m.Clear()
	}
	if e != nil || m == nil {
		t.Fatal(e)
	}
	s := m.Summary()
	if s.Proof != DeclaredIssuerSerialCorrespondenceConditionalTimeBound || s.Profile != "declared-issuer-serial.registry-version-link.v1" || !reflect.DeepEqual(s.Evidence, wantLower) || s.Proposed.Local != wantLower.Proposed || s.ObservedAt != wantLower.ObservedAt || s.ValidUntil != wantLower.ValidUntil || s.Obligations != ([5]Obligation{AuthenticatedRegistryPublicationAndRollbackProtection, AuthenticatedIssuanceLinkAndRowStatus, GuardedV7SnapshotAndLease, CrossRenewalDatabaseUniqueness, FullIdentityCompositionCommitAndRuntimeAuthorization}) {
		t.Fatal("declaration result")
	}
	return s
}
func TestIssuerMappingCheckDeclared(t *testing.T) {
	var signature func(context.Context, Input) (*Model, error) = CheckDeclared
	_ = signature
	in := issuerMappingFixture(t, "serial-one-high")
	before := issuerMappingJSONHash(t, in)
	s := assertDeclaredOnly(t, in)
	if s.Proposed != issuerMappingExpected(t, "serial-one-high").Proposed || issuerMappingJSONHash(t, in) != before {
		t.Fatal("frozen declaration projection")
	}
	if m, e := CheckDeclared(nil, in); m != nil || e != ErrInvalidInput {
		t.Fatal(e)
	}
}
func TestIssuerMappingTemporalInheritance(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	a, e := CheckDeclared(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	sa := a.Summary()
	in.Material.Time.Policy.Policy.Time.Now = in.Material.Time.Policy.Policy.Time.Now.Add(time.Nanosecond)
	in.Material.Time.Time.Time.Now = in.Material.Time.Time.Time.Now.Add(time.Nanosecond)
	b, e := CheckDeclared(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Clear()
	sb := b.Summary()
	if !sa.ObservedAt.Before(sb.ObservedAt) || sa.ValidUntil != sb.ValidUntil || !reflect.DeepEqual(a.Summary(), sa) {
		t.Fatal("new observation")
	}
	copied := *a
	a.Clear()
	if !reflect.DeepEqual(copied.Summary(), Summary{}) || !reflect.DeepEqual(b.Summary(), sb) {
		t.Fatal("independent owner")
	}
}
func TestIssuerMappingErrorPriority(t *testing.T) {
	var packets map[string]Input
	issuerMappingDecode(t, issuerMappingFrozenInputs, &packets)
	n := 0
	policyControls := 0
	for name, in := range packets {
		lower, want := nodeissuerbinding.Evaluate(context.Background(), in.Material)
		if lower != nil {
			lower.Clear()
		}
		if want == nil {
			continue
		}
		n++
		in.Claims.Catalog.Head.Version++
		m, got := CheckDeclared(context.Background(), in)
		if m != nil {
			m.Clear()
			t.Fatal("failure model")
		}
		if !issuerMappingErrorPreserved(got, want) {
			t.Fatal(name, got, want)
		}
		if name == "policy-failure" {
			if !issuerMappingPolicyFailure(got) || !issuerMappingPolicyFailure(want) {
				t.Fatal("missing structured double-failure control", got, want)
			}
			policyControls++
		}
	}
	if n == 0 || policyControls != 1 {
		t.Fatal("no double failures")
	}
	in := issuerMappingFixture(t, "serial-one-high")
	in.Material = nodeissuerbinding.Input{}
	in.Claims.Association.SerialBytes = make([]byte, 21)
	if m, e := CheckDeclared(context.Background(), in); m != nil || e != ErrLimitExceeded {
		t.Fatal("budget", e)
	}
	in.Claims.Association.SerialBytes = []byte{1}
	in.Claims.Association.IssuerID = "bad space"
	if m, e := CheckDeclared(context.Background(), in); m != nil || e != ErrInvalidInput {
		t.Fatal("shape", e)
	}
}
func TestIssuerMappingFabricatedClaimsAreNotTrusted(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	in.Claims.ExpectedHead.Deployment[0] = 99
	in.Claims.ExpectedHead.Digest[0] = 88
	in.Claims.ExpectedHead.Version = 123
	in.Claims.Catalog.Head = in.Claims.ExpectedHead
	in.Claims.Association.Head = in.Claims.ExpectedHead
	in.Claims.Catalog.Entries[0].IssuerID = "fabricated:publisher"
	in.Claims.Association.IssuerID = "fabricated:publisher"
	in.Claims.Catalog.Entries[0].CAVersion = 456
	in.Claims.Association.CAVersion = 456
	s := assertDeclaredOnly(t, in)
	if s.Proposed.Head != in.Claims.ExpectedHead {
		t.Fatal("declared coordinates")
	}
}
func TestIssuerMappingCoherentRollbackClaim(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	in.Claims.ExpectedHead.Version = 1
	in.Claims.Catalog.Head = in.Claims.ExpectedHead
	in.Claims.Association.Head = in.Claims.ExpectedHead
	assertDeclaredOnly(t, in) // Caller-coherent old heads are NOT rollback protection.
}
func TestIssuerMappingCoherentChangedCatalogClaim(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	r := in.Claims.Catalog.Entries[0]
	r.IssuerID = "unselected"
	in.Claims.Catalog.Entries = append(in.Claims.Catalog.Entries, r)
	before := assertDeclaredOnly(t, in)
	in.Claims.Catalog.Entries[1].CADERDigest[0] ^= 1
	after := assertDeclaredOnly(t, in)
	if before.Proposed != after.Proposed {
		t.Fatal("no registry wire/hash proof")
	}
}
func TestIssuerMappingCancellationAndPanic(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	before := issuerMappingJSONHash(t, in)
	probe := &ownershipFailureContext{Context: context.Background()}
	m, e := CheckDeclared(probe, in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for stop := 1; stop <= probe.calls; stop++ {
			failed, e := CheckDeclared(&ownershipFailureContext{Context: context.Background(), stop: stop, failure: failure}, in)
			if failed != nil || e != failure || issuerMappingJSONHash(t, in) != before {
				t.Fatal(stop, e)
			}
		}
	}
	for stop := 1; stop <= probe.calls; stop++ {
		func() {
			defer func() {
				if recover() != "ownership test panic" || issuerMappingJSONHash(t, in) != before {
					t.Fatal("panic", stop)
				}
			}()
			CheckDeclared(&ownershipFailureContext{Context: context.Background(), stop: stop, panicAt: true}, in)
		}()
	}
	t.Logf("enumerated %d real composed context checkpoints", probe.calls)
	// Summary owns its two slices, caller-owned declarations and material remain untouched.
	model, e := CheckDeclared(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer model.Clear()
	first, second := model.Summary(), model.Summary()
	reflect.ValueOf(first.Evidence.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage).Index(0).SetZero()
	if !reflect.DeepEqual(second, model.Summary()) || issuerMappingJSONHash(t, in) != before {
		t.Fatal("snapshot alias")
	}
	in.Claims.Association.IssuerID = strings.ToUpper(in.Claims.Association.IssuerID)
	if m, e := CheckDeclared(context.Background(), in); m != nil || e != ErrEvidenceInsufficient {
		t.Fatal("exact issuer case", e)
	}
}
