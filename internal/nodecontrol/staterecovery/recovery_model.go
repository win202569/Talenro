package staterecovery

type RecoveryModel struct {
	valid    bool
	snapshot RecoveryInput
	summary  RecoverySummary
}

func (m *RecoveryModel) Summary() RecoverySummary {
	if m == nil || !m.valid {
		return RecoverySummary{}
	}
	return m.summary
}
func (m *RecoveryModel) Clear() {
	if m != nil {
		*m = RecoveryModel{}
	}
}
