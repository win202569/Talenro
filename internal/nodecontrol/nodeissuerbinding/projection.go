package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
)

func bindProjection(ctx context.Context, in Input, evidence nodeleafbinding.Summary) (Projection, error) {
	var scratch referenceScratch
	issuer, serial, e := deriveReferences(ctx, evidence, &scratch)
	if e != nil {
		return Projection{}, e
	}
	identity := in.Time.Policy.Binding.Main.Proposed.Identity
	if identity.IssuerID != issuer {
		return Projection{}, stageError(ctx, ErrBindingMismatch)
	}
	if identity.Serial != serial {
		return Projection{}, stageError(ctx, ErrBindingMismatch)
	}
	if e = stageError(ctx, nil); e != nil {
		return Projection{}, e
	}
	return Projection{Leaf: evidence.Proposed, IssuerID: issuer, Serial: serial}, nil
}
