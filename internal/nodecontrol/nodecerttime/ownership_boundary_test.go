package nodecerttime

import (
	"strings"
	"testing"
)

func TestCertTimeOwnershipBoundary(t *testing.T) {
	src := certTimeSources(t)
	if !certTimeOwnershipAllowed(src) {
		t.Fatal("actual protected lifecycle")
	}
	for _, tc := range []struct{ file, old, replacement string }{
		{"owner.go", "o.clear()", ""},
		{"owner.go", "o := &owner{}", "o := &owner{}; ctx.Err()"},
		{"evaluate.go", "if o != nil {", "if false {"},
		{"evaluate.go", "if o != nil {", "ctx.Err(); if o != nil {"},
		{"owned_certificate.go", "clearCertificate(&out)", ""},
		{"owned_binding.go", "clearBinding(&out)", ""},
		{"owned_content.go", "clearContent(&out)", ""},
		{"owned_table.go", "clearTable(&out)", ""},
		{"owned_policy.go", "clearPolicy(&out)", ""},
		{"owned_time.go", "clearTimeBinding(&out)", ""},
		{"owned_summary.go", "clearTimeSummary(&out)", ""},
		{"owned_table.go", "out.Materials[i].Raw = nil", "out.Materials[i].Raw = in.Materials[i].Raw"},
		{"owned_policy.go", "out.Old.Roots[i].Root.Keys = nil", "out.Old.Roots[i].Root.Keys = in.Old.Roots[i].Root.Keys"},
		{"lower.go", "defer certificate.Clear()", "certificate.Clear()"},
		{"lower.go", "defer timeModel.Clear()", "timeModel.Clear()"},
		{"lower.go", "o.summary.Time = timeModel.Summary()", "temporary := timeModel.Summary(); ctx.Err(); o.summary.Time = temporary"},
		{"model.go", "out.Time = statetimebinding.Summary{}", "out.Time = m.owned.summary.Time"},
		{"owned_summary.go", "copySlice(in.Policy.Binding.Proposed.Target.Links)", "in.Policy.Binding.Proposed.Target.Links"},
		{"owned_summary.go", "copySlice(in.Policy.Binding.Proposed.Coverage)", "in.Policy.Binding.Proposed.Coverage"},
		{"owned_summary.go", "clear(in.Policy.Binding.Proposed.Coverage)", ""},
		{"owned_certificate.go", "clear(in.Leaf.DER)", ""},
		{"owned_table.go", "clear(in.Materials[i].Raw)", ""},
		{"owned_policy.go", "in.Old.Clear()", ""},
		{"evaluate.go", "if e = stageError(ctx, nil); e != nil {", "if false {"},
		{"evaluate.go", "o.valid = true", "o.valid = true; ctx.Err()"},
		{"evaluate.go", "return &Model{owned: o}, nil", "return &Model{owned: &owner{}}, nil"},
	} {
		if certTimeOwnershipAllowed(certTimeMutation(t, src, tc.file, tc.old, tc.replacement)) {
			t.Fatal("ownership mutation", tc.file, tc.old)
		}
	}
}

func certTimeOwnershipAllowed(sources map[string][]byte) bool {
	schemas := certTimeSchemas()
	for _, name := range strings.Fields("owner.go owned_certificate.go owned_binding.go owned_content.go owned_table.go owned_policy.go owned_time.go owned_summary.go model.go lower.go evaluate.go") {
		raw, ok := sources[name]
		if !ok || schemas[name] == "" || !certTimeProductAllowed(raw) || certTimeASTHash(raw) != schemas[name] {
			return false
		}
	}
	return true
}
