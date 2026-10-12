package nodeleaf

import "bytes"

// Material holds only owned bytes and a partial-proof description.
// Copies of this wrapper share one lifecycle. It is not concurrency safe.
type Material struct{ owned *owner }

func (m *Material) Summary() Summary {
	if m == nil || m.owned == nil || !m.owned.valid {
		return Summary{}
	}
	return m.owned.summary
}
func (m *Material) DER() []byte {
	if m == nil || m.owned == nil || !m.owned.valid {
		return nil
	}
	return bytes.Clone(m.owned.der)
}
func (m *Material) Clear() {
	if m != nil {
		m.owned.clear()
	}
}
