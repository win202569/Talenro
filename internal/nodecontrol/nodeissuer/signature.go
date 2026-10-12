package nodeissuer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
)

func checkLeafSignatureProfile(cert *x509.Certificate, raw signedLeaf) error {
	if cert == nil || !bytes.Equal(raw.tbs.full, cert.RawTBSCertificate) || !bytes.Equal(raw.issuer.full, cert.RawIssuer) || !bytes.Equal(raw.subject.full, cert.RawSubject) || !bytes.Equal(raw.spki.full, cert.RawSubjectPublicKeyInfo) || !bytes.Equal(raw.innerAlgorithm.full, raw.outerAlgorithm.full) {
		return ErrInvalidMaterial
	}
	seq, err := singleElement(raw.outerAlgorithm.full, 0x30)
	if err != nil {
		return err
	}
	fields := seq.content
	oid, err := takeElement(&fields, 6)
	if err != nil || !minimalOID(oid.content) {
		return ErrInvalidMaterial
	}
	if len(fields) != 0 {
		param, tail, err := readElement(fields)
		if err != nil || len(tail) != 0 || param.tag == 5 && len(param.content) != 0 {
			return ErrInvalidMaterial
		}
	}
	if !bytes.Equal(seq.full, []byte{0x30, 0x0a, 6, 8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 4, 3, 2}) {
		return ErrUnsupportedSignatureProfile
	}
	if cert.SignatureAlgorithm != x509.ECDSAWithSHA256 {
		return ErrInvalidMaterial
	}
	if len(raw.signature.content) < 2 || raw.signature.content[0] != 0 || !bytes.Equal(raw.signature.content[1:], cert.Signature) {
		return ErrInvalidMaterial
	}
	return checkSignatureDER(raw.signature.content[1:])
}
func signatureIntegerInRange(b []byte) bool {
	if b[0]&128 != 0 {
		return false
	}
	if b[0] == 0 {
		b = b[1:]
	}
	if len(b) == 0 || len(b) > 32 {
		return false
	}
	var fixed [32]byte
	defer clear(fixed[:])
	copy(fixed[32-len(b):], b)
	n := [32]byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xbc, 0xe6, 0xfa, 0xad, 0xa7, 0x17, 0x9e, 0x84, 0xf3, 0xb9, 0xca, 0xc2, 0xfc, 0x63, 0x25, 0x51}
	return bytes.Compare(fixed[:], n[:]) < 0
}
func checkSignatureDER(sig []byte) error {
	seq, err := singleElement(sig, 0x30)
	if err != nil {
		return err
	}
	fields := seq.content
	r, err := takeElement(&fields, 2)
	if err != nil {
		return err
	}
	s, err := takeElement(&fields, 2)
	if err != nil || len(fields) != 0 || !minimalInteger(r.content) || !minimalInteger(s.content) {
		return ErrInvalidMaterial
	}
	// Representation is completely consumed before policy/range rejection.
	if len(sig) < 8 || len(sig) > 72 || len(r.content) > 33 || len(s.content) > 33 || !signatureIntegerInRange(r.content) || !signatureIntegerInRange(s.content) {
		return ErrUnsupportedSignatureProfile
	}
	return nil
}
func verifyLeafSignature(ca *x509.Certificate, raw signedLeaf) error {
	if ca == nil {
		return ErrInvalidMaterial
	}
	key, ok := ca.PublicKey.(*ecdsa.PublicKey)
	if !ok || key == nil || key.Curve != elliptic.P256() || key.X == nil || key.Y == nil || !key.Curve.IsOnCurve(key.X, key.Y) || len(raw.signature.content) < 2 {
		return ErrInvalidMaterial
	}
	digest := sha256.Sum256(raw.tbs.full)
	if !ecdsa.VerifyASN1(key, digest[:], raw.signature.content[1:]) {
		return ErrInvalidSignature
	}
	return nil
}
