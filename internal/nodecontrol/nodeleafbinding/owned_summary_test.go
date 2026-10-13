package nodeleafbinding

import (
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"testing"
	"time"
)

func TestLeafBindingOwnedSummary(t *testing.T) {
	var in nodecerttime.Summary
	fillOwnedTest(reflect.ValueOf(&in).Elem(), time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC))
	expected := ownedTestClone(t, in)
	out := ownEvidence(in)
	if !reflect.DeepEqual(out, expected) {
		t.Fatal("lost responsibilities")
	}
	views := ownedTestViews(reflect.ValueOf(out))
	clearEvidence(&out)
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(in, expected) || !reflect.DeepEqual(out, nodecerttime.Summary{}) {
		t.Fatal("summary clear")
	}
}
