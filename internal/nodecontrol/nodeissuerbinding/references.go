package nodeissuerbinding

import (
	"context"
	"crypto/sha256"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
)

const issuerReferenceDomain = "talenro.main-selected-issuer-ref.v1\x00"
const serialReferenceDomain = "talenro.main-selected-serial-ref.v1\x00"

type referenceScratch struct {
	issuer [68]byte
	serial [89]byte
}

func (s *referenceScratch) clear() {
	if s == nil {
		return
	}
	clear(s.issuer[:])
	clear(s.serial[:])
}
func deriveReferences(ctx context.Context, e nodeleafbinding.Summary, s *referenceScratch) (contracts.Digest, contracts.Digest, error) {
	if s != nil {
		defer s.clear()
	}
	if err := stageError(ctx, nil); err != nil {
		return contracts.Digest{}, contracts.Digest{}, err
	}
	if s == nil {
		return contracts.Digest{}, contracts.Digest{}, ErrInvalidEvidence
	}
	c := e.Evidence.Certificate.CA.CADERDigest
	serial := e.Evidence.Certificate.Leaf.Leaf.Serial
	length := e.Evidence.Certificate.Leaf.Leaf.SerialLength
	if c == (contracts.Digest{}) || length < 1 || length > 20 || serial[0] == 0 {
		return contracts.Digest{}, contracts.Digest{}, stageError(ctx, ErrInvalidEvidence)
	}
	for _, b := range serial[length:] {
		if b != 0 {
			return contracts.Digest{}, contracts.Digest{}, stageError(ctx, ErrInvalidEvidence)
		}
	}
	copy(s.issuer[:36], issuerReferenceDomain)
	copy(s.issuer[36:], c[:])
	issuer := contracts.Digest(sha256.Sum256(s.issuer[:]))
	if err := stageError(ctx, nil); err != nil {
		return contracts.Digest{}, contracts.Digest{}, err
	}
	copy(s.serial[:36], serialReferenceDomain)
	copy(s.serial[36:68], issuer[:])
	s.serial[68] = length
	copy(s.serial[69:], serial[:length])
	reference := contracts.Digest(sha256.Sum256(s.serial[:69+int(length)]))
	if err := stageError(ctx, nil); err != nil {
		return contracts.Digest{}, contracts.Digest{}, err
	}
	return issuer, reference, nil
}
