package nodeleafbinding

import (
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"testing"
)

func TestLeafBindingAPI(t *testing.T) {
	if reflect.TypeOf(Input{}) != reflect.TypeOf(nodecerttime.Input{}) {
		t.Fatal("Input is not an alias")
	}
	if None != 0 || SelectedLeafProjectionConditionalTimeBound != 1 {
		t.Fatal("proof contract")
	}
	if selectedLeafProfile != "main-selected-leaf.der-spki-window.v1" {
		t.Fatal("local profile")
	}
	for _, c := range []struct {
		v     any
		names []string
	}{
		{Projection{}, []string{"NodeID", "Leaf", "PublicKey", "Window"}},
		{Summary{}, []string{"Proof", "Evidence", "Proposed", "ObservedAt", "ValidUntil", "Obligations"}},
		{Model{}, []string{"owned"}},
	} {
		typ := reflect.TypeOf(c.v)
		if typ.NumField() != len(c.names) {
			t.Fatal("unexpected fields", typ)
		}
		for i, n := range c.names {
			if typ.Field(i).Name != n {
				t.Fatal(typ, n)
			}
		}
	}
	obligations := []Obligation{UnknownObligation, IssuerAndSerialReferenceMapping, KeyHandleAndPrivateKeyPossession, IdentityEpochLineageHistoryAndRowStatus, MainStateAndObservationProvenance, FullIdentityCompositionCommitAndRuntimeAuthorization}
	for i, v := range obligations {
		if int(v) != i {
			t.Fatal("obligation contract", i)
		}
	}
	fixed := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidEvidence, ErrBindingMismatch}
	texts := []string{"nodeleafbinding: invalid input", "nodeleafbinding: limit exceeded", "nodeleafbinding: invalid evidence", "nodeleafbinding: binding mismatch"}
	for i, e := range fixed {
		if e.Error() != texts[i] {
			t.Fatal(e)
		}
		for j := 0; j < i; j++ {
			if e == fixed[j] {
				t.Fatal("aliased errors")
			}
		}
	}
}
