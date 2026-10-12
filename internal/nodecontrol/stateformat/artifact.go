package stateformat

import (
	"bytes"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type Kind uint8

const (
	Invalid Kind = iota
	Desired
	InitialRoot
	RootRotation
	Metadata
	TimeAttestation
)

// Bindings contains only parsed format coordinates, never verified authority.
type Bindings struct {
	CheckpointSequence                                                 uint64
	RequestNonce                                                       [32]byte
	Kind                                                               Kind
	Epoch, Sequence, RootVersion, PreviousRootVersion, MetadataVersion uint64
	NodeID                                                             uuid.UUID
	Generation, InventoryVersion, ResourceEnvelopeVersion              uint64
	ResourceEnvelopeDigest, SigningKeyID                               contracts.Digest
	Signature                                                          [64]byte
	IssuedAt, ValidUntil                                               time.Time
}
type artifactBindings = Bindings

// ParsedArtifact owns format evidence, not authorization. Do not copy the
// struct to create another owner, or use accessors concurrently with Clear.
type ParsedArtifact struct {
	embeddedRoot                  *ParsedRootSet
	slots                         []string
	root                          RootMaterial
	metadata                      MetadataMaterial
	signatures                    SignatureMaterial
	payload, envelope             []byte
	bindings                      Bindings
	payloadDigest, envelopeDigest contracts.Digest
}

func (a *ParsedArtifact) Bindings() Bindings {
	if a == nil {
		return Bindings{}
	}
	return a.bindings
}
func (a *ParsedArtifact) Payload() []byte {
	if a == nil {
		return nil
	}
	return bytes.Clone(a.payload)
}
func (a *ParsedArtifact) Envelope() []byte {
	if a == nil {
		return nil
	}
	return bytes.Clone(a.envelope)
}
func (a *ParsedArtifact) PayloadDigest() contracts.Digest {
	if a == nil {
		return contracts.Digest{}
	}
	return a.payloadDigest
}
func (a *ParsedArtifact) EnvelopeDigest() contracts.Digest {
	if a == nil {
		return contracts.Digest{}
	}
	return a.envelopeDigest
}
func (a *ParsedArtifact) Transcript() ([]byte, error) {
	if a == nil {
		return nil, ErrInvalidFormat
	}
	var domain string
	switch a.bindings.Kind {
	case TimeAttestation:
		domain = "TALENRO-NODE-TIME-ATTESTATION-V1"
	case Desired:
		domain = "TALENRO-NODE-DESIRED-STATE-V1"
	case Metadata:
		domain = "TALENRO-NODE-STATE-METADATA-V1"
	case RootRotation:
		domain = "TALENRO-NODE-STATE-ROOT-ROTATION-V1"
	default:
		return nil, ErrInvalidFormat
	}
	message := make([]byte, len(domain)+1+len(a.payload))
	copy(message, domain)
	copy(message[len(domain)+1:], a.payload)
	return message, nil
}

// Clear erases reachable owned bytes and fixed fields only; Go runtime copies
// and detached accessor copies are outside this object's ownership.
func (a *ParsedArtifact) Clear() {
	if a == nil {
		return
	}
	clear(a.payload)
	clear(a.envelope)
	clear(a.slots)
	a.embeddedRoot.Clear()
	a.root.Clear()
	a.metadata.Clear()
	a.signatures.Clear()
	*a = ParsedArtifact{}
}
