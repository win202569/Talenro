package nodeleafkeys

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
)

func checkLowerSummary(s nodeleaf.Summary) error {
	want := [6]nodeleaf.Obligation{
		nodeleaf.LeafSignatureAndIssuerTrust,
		nodeleaf.KeyIdentifierDerivationAndIssuerBinding,
		nodeleaf.TrustedCurrentTime,
		nodeleaf.IdentityAndAuthorityProvenance,
		nodeleaf.PrivateKeyPossession,
		nodeleaf.MainStateAndRuntimeBinding,
	}
	if s.Proof != nodeleaf.NodeLeafMaterialShapeValid || s.Pending != want {
		return ErrInvalidMaterial
	}
	return nil
}
func checkLeafSnapshot(der, expectedSPKI []byte, lower nodeleaf.Summary) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate(der)
	if err != nil || !bytes.Equal(cert.Raw, der) || !bytes.Equal(cert.RawSubjectPublicKeyInfo, expectedSPKI) || contracts.Digest(sha256.Sum256(der)) != lower.LeafDERDigest || contracts.Digest(sha256.Sum256(cert.RawSubjectPublicKeyInfo)) != lower.SPKIDERDigest {
		return nil, ErrInvalidMaterial
	}
	return cert, nil
}
func keyIdentifierForSPKI(der []byte) KeyIdentifier { return KeyIdentifier(sha256.Sum256(der)) }
func checkKeyIdentifiers(cert *x509.Certificate, declaredIssuerSPKI []byte) (KeyIdentifier, KeyIdentifier, contracts.Digest, error) {
	if cert == nil {
		return KeyIdentifier{}, KeyIdentifier{}, contracts.Digest{}, ErrInvalidMaterial
	}
	if len(cert.SubjectKeyId) != 32 {
		return KeyIdentifier{}, KeyIdentifier{}, contracts.Digest{}, ErrUnsupportedKeyIdentifierProfile
	}
	if len(cert.AuthorityKeyId) != 32 {
		return KeyIdentifier{}, KeyIdentifier{}, contracts.Digest{}, ErrUnsupportedKeyIdentifierProfile
	}
	ski := keyIdentifierForSPKI(cert.RawSubjectPublicKeyInfo)
	aki := keyIdentifierForSPKI(declaredIssuerSPKI)
	if !bytes.Equal(cert.SubjectKeyId, ski[:]) {
		return KeyIdentifier{}, KeyIdentifier{}, contracts.Digest{}, ErrBindingMismatch
	}
	if !bytes.Equal(cert.AuthorityKeyId, aki[:]) {
		return KeyIdentifier{}, KeyIdentifier{}, contracts.Digest{}, ErrBindingMismatch
	}
	return ski, aki, contracts.Digest(aki), nil
}
