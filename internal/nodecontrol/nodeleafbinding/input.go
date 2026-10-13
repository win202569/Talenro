package nodeleafbinding

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"time"
)

const selectedLeafProfile = "main-selected-leaf.der-spki-window.v1"

type Input = nodecerttime.Input
type Proof uint8

const (
	None Proof = iota
	SelectedLeafProjectionConditionalTimeBound
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	IssuerAndSerialReferenceMapping
	KeyHandleAndPrivateKeyPossession
	IdentityEpochLineageHistoryAndRowStatus
	MainStateAndObservationProvenance
	FullIdentityCompositionCommitAndRuntimeAuthorization
)

// Projection contains only the proposed fields bound by the local profile.
// PublicKey identifies exact leaf SPKI DER, not an implicitly compatible
// database fingerprint, key handle, issuer ID or private-key credential.
type Projection struct {
	NodeID          string
	Leaf, PublicKey contracts.Digest
	Window          statecontract.TimeWindow
}

// Summary is conditional evidence for one observation, not authorization.
type Summary struct {
	Proof    Proof
	Evidence nodecerttime.Summary
	Proposed Projection
	// ObservedAt and ValidUntil inherit the lower conditional time interval.
	// They do not permit reuse for a new observation or runtime actions.
	ObservedAt, ValidUntil time.Time
	Obligations            [5]Obligation
}

// Model copies share one owner. Clearing any copy invalidates every copy.
// Concurrent access, including Summary and Clear, is unsupported.
type Model struct{ owned *owner }
