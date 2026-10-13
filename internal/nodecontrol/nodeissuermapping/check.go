package nodeissuermapping

import "context"

// CheckDeclared checks caller declarations against one real lower evaluation.
// It neither authenticates those declarations nor authorizes any runtime action.
func CheckDeclared(ctx context.Context, in Input) (result *Model, err error) {
	if e := preflight(ctx, in); e != nil {
		return nil, e
	}
	o := &owner{}
	defer func() {
		if result == nil {
			o.clear()
		}
	}()
	o.claims, err = ownClaims(ctx, in.Claims)
	if err != nil {
		return nil, err
	}
	if e := evaluateLower(ctx, o, in.Material); e != nil {
		return nil, e
	}
	proposed, e := bindProjection(ctx, o.claims, o.summary.Evidence)
	if e != nil {
		return nil, e
	}
	o.summary.Proof = DeclaredIssuerSerialCorrespondenceConditionalTimeBound
	o.summary.Profile = declaredMappingProfile
	o.summary.Proposed = proposed
	o.summary.ObservedAt = o.summary.Evidence.ObservedAt
	o.summary.ValidUntil = o.summary.Evidence.ValidUntil
	o.summary.Obligations = [5]Obligation{AuthenticatedRegistryPublicationAndRollbackProtection, AuthenticatedIssuanceLinkAndRowStatus, GuardedV7SnapshotAndLease, CrossRenewalDatabaseUniqueness, FullIdentityCompositionCommitAndRuntimeAuthorization}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	o.valid = true
	return &Model{owned: o}, nil
}
