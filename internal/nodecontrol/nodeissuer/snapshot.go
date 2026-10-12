package nodeissuer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
)

type certificateSnapshots struct{ ca, leaf *x509.Certificate }

func serialSnapshotMatches(cert *x509.Certificate, serial [20]byte, length uint8) bool {
	if cert == nil || cert.SerialNumber == nil || cert.SerialNumber.Sign() <= 0 {
		return false
	}
	b := cert.SerialNumber.Bytes()
	defer clear(b)
	if len(b) > 20 || len(b) != int(length) {
		return false
	}
	var fixed [20]byte
	defer clear(fixed[:])
	copy(fixed[:], b)
	return fixed == serial
}
func parseSnapshots(ctx context.Context, in Input, s lowerSummaries) (certificateSnapshots, error) {
	var out certificateSnapshots
	if err := stageError(ctx, nil); err != nil {
		return out, err
	}
	ca, err := x509.ParseCertificate(in.CA.DER)
	if err != nil {
		err = ErrInvalidMaterial
	}
	if err = stageError(ctx, err); err != nil {
		return out, err
	}
	path := -1
	if s.ca.HasPathLen {
		path = int(s.ca.PathLen)
	}
	if ca == nil || !bytes.Equal(ca.Raw, in.CA.DER) || sha256.Sum256(ca.Raw) != s.ca.CADERDigest || sha256.Sum256(ca.RawSubjectPublicKeyInfo) != s.ca.SPKIDERDigest || sha256.Sum256(ca.RawSubject) != s.ca.SubjectDERDigest || sha256.Sum256(ca.RawIssuer) != s.ca.IssuerDERDigest || !bytes.Equal(ca.SubjectKeyId, s.ca.SubjectKeyIdentifier[:]) || !bytes.Equal(ca.AuthorityKeyId, s.ca.AuthorityKeyIdentifier[:]) || !serialSnapshotMatches(ca, s.ca.Serial, s.ca.SerialLength) || !ca.NotBefore.UTC().Equal(s.ca.NotBefore) || !ca.NotAfter.UTC().Equal(s.ca.NotAfter) || ca.MaxPathLen != path || ca.MaxPathLenZero != (s.ca.HasPathLen && s.ca.PathLen == 0) || !s.ca.HasPathLen && s.ca.PathLen != 0 || (ca.KeyUsage&x509.KeyUsageCRLSign != 0) != s.ca.CRLSignAllowed {
		return out, stageError(ctx, ErrInvalidMaterial)
	}
	if err = stageError(ctx, nil); err != nil {
		return out, err
	}
	leaf, err := x509.ParseCertificate(in.Leaf.DER)
	if err != nil {
		err = ErrInvalidMaterial
	}
	if err = stageError(ctx, err); err != nil {
		return out, err
	}
	if leaf == nil || !bytes.Equal(leaf.Raw, in.Leaf.DER) || sha256.Sum256(leaf.Raw) != s.leaf.Leaf.LeafDERDigest || sha256.Sum256(leaf.RawSubjectPublicKeyInfo) != s.leaf.Leaf.SPKIDERDigest || sha256.Sum256(in.CA.DeclaredSPKI) != s.leaf.DeclaredIssuerSPKIDERDigest || !bytes.Equal(leaf.SubjectKeyId, s.leaf.SubjectKeyIdentifier[:]) || !bytes.Equal(leaf.AuthorityKeyId, s.leaf.AuthorityKeyIdentifier[:]) || !serialSnapshotMatches(leaf, s.leaf.Leaf.Serial, s.leaf.Leaf.SerialLength) || !leaf.NotBefore.UTC().Equal(s.leaf.Leaf.NotBefore) || !leaf.NotAfter.UTC().Equal(s.leaf.Leaf.NotAfter) || in.Leaf.NodeID != s.leaf.Leaf.NodeID || in.Leaf.TrustDomain != s.leaf.Leaf.TrustDomain {
		return out, stageError(ctx, ErrInvalidMaterial)
	}
	if err = stageError(ctx, nil); err != nil {
		return out, err
	}
	return certificateSnapshots{ca: ca, leaf: leaf}, nil
}
