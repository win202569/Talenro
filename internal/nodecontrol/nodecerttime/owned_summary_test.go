package nodecerttime

import (
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"testing"
	"time"
)

func TestCertTimeOwnedSummary(t *testing.T) {
	var in statetimebinding.Summary
	fillOwnedTest(reflect.ValueOf(&in).Elem(), time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC))
	expected := ownedTestClone(t, in)
	out := ownTimeSummary(in)
	if !reflect.DeepEqual(out, expected) {
		t.Fatal("lost responsibilities")
	}
	views := ownedTestViews(reflect.ValueOf(out))
	clearTimeSummary(&out)
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(in, expected) || !reflect.DeepEqual(out, statetimebinding.Summary{}) {
		t.Fatal("summary clear")
	}
}
