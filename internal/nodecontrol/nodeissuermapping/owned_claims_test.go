package nodeissuermapping

import (
	"context"
	"reflect"
	"testing"
)

func TestIssuerMappingOwnedClaims(t *testing.T) {
	assertOwnedUnit(t, ownedTestClaims(), ownClaims, clearClaims)
}
func TestIssuerMappingOwnedAliasIsolation(t *testing.T) {
	in := ownedTestClaims()
	in.Catalog.Entries = append(in.Catalog.Entries, in.Catalog.Entries[0])
	before := ownedTestClone(t, in)
	a, e := ownClaims(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer clearClaims(&a)
	b, e := ownClaims(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer clearClaims(&b)
	a.Catalog.Entries[0].IssuerID = "changed"
	a.Association.SerialBytes[0] = 9
	if !reflect.DeepEqual(in, before) || !reflect.DeepEqual(b, before) || a.Catalog.Entries[2].IssuerID != "x" {
		t.Fatal("alias")
	}
}
func TestIssuerMappingOwnedFailureEnumeration(t *testing.T) {
	in := ownedTestClaims()
	before := string(ownedTestJSON(t, in))
	probe := &ownershipFailureContext{Context: context.Background()}
	out, e := ownClaims(probe, in)
	if e != nil {
		t.Fatal(e)
	}
	clearClaims(&out)
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for stop := 1; stop <= probe.calls; stop++ {
			ctx := &ownershipFailureContext{Context: context.Background(), stop: stop, failure: failure}
			out, e := ownClaims(ctx, in)
			if e != failure || !reflect.DeepEqual(out, Claims{}) || string(ownedTestJSON(t, in)) != before {
				t.Fatal(stop, e)
			}
		}
	}
	for stop := 1; stop <= probe.calls; stop++ {
		func() {
			defer func() {
				if recover() != "ownership test panic" || string(ownedTestJSON(t, in)) != before {
					t.Fatal("panic", stop)
				}
			}()
			ownClaims(&ownershipFailureContext{Context: context.Background(), stop: stop, panicAt: true}, in)
		}()
	}
	t.Logf("ownership checkpoints %d", probe.calls)
}
