package nodeleafkeys

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
)

// Evaluate checks declared key-identifier relations, never issuer authentication.
// Synchronous parsing cannot be interrupted internally; cancellation is checked
// at stage boundaries and takes precedence over each completed stage's error.
func Evaluate(ctx context.Context, in Input) (*Material, error) {
	err := preflight(ctx, in)
	if ctx != nil {
		if canceled := ctx.Err(); canceled != nil {
			return nil, canceled
		}
	}
	if err != nil {
		return nil, err
	}
	o := ownInput(in)
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lower, err := nodeleaf.Parse(ctx, nodeleaf.Input{DER: o.der, ExpectedSPKI: o.expectedSPKI, NodeID: in.Leaf.NodeID, TrustDomain: in.Leaf.TrustDomain})
	defer lower.Clear()
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	leafSummary := lower.Summary()
	err = checkLowerSummary(leafSummary)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	err = checkDeclaredIssuerSPKI(o.declaredIssuerSPKI)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	cert, err := checkLeafSnapshot(o.der, o.expectedSPKI, leafSummary)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	ski, aki, issuerDigest, err := checkKeyIdentifiers(cert, o.declaredIssuerSPKI)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o.summary = Summary{
		Proof: NodeLeafDeclaredKeyIdentifiersBound, Leaf: leafSummary,
		SubjectKeyIdentifier: ski, AuthorityKeyIdentifier: aki,
		DeclaredIssuerSPKIDERDigest: issuerDigest,
	}
	o.valid = true
	transferred = true
	return &Material{owned: o}, nil
}
