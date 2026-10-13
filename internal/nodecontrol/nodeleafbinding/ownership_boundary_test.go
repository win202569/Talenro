package nodeleafbinding

import "testing"

// Exact normalized schemas lock this reviewed revision, not all future Go.
func leafBindingOwnershipAllowed(sources map[string][]byte) bool {
	return leafBindingCapabilityAllowed(sources)
}

func TestLeafBindingOwnershipBoundary(t *testing.T) {
	src := leafBindingSources(t)
	if !leafBindingOwnershipAllowed(src) {
		t.Fatal("actual protected lifecycle")
	}
	for _, c := range []struct{ file, old, replacement string }{
		{"owner.go", "preflight(ctx, in)", "stageError(ctx, nil)"},
		{"owner.go", "o.clear()", ""},
		{"owner.go", "o := &owner{}", "o := &owner{}; ctx.Err()"},
		{"owner.go", "ownCertificate(ctx, in.Certificate)", "ownCertificate(ctx, nodeissuer.Input{})"},
		{"owner.go", "ownTimeBinding(ctx, in.Time)", "ownTimeBinding(ctx, statetimebinding.Input{})"},
		{"evaluate.go", "if result == nil {", "if false {"},
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
		{"model.go", "out.Evidence = nodecerttime.Summary{}", "out.Evidence = m.owned.summary.Evidence"},
		{"model.go", "clearEvidence(&out.Evidence)", ""},
		{"owned_summary.go", "copySlice(in.Time.Policy.Binding.Proposed.Target.Links)", "in.Time.Policy.Binding.Proposed.Target.Links"},
		{"owned_summary.go", "copySlice(in.Time.Policy.Binding.Proposed.Coverage)", "in.Time.Policy.Binding.Proposed.Coverage"},
		{"owned_summary.go", "clear(in.Time.Policy.Binding.Proposed.Coverage)", ""},
		{"owned_certificate.go", "clear(in.Leaf.DER)", ""},
		{"owned_table.go", "clear(in.Materials[i].Raw)", ""},
		{"owned_policy.go", "in.Old.Clear()", ""},
		{"evaluate.go", "if err = stageError(ctx, nil); err != nil {", "if false {"},
		{"evaluate.go", "o.valid = true", "o.valid = true; ctx.Err()"},
		{"evaluate.go", "return &Model{owned: o}, nil", "return &Model{owned: &owner{}}, nil"},
		{"projection.go", "main.Scope.NodeID != leaf.NodeID.String()", "false"},
		{"projection.go", "identity.Header.Domain != statecontract.DomainIdentity", "false"},
		{"projection.go", "identity.Header.Coverage != statecontract.Present", "false"},
		{"projection.go", "identity.Leaf.Kind != statecontract.CertificateDER", "false"},
		{"projection.go", "identity.Leaf.Payload != leaf.LeafDERDigest", "false"},
		{"projection.go", "identity.Leaf.Envelope != (contracts.Digest{})", "false"},
		{"projection.go", "identity.PublicKey != leaf.SPKIDERDigest", "false"},
		{"projection.go", "!windowMatches(identity.Window, leaf.NotBefore, leaf.NotAfter)", "false"},
		{"projection.go", "time.Unix(w.Issued, 0).UTC().Equal(start)", "w.Issued == start.Unix()"},
		{"projection.go", "w.Issued < -62135596800", "false"},
		{"projection.go", "w.Issued >= w.Until", "false"},
		{"projection.go", "return time.Unix(w.Issued", "unused := time.Unix(w.Issued, 0); _ = unused; return time.Unix(w.Issued"},
	} {
		if leafBindingOwnershipAllowed(leafBindingMutation(t, src, c.file, c.old, c.replacement)) {
			t.Fatal("ownership/flow mutation", c.file, c.old)
		}
	}
}
