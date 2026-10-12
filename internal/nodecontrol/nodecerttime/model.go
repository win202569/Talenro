package nodecerttime

import "talenro.local/platform/internal/nodecontrol/statetimebinding"

type Model struct{ owned *owner }

func (m *Model) Summary() (out Summary) {
	if m == nil || m.owned == nil || !m.owned.valid {
		return Summary{}
	}
	out = m.owned.summary
	out.Time = statetimebinding.Summary{}
	done := false
	defer func() {
		if !done {
			clearTimeSummary(&out.Time)
			out = Summary{}
		}
	}()
	out.Time = ownTimeSummary(m.owned.summary.Time)
	done = true
	return out
}
func (m *Model) Clear() {
	if m != nil {
		m.owned.clear()
	}
}
