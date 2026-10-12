package nodeleafkeys

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
)

func checkDeclaredIssuerSPKI(der []byte) error {
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return ErrInvalidInput
	}
	switch k := key.(type) {
	case *rsa.PublicKey:
		// Parser/canonical representation only, not CA strength or trust policy.
		if k == nil {
			return ErrInvalidInput
		}
	case *ecdsa.PublicKey:
		if k == nil || (k.Curve != elliptic.P256() && k.Curve != elliptic.P384() && k.Curve != elliptic.P521()) || k.X == nil || k.Y == nil || (k.X.Sign() == 0 && k.Y.Sign() == 0) || !k.Curve.IsOnCurve(k.X, k.Y) {
			return ErrInvalidInput
		}
	case ed25519.PublicKey:
		// Encoding shape only; no private-key, point/subgroup or CA certification.
		if len(k) != ed25519.PublicKeySize {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}
	canonical, err := x509.MarshalPKIXPublicKey(key)
	defer clear(canonical)
	if err != nil || !bytes.Equal(canonical, der) {
		return ErrInvalidInput
	}
	return nil
}
