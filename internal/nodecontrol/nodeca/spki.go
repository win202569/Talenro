package nodeca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
)

func p256Key(key any) bool {
	k, ok := key.(*ecdsa.PublicKey)
	return ok && k != nil && k.Curve == elliptic.P256() && k.X != nil && k.Y != nil && (k.X.Sign() != 0 || k.Y.Sign() != 0) && k.Curve.IsOnCurve(k.X, k.Y)
}
func canonicalSPKI(der []byte, bad error) (*ecdsa.PublicKey, error) {
	// Struct decoding may ignore unmatched trailing SEQUENCE elements. Consume
	// the known SPKI grammar before classifying a decoded key as unsupported.
	seq, e := singleElement(der, 0x30)
	if e != nil {
		return nil, bad
	}
	body := seq.content
	algorithm, e := takeElement(&body, 0x30)
	if e != nil || checkAlgorithm(algorithm) != nil {
		return nil, bad
	}
	bits, e := takeElement(&body, 3)
	if e != nil || len(body) != 0 || !bitStringShape(bits.content) {
		return nil, bad
	}
	algorithmBody := algorithm.content
	id, e := takeElement(&algorithmBody, 6)
	if e != nil {
		return nil, bad
	}
	isDSA := bytes.Equal(id.full, []byte{6, 7, 0x2a, 0x86, 0x48, 0xce, 0x38, 4, 1})
	if isDSA && !completeDSASPKI(algorithmBody, bits.content) {
		return nil, bad
	}
	key, e := x509.ParsePKIXPublicKey(der)
	if e != nil {
		return nil, bad
	}
	// The pinned parser accepts DSA but its marshaler does not. Consume that
	// known encoding explicitly, then retain the valid-unsupported result.
	if isDSA {
		return nil, ErrUnsupportedCAProfile
	}
	encoded, e := x509.MarshalPKIXPublicKey(key)
	defer clear(encoded)
	if e != nil || !bytes.Equal(encoded, der) {
		return nil, bad
	}
	// Re-encoding before this decision also detects discarded RSA payload
	// fields, nonzero unused bits and other known-key representation changes.
	if !p256Key(key) {
		return nil, ErrUnsupportedCAProfile
	}
	// Exact re-encoding freezes named-curve OIDs, uncompressed point and unused bits.
	if len(der) != 91 {
		return nil, bad
	}
	return key.(*ecdsa.PublicKey), nil
}

func completeDSASPKI(parameters, bits []byte) bool {
	if len(bits) == 0 || bits[0] != 0 {
		return false
	}
	y, e := singleElement(bits[1:], 2)
	if e != nil || !minimalInteger(y.content) {
		return false
	}
	params, e := singleElement(parameters, 0x30)
	if e != nil {
		return false
	}
	body := params.content
	for i := 0; i < 3; i++ {
		n, e := takeElement(&body, 2)
		if e != nil || !minimalInteger(n.content) {
			return false
		}
	}
	return len(body) == 0
}
func checkDeclaredSPKI(der []byte) error { _, e := canonicalSPKI(der, ErrInvalidInput); return e }
func checkCASPKI(cert *x509.Certificate, raw rawCA) error {
	if cert == nil || !bytes.Equal(cert.RawSubjectPublicKeyInfo, raw.spki.full) {
		return ErrInvalidMaterial
	}
	key, e := canonicalSPKI(raw.spki.full, ErrInvalidMaterial)
	if e != nil {
		return e
	}
	if !p256Key(cert.PublicKey) {
		return ErrInvalidMaterial
	}
	parsed := cert.PublicKey.(*ecdsa.PublicKey)
	if key.X.Cmp(parsed.X) != 0 || key.Y.Cmp(parsed.Y) != 0 {
		return ErrInvalidMaterial
	}
	return nil
}
