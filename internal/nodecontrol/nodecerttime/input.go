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

type Summary struct {
	Proof                  Proof
	Certificate            nodeissuer.Summary
	Time                   statetimebinding.Summary
	ObservedAt, ValidUntil time.Time
}
