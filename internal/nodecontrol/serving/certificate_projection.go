package serving

import (
	"crypto/sha256"

	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type certificateQueryParams struct {
	Identity                          CertificateLookupInput
	AuthorityEpoch                    int64
	NodeScopeDigest, TrustScopeDigest contracts.Digest
	RequireAgentTrust                 bool
	TrustDomain                       string
}

// expectedEpoch must eventually come from the same Coordinator-bound ready
// snapshot. Validating its range alone is NOT proof of that provenance.
func certificateQueryArguments(lookup CertificateLookup, expectedEpoch uint64) (certificateQueryParams, error) {
	if lookup.Validate() != nil || !factsCoordinate(expectedEpoch) {
		return certificateQueryParams{}, authority.ErrInvalidArgument
	}
	i := lookup.Input()
	nodeTranscript := append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), i.NodeID[:]...)
	p := certificateQueryParams{Identity: i, AuthorityEpoch: int64(expectedEpoch),
		NodeScopeDigest:  sha256.Sum256(nodeTranscript),
		TrustScopeDigest: sha256.Sum256([]byte("TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust"))}
	if selector, ok := lookup.AgentTrustSelector(); ok {
		p.RequireAgentTrust = true
		p.TrustDomain = selector.TrustDomain
	}
	return p, nil
}

// selectCertificateProjection never publishes the first row of an ambiguous
// result. SQL limits candidates to two, which is enough to detect ambiguity.
// Its result remains private until the bound source's post-check succeeds.
func selectCertificateProjection(lookup CertificateLookup, rows []CertificateSnapshot) (CertificateFacts, error) {
	if lookup.Validate() != nil {
		return CertificateFacts{}, authority.ErrInvalidArgument
	}
	switch len(rows) {
	case 0:
		return CertificateFacts{}, authority.ErrNotFound
	case 1:
		return newCertificateFacts(lookup, rows[0])
	default:
		return CertificateFacts{}, authority.ErrConflict
	}
}
