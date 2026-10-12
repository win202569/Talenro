package nodeca

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
)

func stageError(ctx context.Context, err error) error {
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	return err
}

// Parse checks partial material facts only. All post-preflight reads are owned.
func Parse(ctx context.Context, in Input) (*Material, error) {
	if err := preflight(ctx, in); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
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

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := stageError(ctx, checkDeclaredSPKI(o.declaredSPKI)); err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(o.der)
	if err != nil || cert == nil || !bytes.Equal(cert.Raw, o.der) {
		err = ErrInvalidMaterial
	}
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	raw, err := parseRawCA(o.der)
	if err == nil && (!bytes.Equal(raw.tbs.full, cert.RawTBSCertificate) || !bytes.Equal(raw.signature.content[1:], cert.Signature)) {
		err = ErrInvalidMaterial
	}
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	fields, err := checkCAFields(cert, raw)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	extensions, err := checkCAExtensions(cert, raw.extensions)
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	spkiDigest := sha256.Sum256(raw.spki.full)
	if !bytes.Equal(raw.spki.full, o.declaredSPKI) {
		err = ErrBindingMismatch
	} else if !bytes.Equal(extensions.SubjectKeyIdentifier[:], spkiDigest[:]) {
		err = ErrBindingMismatch
	}
	if err = stageError(ctx, err); err != nil {
		return nil, err
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	summary := Summary{
		Proof:       NodeCAMaterialShapeAndDeclaredKeyBound,
		Pending:     [6]Obligation{CertificateSignatureAndIssuerChain, AuthorityIdentifierAndParentBinding, TrustedCurrentTime, DeploymentAuthorizationAndPurpose, BundleStatusAndRollback, LeafIssuerRelationAndRuntimeUse},
		CADERDigest: sha256.Sum256(o.der), SPKIDERDigest: spkiDigest, SubjectDERDigest: sha256.Sum256(raw.subject.full), IssuerDERDigest: sha256.Sum256(raw.issuer.full),
		SubjectKeyIdentifier: extensions.SubjectKeyIdentifier, AuthorityKeyIdentifier: extensions.AuthorityKeyIdentifier,
		Serial: fields.Serial, SerialLength: fields.SerialLength, NotBefore: fields.NotBefore, NotAfter: fields.NotAfter,
		HasPathLen: extensions.HasPathLen, PathLen: extensions.PathLen, CRLSignAllowed: extensions.CRLSignAllowed,
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	o.summary = summary
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	o.valid = true
	transferred = true
	return &Material{owned: o}, nil
}
