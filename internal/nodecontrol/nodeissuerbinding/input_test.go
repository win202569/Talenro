package nodeissuerbinding

import (
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	"testing"
)

func TestIssuerBindingAPI(t *testing.T) {
	if reflect.TypeOf(Input{}) != reflect.TypeOf(nodeleafbinding.Input{}) {
		t.Fatal("Input alias")
	}
	if None != 0 || SelectedIssuerSerialReferencesConditionalTimeBound != 1 {
		t.Fatal("proof")
	}
	if reflect.TypeOf(None).Kind() != reflect.Uint8 || reflect.TypeOf(UnknownObligation).Kind() != reflect.Uint8 {
		t.Fatal("enum width")
	}
	if selectedIssuerSerialProfile != "main-selected-issuer-serial.ca-der-bound.v1" {
		t.Fatal("local profile")
	}
	for _, c := range []struct {
		v     any
		names []string
	}{
		{Projection{}, []string{"Leaf", "IssuerID", "Serial"}},
		{Summary{}, []string{"Proof", "Profile", "Evidence", "Proposed", "ObservedAt", "ValidUntil", "Obligations"}},
		{Model{}, []string{"owned"}},
		{owner{}, []string{"input", "summary", "valid"}},
	} {
		typ := reflect.TypeOf(c.v)
		if typ.NumField() != len(c.names) {
			t.Fatal("fields", typ)
		}
		for i, n := range c.names {
			if typ.Field(i).Name != n {
				t.Fatal(typ, n)
			}
		}
	}
	for i, v := range []Obligation{UnknownObligation, StoreIssuerMappingAndDeploymentProvenance, KeyHandleAndPrivateKeyPossession, IdentityEpochLineageHistoryAndRowStatus, MainStateAndObservationProvenance, FullIdentityCompositionCommitAndRuntimeAuthorization} {
		if int(v) != i {
			t.Fatal("obligation", i)
		}
	}
}
func TestIssuerBindingFixedErrors(t *testing.T) {
	fixed := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidEvidence, ErrBindingMismatch}
	texts := []string{"nodeissuerbinding: invalid input", "nodeissuerbinding: limit exceeded", "nodeissuerbinding: invalid evidence", "nodeissuerbinding: binding mismatch"}
	for i, e := range fixed {
		if e.Error() != texts[i] {
			t.Fatal("fixed error", i)
		}
		for j := 0; j < i; j++ {
			if fixed[j] == e {
				t.Fatal("aliased errors")
			}
		}
	}
}
