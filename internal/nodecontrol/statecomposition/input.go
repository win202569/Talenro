package statecomposition

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

type AnchorInput struct {
	Scope             statematerials.Scope
	Payload, Envelope contracts.Digest
	Raw               []byte
}
type RotationLink struct{ Rotation, NewRoot statematerials.MaterialRef }
type Target struct {
	AnchorRoot        statematerials.MaterialRef
	Links             []RotationLink
	Metadata, Desired statematerials.MaterialRef
}
type Input struct {
	Materials statematerials.Input
	Anchor    AnchorInput
	Target    Target
}
