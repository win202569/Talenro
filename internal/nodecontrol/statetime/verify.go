package statetime

import (
	"context"
	"crypto/ed25519"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
)

type verifiedFacts struct {
	att, meta  *stateformat.ParsedArtifact
	material   stateformat.MetadataMaterial
	key        stateformat.MetadataKeyMaterial
	transcript []byte
}

func (f *verifiedFacts) Clear() {
	if f == nil {
		return
	}
	f.att.Clear()
	f.meta.Clear()
	f.material.Clear()
	clear(f.transcript)
	*f = verifiedFacts{}
}
func canceled(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	return ctx.Err()
}
func validU(n uint64) bool { return n > 0 && n <= 9223372036854775807 }
func verifyModel(ctx context.Context, in ModelInput) (result *verifiedFacts, err error) {
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	for _, raw := range [][]byte{in.AttestationEnvelope, in.MetadataEnvelope} {
		if len(raw) == 0 || len(raw) > 262144 {
			return nil, ErrInvalidInput
		}
	}
	if in.Metadata.Status != Available || in.Time.Status != Available || in.Request.Status != Available || in.Checkpoint.Status != Available {
		return nil, ErrAssuranceMissing
	}
	for _, d := range []contracts.Digest{in.Metadata.SourceID, in.Metadata.ObservationID, in.Metadata.StateID, in.Metadata.Payload, in.Metadata.Envelope, in.Time.ProviderID, in.Time.ObservationID, in.Request.RequestID, in.Request.ObservationID, in.Checkpoint.SourceID, in.Checkpoint.ObservationID, in.Checkpoint.StateID} {
		if d == (contracts.Digest{}) {
			return nil, ErrAssuranceMissing
		}
	}
	if !validU(in.Metadata.Epoch) || !validU(in.Metadata.Version) || !validU(in.Request.Epoch) || !validU(in.Checkpoint.Epoch) || !validU(in.Checkpoint.Sequence) {
		return nil, ErrInvalidInput
	}
	for _, n := range []uuid.UUID{in.Metadata.NodeID, in.Request.NodeID, in.Checkpoint.NodeID} {
		if n == uuid.Nil || n.Variant() != uuid.RFC4122 {
			return nil, ErrInvalidInput
		}
	}
	f := &verifiedFacts{}
	defer func() {
		if result == nil {
			f.Clear()
		}
	}()
	f.att, err = stateformat.ParseEnvelope(stateformat.TimeAttestation, in.AttestationEnvelope)
	if err != nil {
		return nil, ErrInvalidInput
	}
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	f.meta, err = stateformat.ParseEnvelope(stateformat.Metadata, in.MetadataEnvelope)
	if err != nil {
		return nil, ErrInvalidInput
	}
	a, m := f.att.Bindings(), f.meta.Bindings()
	if in.Metadata.Payload != f.meta.PayloadDigest() || in.Metadata.Envelope != f.meta.EnvelopeDigest() || in.Metadata.Version != m.MetadataVersion || in.Metadata.Epoch != m.Epoch || a.Epoch != m.Epoch || in.Request.Epoch != a.Epoch || in.Checkpoint.Epoch != a.Epoch || in.Metadata.NodeID != a.NodeID || in.Request.NodeID != a.NodeID || in.Checkpoint.NodeID != a.NodeID || in.Request.ExpectedNonce != a.RequestNonce || in.Metadata.StateID != in.Checkpoint.StateID {
		return nil, ErrBindingMismatch
	}
	f.material, err = f.meta.MetadataMaterial()
	if err != nil {
		return nil, ErrInvalidInput
	}
	for _, id := range f.material.RevokedIDs {
		if id == a.SigningKeyID {
			return nil, ErrUnauthorizedKey
		}
	}
	found := false
	for _, k := range f.material.Keys {
		if k.KeyID == a.SigningKeyID {
			if k.Status != "active" && k.Status != "retiring" {
				return nil, ErrUnauthorizedKey
			}
			f.key = k
			found = true
			break
		}
	}
	if !found {
		return nil, ErrUnauthorizedKey
	}
	f.transcript, err = f.att.Transcript()
	if err != nil {
		return nil, ErrInvalidInput
	}
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	ok := ed25519.Verify(f.key.PublicKey[:], f.transcript, a.Signature[:])
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	if !ok {
		return nil, ErrSignatureMismatch
	}
	return f, nil
}
