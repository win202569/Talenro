package nodeissuer

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
