package statematerials

import "talenro.local/platform/internal/nodecontrol/contracts"

type MaterialKind uint8

const (
	Unknown MaterialKind = iota
	RootSet
	RootRotation
	Metadata
	Desired
)

type Scope struct {
	NodeID string
	Epoch  uint64
}
type MaterialRef struct {
	Kind              MaterialKind
	Payload, Envelope contracts.Digest
}
type Material struct {
	Ref MaterialRef
	Raw []byte
}
type Window struct{ Issued, Until int64 }
type RootExpectation struct {
	Version, Threshold uint64
	KeyIDs             []contracts.Digest
}
type RotationExpectation struct {
	PreviousVersion, NewVersion    uint64
	OldRootPayload, NewRootPayload contracts.Digest
	Window                         Window
}
type MetadataExpectation struct {
	Version, RootVersion uint64
	RootPayload          contracts.Digest
	KeyIDs, RevokedIDs   []contracts.Digest
	Window               Window
}
type DesiredExpectation struct {
	NodeID                                                      string
	Generation, InventoryVersion, ResourceVersion               uint64
	ResourcePayload, RootPayload, MetadataPayload, SigningKeyID contracts.Digest
	Window                                                      Window
	SlotIDs                                                     []string
}
type Expectation struct {
	Ref             MaterialRef
	Epoch, Sequence uint64
	Root            RootExpectation
	Rotation        RotationExpectation
	Metadata        MetadataExpectation
	Desired         DesiredExpectation
}
type Input struct {
	Scope     Scope
	Materials []Material
	Expected  []Expectation
}
type materialKey struct {
	kind    MaterialKind
	payload contracts.Digest
}

func key(r MaterialRef) materialKey { return materialKey{r.Kind, r.Payload} }
