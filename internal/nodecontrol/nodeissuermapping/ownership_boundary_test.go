package nodeissuermapping

import "testing"

func TestIssuerMappingOwnershipBoundary(t *testing.T) {
	src := issuerMappingSources(t)
	if !issuerMappingOwnershipAllowed(src) {
		t.Fatal("actual protected lifecycle")
	}
	for _, c := range []struct{ file, old, replacement string }{
		{"check.go", "preflight(ctx,in)", "stageError(ctx,nil)"},
		{"check.go", "o.clear()", ""},
		{"check.go", "o:=&owner{}", "o:=&owner{}; ctx.Err()"},
		{"check.go", "if e:=ctx.Err();e!=nil", "if e:=error(nil);e!=nil"},
		{"check.go", "o.valid=true", "o.valid=true;ctx.Err()"},
		{"lower.go", "nodeissuerbinding.Evaluate(ctx,material)", "nodeissuerbinding.Evaluate(ctx,nodeissuerbinding.Input{})"},
		{"lower.go", "defer lower.Clear()", "lower.Clear()"},
		{"lower.go", "defer lower.Clear()", "ctx.Err();defer lower.Clear()"},
		{"lower.go", "o.summary.Evidence=lower.Summary()", "temporary:=lower.Summary();ctx.Err();o.summary.Evidence=temporary"},
		{"lower.go", "s.ObservedAt!=s.Evidence.ObservedAt", "false"},
		{"lower.go", "s.Proposed.Leaf!=s.Evidence.Proposed", "false"},
		{"lower.go", "leaf.SerialLength>20", "leaf.SerialLength>21"},
		{"owned_claims.go", "out.Catalog.Entries=nil", "out.Catalog.Entries=in.Catalog.Entries"},
		{"owned_claims.go", "clearClaims(&out)", ""},
		{"owned_claims.go", "copyRaw(in.Association.SerialBytes)", "in.Association.SerialBytes"},
		{"owned_summary.go", "clearEvidence(&out)", ""},
		{"owned_summary.go", "copySlice(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links)", "in.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links"},
		{"owned_summary.go", "copySlice(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage)", "in.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage"},
		{"model.go", "out.Evidence=nodeissuerbinding.Summary{}", "out.Evidence=m.owned.summary.Evidence"},
		{"owner.go", "clearClaims(&o.claims)", ""},
	} {
		if issuerMappingOwnershipAllowed(issuerMappingMutation(t, src, c.file, c.old, c.replacement)) {
			t.Fatal("ownership mutation", c.file, c.old)
		}
	}
}
func issuerMappingOwnershipAllowed(sources map[string][]byte) bool {
	return issuerMappingCapabilityAllowed(sources)
}
