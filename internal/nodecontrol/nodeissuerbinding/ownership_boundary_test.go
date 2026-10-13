package nodeissuerbinding

import "testing"

func issuerBindingOwnershipAllowed(sources map[string][]byte) bool {
	return issuerBindingCapabilityAllowed(sources)
}

func TestIssuerBindingOwnershipBoundary(t *testing.T) {
	src := issuerBindingSources(t)
	if !issuerBindingOwnershipAllowed(src) {
		t.Fatal("actual protected lifecycle")
	}
	for _, c := range []struct{ file, old, replacement string }{
		{"owner.go", "preflight(ctx, in)", "stageError(ctx, nil)"},
		{"owner.go", "o.clear()", ""},
		{"owner.go", "o := &owner{}", "o := &owner{}; ctx.Err()"},
		{"owner.go", "ownCertificate(ctx, in.Certificate)", "ownCertificate(ctx, nodeissuer.Input{})"},
		{"owner.go", "ownTimeBinding(ctx, in.Time)", "ownTimeBinding(ctx, statetimebinding.Input{})"},
		{"evaluate.go", "if !transferred {", "if false {"},
		{"owned_certificate.go", "clearCertificate(&out)", ""},
		{"owned_binding.go", "clearBinding(&out)", ""},
		{"owned_content.go", "clearContent(&out)", ""},
		{"owned_table.go", "clearTable(&out)", ""},
		{"owned_policy.go", "clearPolicy(&out)", ""},
		{"owned_time.go", "clearTimeBinding(&out)", ""},
		{"owned_summary.go", "clearEvidence(&out)", ""},
		{"owned_table.go", "out.Materials[i].Raw = nil", "out.Materials[i].Raw = in.Materials[i].Raw"},
		{"owned_policy.go", "out.Old.Roots[i].Root.Keys = nil", "out.Old.Roots[i].Root.Keys = in.Old.Roots[i].Root.Keys"},
		{"lower.go", "defer lower.Clear()", "lower.Clear()"},
		{"lower.go", "defer lower.Clear()", "ctx.Err(); defer lower.Clear()"},
		{"lower.go", "o.summary.Evidence = lower.Summary()", "temporary := lower.Summary(); ctx.Err(); o.summary.Evidence = temporary"},
		{"lower.go", "s.Proposed != expected", "false"},
		{"lower.go", "s.ObservedAt != s.Evidence.ObservedAt", "false"},
		{"lower.go", "s.ValidUntil != s.Evidence.ValidUntil", "false"},
		{"model.go", "out.Evidence = nodeleafbinding.Summary{}", "out.Evidence = m.owned.summary.Evidence"},
		{"model.go", "clearEvidence(&out.Evidence)", ""},
		{"owned_summary.go", "copySlice(in.Evidence.Time.Policy.Binding.Proposed.Target.Links)", "in.Evidence.Time.Policy.Binding.Proposed.Target.Links"},
		{"owned_summary.go", "copySlice(in.Evidence.Time.Policy.Binding.Proposed.Coverage)", "in.Evidence.Time.Policy.Binding.Proposed.Coverage"},
		{"owned_summary.go", "clear(in.Evidence.Time.Policy.Binding.Proposed.Coverage)", ""},
		{"owned_certificate.go", "clear(in.Leaf.DER)", ""},
		{"owned_table.go", "clear(in.Materials[i].Raw)", ""},
		{"owned_policy.go", "in.Old.Clear()", ""},
		{"evaluate.go", "if e = stageError(ctx, nil); e != nil {", "if false {"},
		{"evaluate.go", "o.valid = true", "o.valid = true; ctx.Err()"},
		{"evaluate.go", "return &Model{owned: o}, nil", "return &Model{owned: &owner{}}, nil"},
		{"evaluate.go", "o.summary.Profile = selectedIssuerSerialProfile", "o.summary.Profile = \"other\""},
		{"evaluate.go", "o.summary.Proof = SelectedIssuerSerialReferencesConditionalTimeBound", "o.summary.Proof = None"},
		{"projection.go", "identity.IssuerID != issuer", "false"},
		{"projection.go", "identity.Serial != serial", "false"},
		{"references.go", "defer s.clear()", "s.clear()"},
		{"references.go", "s == nil", "false"},
		{"references.go", "length > 20", "length > 21"},
		{"references.go", "serial[0] == 0", "false"},
		{"references.go", "serial[length:]", "serial[20:]"},
		{"references.go", "e.Evidence.Certificate.CA.CADERDigest", "e.Evidence.Certificate.Leaf.Leaf.LeafDERDigest"},
		{"references.go", "issuer-ref.v1", "issuer-ref.v2"},
		{"references.go", "serial-ref.v1", "issuer-ref.v1"},
		{"references.go", "\\x00", "x00"},
		{"references.go", "sha256.Sum256(s.issuer[:])", "sha256.Sum256(s.serial[:])"},
		{"references.go", "sha256.Sum256(s.serial[:69+int(length)])", "sha256.Sum256(s.serial[:])"},
	} {
		if issuerBindingOwnershipAllowed(issuerBindingMutation(t, src, c.file, c.old, c.replacement)) {
			t.Fatal("ownership/flow mutation", c.file, c.old)
		}
	}
}
