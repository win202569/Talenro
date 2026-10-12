package nodecerttime

import (
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"time"
)

type Input struct {
	Certificate nodeissuer.Input
	Time        statetimebinding.Input
}
type Proof uint8

const (
	None Proof = iota
	NodeCertificateConditionalTimeBound
)

// Summary reports conditional evidence for one evaluated observation, not a
// credential for later observations or runtime authorization.
type Summary struct {
	Proof       Proof
	Certificate nodeissuer.Summary
	Time        statetimebinding.Summary
	// ObservedAt is the caller-supplied conditional time, not trusted wall time.
	// ValidUntil is this observation's minimum time/CA/leaf deadline. It does
	// not permit caching or reuse for a new observation; Evaluate and applicable
	// state checks are still required for that observation.
	ObservedAt, ValidUntil time.Time
}
