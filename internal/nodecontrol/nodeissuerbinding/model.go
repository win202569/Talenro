package nodeissuerbinding

import "talenro.local/platform/internal/nodecontrol/nodeleafbinding"

// Model owns one conditional evaluation. Shallow copies share the same owner:
// Clear through any copy invalidates all copies. Model does not support
// concurrent access, including Summary and Clear calls.

// Summary returns a detached snapshot, or zero after Clear. The snapshot does
// not authorize caching or reuse for a later observation.
func (m *Model) Summary() (out Summary) {
	if m == nil || m.owned == nil || !m.owned.valid {
		return Summary{}
	}
	out = m.owned.summary
	out.Evidence = nodeleafbinding.Summary{}
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

// Clear invalidates the shared owner and every shallow Model copy. It is
// nil-safe and idempotent, but must not run concurrently with other access.
func (m *Model) Clear() {
	if m != nil {
		m.owned.clear()
	}
}
