package stateverify

import (
	"context"
	"crypto/ed25519"
	"talenro.local/platform/internal/nodecontrol/stateformat"
)

func canceled(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if ctx.Err() != nil {
		return ErrCanceled
	}
	return nil
}

// Materials are borrowed only from the sole format parser's validated output.
// Unknown shares are structurally valid but do not count; every known share is
// checked, even after quorum. Never pool the current and new rotation groups.
func verifyThreshold(ctx context.Context, root stateformat.RootMaterial, shares []stateformat.SignatureShare, transcript []byte) error {
	if e := canceled(ctx); e != nil {
		return e
	}
	var count uint64
	for _, s := range shares {
		for _, k := range root.Keys {
			if s.KeyID != k.KeyID {
				continue
			}
			if e := canceled(ctx); e != nil {
				return e
			}
			valid := ed25519.Verify(k.PublicKey[:], transcript, s.Signature[:])
			if e := canceled(ctx); e != nil {
				return e
			}
			if !valid {
				return ErrSignatureMismatch
			}
			count++
			break
		}
	}
	if e := canceled(ctx); e != nil {
		return e
	}
	if root.Threshold == 0 || count < root.Threshold {
		return ErrSignatureMismatch
	}
	return nil
}
