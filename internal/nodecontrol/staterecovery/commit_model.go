package staterecovery

import "talenro.local/platform/internal/nodecontrol/contracts"

type CommitModel struct {
	valid            bool
	initial          RecoveryInput
	base             RecoverySummary
	summary          CommitSummary
	recheckScanID    contracts.Digest
	eventIDs         [5]contracts.Digest
	eventCount       uint8
	acknowledged     bool
	interruptedPhase CommitPhase
}

func (m *CommitModel) Summary() CommitSummary {
	if m == nil || !m.valid {
		return CommitSummary{}
	}
	return m.summary
}
func (m *CommitModel) Clear() {
	if m != nil {
		*m = CommitModel{}
	}
}
