package nodeleafkeys

// Material is a shared-lifecycle wrapper, not an authorization capability.
// Clear only erases reachable owned mutable buffers; parser/runtime copies,
// immutable strings and caller-owned data are outside this guarantee.
type Material struct{ owned *owner }

func (m *Material) Summary() Summary {
	if m == nil || m.owned == nil || !m.owned.valid {
		return Summary{}
	}
	return m.owned.summary
}
func (m *Material) Clear() {
	if m != nil {
		m.owned.clear()
	}
}
