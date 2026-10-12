package nodeissuer

import (
	"bytes"
	"crypto/sha256"
)

func checkIssuerRelation(in Input, s lowerSummaries, certs certificateSnapshots) error {
	if certs.ca == nil || certs.leaf == nil {
		return ErrInvalidMaterial
	}
	if !bytes.Equal(certs.ca.RawSubjectPublicKeyInfo, in.CA.DeclaredSPKI) {
		return ErrIssuerRelationMismatch
	}
	if !bytes.Equal(certs.leaf.RawIssuer, certs.ca.RawSubject) {
		return ErrIssuerRelationMismatch
	}
	id := sha256.Sum256(certs.ca.RawSubjectPublicKeyInfo)
	if len(certs.ca.SubjectKeyId) != 32 || len(certs.leaf.AuthorityKeyId) != 32 || !bytes.Equal(certs.ca.SubjectKeyId, id[:]) || !bytes.Equal(certs.leaf.AuthorityKeyId, id[:]) || !bytes.Equal(certs.ca.SubjectKeyId, s.ca.SubjectKeyIdentifier[:]) || !bytes.Equal(certs.leaf.AuthorityKeyId, s.leaf.AuthorityKeyIdentifier[:]) {
		return ErrIssuerRelationMismatch
	}
	return nil
}
