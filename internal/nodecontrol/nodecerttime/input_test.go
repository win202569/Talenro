package nodecerttime

import (
	"reflect"
	"testing"
)

// Detects a public input escape hatch or loss of full lower summaries.
func TestCertTimePublicContract(t *testing.T) {
	if None != 0 || NodeCertificateConditionalTimeBound != 1 || reflect.TypeOf(None).Kind() != reflect.Uint8 {
		t.Fatal("proof contract")
	}
	typ := reflect.TypeOf(Input{})
	if typ.NumField() != 2 || typ.Field(0).Name != "Certificate" || typ.Field(1).Name != "Time" {
		t.Fatal("input contract")
	}
	s := reflect.TypeOf(Summary{})
	names := []string{"Proof", "Certificate", "Time", "ObservedAt", "ValidUntil"}
	if s.NumField() != len(names) {
		t.Fatal("summary shape")
	}
	for i, n := range names {
		if s.Field(i).Name != n {
			t.Fatal("summary field", n)
		}
	}
}
