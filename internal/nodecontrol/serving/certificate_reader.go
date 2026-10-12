package serving

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// readCertificateFacts is private staging composition, not a production reader.
// No caller-supplied epoch, exported constructor, or v7 authorization is added.
func readCertificateFacts(ctx context.Context, source BoundAuthorityReadSource, lookup CertificateLookup) (CertificateFacts, error) {
	return readCertificateFactsWithEpoch(ctx, source, lookup, authority.BoundReadEpoch)
}

// The private resolver seam permits isolated composition tests without exposing
// a factory for authority's private source. The real entry always fixes it to
// BoundReadEpoch; it must never become configuration or a public capability.
func readCertificateFactsWithEpoch(ctx context.Context, source BoundAuthorityReadSource, lookup CertificateLookup, resolve func(store.DBTX) (uint64, error)) (CertificateFacts, error) {
	return guardedRead(ctx, source, func(boundCtx context.Context, db store.DBTX) (CertificateFacts, error) {
		if resolve == nil {
			return CertificateFacts{}, authority.ErrAuthorityUnavailable
		}
		epoch, err := resolve(db)
		if readCanceled(err) {
			return CertificateFacts{}, authority.ErrCanceled
		}
		if err != nil || !factsCoordinate(epoch) {
			return CertificateFacts{}, authority.ErrAuthorityUnavailable
		}
		return queryCertificateFacts(boundCtx, db, lookup, epoch)
	}, discardCertificateFacts)
}
