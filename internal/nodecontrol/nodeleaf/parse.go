package nodeleaf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// Parse checks partial material shape and exact caller expectations, never trust.
// Input mutation and reads concurrent with Clear are not supported.
func Parse(ctx context.Context, in Input) (*Material, error) {
	if err := preflight(ctx, in); err != nil {
		return nil, err
	}
	o := ownInput(in)
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	err := checkExpectedSPKI(o.expectedSPKI)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(o.der)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil || !bytes.Equal(cert.Raw, o.der) {
		return nil, ErrInvalidMaterial
	}
	raw, err := parseRawLeaf(o.der)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	uri, err := checkExtensions(raw.extensions, cert)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	serial, length, err := checkLeafFields(cert, raw)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(uri, []byte(expectedURI(in))) {
		return nil, ErrBindingMismatch
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !bytes.Equal(cert.RawSubjectPublicKeyInfo, o.expectedSPKI) {
		return nil, ErrBindingMismatch
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o.summary = Summary{
		Proof:   NodeLeafMaterialShapeValid,
		Pending: [6]Obligation{LeafSignatureAndIssuerTrust, KeyIdentifierDerivationAndIssuerBinding, TrustedCurrentTime, IdentityAndAuthorityProvenance, PrivateKeyPossession, MainStateAndRuntimeBinding},
		NodeID:  in.NodeID, TrustDomain: in.TrustDomain,
		LeafDERDigest: contracts.Digest(sha256.Sum256(o.der)),
		SPKIDERDigest: contracts.Digest(sha256.Sum256(cert.RawSubjectPublicKeyInfo)),
		Serial:        serial, SerialLength: length, NotBefore: cert.NotBefore.UTC(), NotAfter: cert.NotAfter.UTC(),
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o.valid = true
	transferred = true
	return &Material{owned: o}, nil
}
