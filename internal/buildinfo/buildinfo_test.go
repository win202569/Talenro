package buildinfo

import "testing"

func TestCurrentDefaultsAreStable(t *testing.T) {
	got := Current()
	if got.Version != "dev" || got.Commit != "unknown" || got.BuiltAt != "unknown" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}
