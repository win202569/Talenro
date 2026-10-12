package nodeissuer

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
)

type lowerSummaries struct {
	ca   nodeca.Summary
	leaf nodeleafkeys.Summary
}

func stageError(ctx context.Context, err error) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	return err
}
func checkCAProof(s nodeca.Summary) error {
	if s.Proof != nodeca.NodeCAMaterialShapeAndDeclaredKeyBound || s.Pending != ([6]nodeca.Obligation{nodeca.CertificateSignatureAndIssuerChain, nodeca.AuthorityIdentifierAndParentBinding, nodeca.TrustedCurrentTime, nodeca.DeploymentAuthorizationAndPurpose, nodeca.BundleStatusAndRollback, nodeca.LeafIssuerRelationAndRuntimeUse}) {
		return ErrInvalidMaterial
	}
	return nil
}
func checkLowerProofs(s lowerSummaries) error {
	if err := checkCAProof(s.ca); err != nil {
		return err
	}
	if s.leaf.Proof != nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound || s.leaf.Leaf.Proof != nodeleaf.NodeLeafMaterialShapeValid || s.leaf.Leaf.Pending != ([6]nodeleaf.Obligation{nodeleaf.LeafSignatureAndIssuerTrust, nodeleaf.KeyIdentifierDerivationAndIssuerBinding, nodeleaf.TrustedCurrentTime, nodeleaf.IdentityAndAuthorityProvenance, nodeleaf.PrivateKeyPossession, nodeleaf.MainStateAndRuntimeBinding}) {
		return ErrInvalidMaterial
	}
	return nil
}
func evaluateLower(ctx context.Context, in Input) (lowerSummaries, error) {
	var out lowerSummaries
	if err := stageError(ctx, nil); err != nil {
		return out, err
	}
	ca, err := nodeca.Parse(ctx, in.CA)
	defer ca.Clear()
	if err = stageError(ctx, err); err != nil {
		return out, err
	}
	if ca == nil {
		return out, stageError(ctx, ErrInvalidMaterial)
	}
	caSummary := ca.Summary()
	if err = stageError(ctx, checkCAProof(caSummary)); err != nil {
		return out, err
	}
	if err = stageError(ctx, nil); err != nil {
		return out, err
	}
	leaf, err := nodeleafkeys.Evaluate(ctx, nodeleafkeys.Input{Leaf: in.Leaf, DeclaredIssuerSPKI: in.CA.DeclaredSPKI})
	defer leaf.Clear()
	if err = stageError(ctx, err); err != nil {
		return out, err
	}
	if leaf == nil {
		return out, stageError(ctx, ErrInvalidMaterial)
	}
	candidate := lowerSummaries{ca: caSummary, leaf: leaf.Summary()}
	if err = stageError(ctx, checkLowerProofs(candidate)); err != nil {
		return out, err
	}
	return candidate, nil
}
