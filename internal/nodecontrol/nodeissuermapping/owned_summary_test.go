package nodeissuermapping

import (
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	"testing"
	"time"
)

func TestIssuerMappingOwnedSummary(t *testing.T) {
	var in nodeissuerbinding.Summary
	fillOwnedTest(reflect.ValueOf(&in).Elem(), time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC))
	expected := ownedTestClone(t, in)
	out := ownEvidence(in)
	if !reflect.DeepEqual(out, expected) {
		t.Fatal("complete")
	}
	views := ownedTestViews(reflect.ValueOf(out))
	clearEvidence(&out)
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(in, expected) || !reflect.DeepEqual(out, nodeissuerbinding.Summary{}) {
		t.Fatal("clear")
	}
	clearEvidence(nil)
}
