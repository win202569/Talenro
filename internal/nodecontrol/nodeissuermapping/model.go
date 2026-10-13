package nodeissuermapping

import "talenro.local/platform/internal/nodecontrol/nodeissuerbinding"

// Summary returns a detached conditional snapshot, never production authority.
func (m *Model) Summary() (out Summary) {
	if m == nil || m.owned == nil || !m.owned.valid {
		return Summary{}
	}
	out = m.owned.summary
	out.Evidence = nodeissuerbinding.Summary{}
	done := false
	defer func() {
		if !done {
			clearEvidence(&out.Evidence)
			out = Summary{}
		}
	}()
	out.Evidence = ownEvidence(m.owned.summary.Evidence)
	done = true
	return out
}

// Clear invalidates all shallow copies. It is nil-safe and idempotent.
func (m *Model) Clear() {
	if m != nil {
		m.owned.clear()
	}
}
