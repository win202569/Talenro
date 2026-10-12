package nodeissuer

import (
	"context"
	"crypto/sha256"
)

// Evaluate proves only the direct relation of these exact caller-provided inputs.
// It does not discharge any of either lower layer's pending obligations.
func Evaluate(ctx context.Context, in Input) (*Material, error) {
	if err := preflight(ctx, in); err != nil {
		return nil, err
	}
	o := ownInput(in)
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	if err := stageError(ctx, nil); err != nil {
		return nil, err
	}
	owned := o.input(in.Leaf.NodeID, in.Leaf.TrustDomain)
	s, err := evaluateLower(ctx, owned)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	certs, err := parseSnapshots(ctx, owned, s)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	raw, err := parseLeafEnvelope(owned.Leaf.DER)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	err = checkLeafSignatureProfile(certs.leaf, raw)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	err = checkIssuerRelation(owned, s, certs)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	err = verifyLeafSignature(certs.ca, raw)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}
	o.summary = Summary{Proof: NodeLeafDirectIssuerSignatureBound, Leaf: s.leaf, CA: s.ca, LeafTBSDERDigest: sha256.Sum256(raw.tbs.full), LeafIssuerDERDigest: sha256.Sum256(raw.issuer.full)}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	o.valid = true
	transferred = true
	return &Material{owned: o}, nil
}
