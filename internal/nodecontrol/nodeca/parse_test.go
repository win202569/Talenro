package nodeca

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func expectedPending() [6]Obligation {
	return [6]Obligation{CertificateSignatureAndIssuerChain, AuthorityIdentifierAndParentBinding, TrustedCurrentTime, DeploymentAuthorizationAndPurpose, BundleStatusAndRollback, LeafIssuerRelationAndRuntimeUse}
}
func fixtureDigest(s string) contracts.Digest {
	var d contracts.Digest
	copy(d[:], fixtureHex(s))
	return d
}
func requireCA(t testing.TB, in Input) *Material {
	t.Helper()
	m, e := Parse(context.Background(), in)
	if e != nil || m == nil {
		t.Fatalf("Parse %v", e)
	}
	if m.Summary().Proof != NodeCAMaterialShapeAndDeclaredKeyBound || m.Summary().Pending != expectedPending() {
		t.Fatal("partial proof")
	}
	t.Cleanup(m.Clear)
	return m
}
func requireCAError(t testing.TB, ctx context.Context, in Input, want error) {
	t.Helper()
	m, e := Parse(ctx, in)
	if m != nil || !errors.Is(e, want) {
		if m != nil {
			m.Clear()
		}
		t.Fatalf("Parse got material=%v error=%v want=%v", m != nil, e, want)
	}
	allowed := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidMaterial, ErrUnsupportedCAProfile, ErrBindingMismatch, context.Canceled, context.DeadlineExceeded}
	known := false
	for _, v := range allowed {
		known = known || e == v
	}
	if !known {
		t.Fatal("dynamic error", e)
	}
}

func TestCAParse(t *testing.T) {
	for _, parent := range []bool{false, true} {
		in := fixtureInput(parent)
		m := requireCA(t, in)
		got := m.Summary()
		derHash, issuerHash, aki := selfDERHash, subjectNameHash, subjectSPKIHash
		if parent {
			derHash = parentDERHash
			issuerHash = parentNameHash
			aki = parentKeyHash
		}
		want := Summary{Proof: NodeCAMaterialShapeAndDeclaredKeyBound, Pending: expectedPending(), CADERDigest: fixtureDigest(derHash), SPKIDERDigest: fixtureDigest(subjectSPKIHash), SubjectDERDigest: fixtureDigest(subjectNameHash), IssuerDERDigest: fixtureDigest(issuerHash), Serial: [20]byte{128}, SerialLength: 1, NotBefore: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), NotAfter: time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC)}
		copy(want.SubjectKeyIdentifier[:], fixtureHex(subjectSPKIHash))
		copy(want.AuthorityKeyIdentifier[:], fixtureHex(aki))
		if got != want {
			t.Fatalf("all-field summary\ngot %+v\nwant %+v", got, want)
		}
		ownDER, ownSPKI := m.owned.der, m.owned.declaredSPKI
		clear(in.DER)
		clear(in.DeclaredSPKI)
		if m.Summary() != want || !bytes.Equal(ownDER, fixtureHex(func() string {
			if parent {
				return parentCAHex
			}
			return selfCAHex
		}())) {
			t.Fatal("input alias")
		}
		copied := *m
		m.Clear()
		if copied.Summary() != (Summary{}) || got != want || !bytes.Equal(ownDER, make([]byte, len(ownDER))) || !bytes.Equal(ownSPKI, make([]byte, len(ownSPKI))) {
			t.Fatal("owned clear")
		}
	}
	for _, path := range []byte{0, 1, 255} {
		for _, crl := range []bool{false, true} {
			in := fixtureExt(t, fixtureInput(false), 19, func(e *pkix.Extension) {
				b := []byte{path}
				if path >= 128 {
					b = append([]byte{0}, b...)
				}
				e.Value = fixtureJoin([][]byte{{1, 1, 255}, fixtureTLV(2, b)})
			})
			if crl {
				in = fixtureExt(t, in, 15, func(e *pkix.Extension) { e.Value = []byte{3, 2, 1, 6} })
			}
			exts := fixtureExts(t, in)
			for i, j := 0, len(exts)-1; i < j; i, j = i+1, j-1 {
				exts[i], exts[j] = exts[j], exts[i]
			}
			in = fixtureSetExts(t, in, exts)
			got := requireCA(t, in).Summary()
			if !got.HasPathLen || got.PathLen != path || got.CRLSignAllowed != crl {
				t.Fatal("path/KU")
			}
		}
	}
}
func TestCABindings(t *testing.T) {
	in := fixtureInput(false)
	in.DeclaredSPKI = fixtureHex(rfcSPKIHex)
	requireCAError(t, context.Background(), in, ErrBindingMismatch)
	wrongSKI := fixtureExt(t, fixtureInput(false), 14, func(e *pkix.Extension) { e.Value = fixtureTLV(4, make([]byte, 32)) })
	requireCAError(t, context.Background(), wrongSKI, ErrBindingMismatch)
	wrongSKI.DeclaredSPKI = fixtureHex(rfcSPKIHex)
	requireCAError(t, context.Background(), wrongSKI, ErrBindingMismatch)
}
func TestCAPartialProof(t *testing.T) {
	in := fixtureInput(false)
	unknownAlg := fixtureHex("300506032a0304")
	cases := []Input{in, fixtureRewrite(t, in, "signature", []byte{3, 2, 0, 1}), fixtureRewrite(t, fixtureRewrite(t, in, "innerAlgorithm", unknownAlg), "outerAlgorithm", unknownAlg)}
	nested := []byte{5, 0}
	for i := 0; i < 40; i++ {
		nested = fixtureTLV(0x30, nested)
	}
	deepAlg := fixtureJoin([][]byte{fixtureHex("06032a0304"), nested})
	deep := fixtureRewrite(t, fixtureRewrite(t, in, "innerAlgorithm", deepAlg), "outerAlgorithm", deepAlg)
	fixtureCert(t, deep)
	cases = append(cases, deep)
	expired := fixtureRewrite(t, in, "notBefore", fixtureTLV(0x17, []byte("000101000000Z")))
	expired = fixtureRewrite(t, expired, "notAfter", fixtureTLV(0x17, []byte("010101000000Z")))
	cases = append(cases, expired)
	future := fixtureRewrite(t, in, "notBefore", fixtureTLV(0x18, []byte("90000101000000Z")))
	future = fixtureRewrite(t, future, "notAfter", fixtureTLV(0x18, []byte("90010101000000Z")))
	cases = append(cases, future)
	for i, v := range cases {
		got := requireCA(t, v).Summary()
		if got.Proof != 1 || got.Pending != expectedPending() {
			t.Fatalf("case %d upgraded", i)
		}
	}
	if fixtureCert(t, cases[2]).SignatureAlgorithm != x509.UnknownSignatureAlgorithm {
		t.Fatal("unknown signature oracle")
	}
	requireCAError(t, context.Background(), fixtureRewrite(t, in, "signature", []byte{3, 1, 0}), ErrUnsupportedCAProfile)
	requireCAError(t, context.Background(), fixtureRewrite(t, in, "signature", []byte{3, 2, 1, 2}), ErrUnsupportedCAProfile)
}
func TestCAPartialProofParentRoles(t *testing.T) {
	for _, parent := range []bool{false, true} {
		in := fixtureExt(t, fixtureInput(parent), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, fixtureTLV(0x80, bytes.Repeat([]byte{0x42}, 32))) })
		got := requireCA(t, in).Summary()
		if !bytes.Equal(got.AuthorityKeyIdentifier[:], bytes.Repeat([]byte{0x42}, 32)) || got.SubjectKeyIdentifier == got.AuthorityKeyIdentifier || got.Pending != expectedPending() {
			t.Fatal("AKI treated as subject binding")
		}
	}
	in := fixtureInput(true)
	parentSPKI, e := x509.MarshalPKIXPublicKey(&fixtureKey(2, elliptic.P256()).PublicKey)
	if e != nil {
		t.Fatal(e)
	}
	in.DeclaredSPKI = parentSPKI
	requireCAError(t, context.Background(), in, ErrBindingMismatch)
}

type countingContext struct {
	context.Context
	calls, cancelAt int
}

func (c *countingContext) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return context.Canceled
	}
	return nil
}
func TestCACancellation(t *testing.T) {
	baseline := &countingContext{Context: context.Background(), cancelAt: 1000}
	m, e := Parse(baseline, fixtureInput(false))
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	if baseline.calls < 14 {
		t.Fatalf("too few stage boundaries: %d", baseline.calls)
	}
	for n := 1; n <= baseline.calls; n++ {
		ctx := &countingContext{Context: context.Background(), cancelAt: n}
		requireCAError(t, ctx, fixtureInput(false), context.Canceled)
	}
	bad := fixtureInput(false)
	bad.DeclaredSPKI = []byte{1}
	measure := &countingContext{Context: context.Background(), cancelAt: 1000}
	requireCAError(t, measure, bad, ErrInvalidInput)
	override := &countingContext{Context: context.Background(), cancelAt: measure.calls}
	requireCAError(t, override, bad, context.Canceled)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	requireCAError(t, canceled, Input{DER: make([]byte, 65537)}, context.Canceled)
	expired, cancelDeadline := context.WithDeadline(context.Background(), epochTime())
	defer cancelDeadline()
	requireCAError(t, expired, fixtureInput(false), context.DeadlineExceeded)
}
func TestCAParserPriority(t *testing.T) {
	in := fixtureInput(false)
	exts := fixtureExts(t, in)
	exts = append(exts, exts[0])
	duplicate := fixtureSetExts(t, in, exts)
	if _, e := x509.ParseCertificate(duplicate.DER); e == nil {
		t.Fatal("fixed parser unexpectedly accepts duplicate")
	}
	requireCAError(t, context.Background(), duplicate, ErrInvalidMaterial)
	trailing := fixtureExt(t, in, 14, func(e *pkix.Extension) { e.Value = append(e.Value, 0) })
	requireCAError(t, context.Background(), trailing, ErrInvalidMaterial)
	shortID := fixtureExt(t, in, 14, func(e *pkix.Extension) { e.Value = fixtureTLV(4, make([]byte, 20)) })
	requireCAError(t, context.Background(), shortID, ErrUnsupportedCAProfile)
	missing := fixtureExts(t, in)[1:]
	requireCAError(t, context.Background(), fixtureSetExts(t, in, missing), ErrUnsupportedCAProfile)
}
func TestCACombinedFailurePriority(t *testing.T) {
	in := fixtureInput(false)
	badDER := Input{DER: []byte{0}, DeclaredSPKI: bytes.Clone(in.DeclaredSPKI)}
	badBoth := badDER
	badBoth.DeclaredSPKI = []byte{0}
	version := fixtureRewrite(t, in, "version", fixtureTLV(0xa0, []byte{2, 1, 1}))
	version = fixtureExt(t, version, 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, fixtureTLV(0x80, make([]byte, 20))) })
	profileBeforeBinding := fixtureExt(t, in, 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, fixtureTLV(0x80, make([]byte, 20))) })
	profileBeforeBinding.DeclaredSPKI = fixtureHex(rfcSPKIHex)
	cases := []struct {
		ctx  context.Context
		in   Input
		want error
	}{
		{nil, Input{DER: make([]byte, 65537)}, ErrInvalidInput}, {context.Background(), Input{DeclaredSPKI: make([]byte, 8193)}, ErrLimitExceeded},
		{context.Background(), Input{DER: make([]byte, 65537)}, ErrLimitExceeded}, {context.Background(), Input{}, ErrInvalidInput},
		{context.Background(), badBoth, ErrInvalidInput}, {context.Background(), badDER, ErrInvalidMaterial}, {context.Background(), version, ErrUnsupportedCAProfile},
		{context.Background(), profileBeforeBinding, ErrUnsupportedCAProfile},
	}
	for _, c := range cases {
		requireCAError(t, c.ctx, c.in, c.want)
	}
	malformedName := fixtureRewrite(t, in, "subject", []byte{0x30, 2, 0x31, 0})
	malformedName = fixtureRewrite(t, malformedName, "serial", []byte{2, 1, 0})
	requireCAError(t, context.Background(), malformedName, ErrInvalidMaterial)
}

func TestCAReviewRDNCanonicalOrder(t *testing.T) {
	sorted := fixtureHex("30163114300806035504030c01613008060355040b0c0162")
	reversed := fixtureHex("301631143008060355040b0c0162300806035504030c0161")
	for _, path := range []string{"issuer", "subject"} {
		good := fixtureRewrite(t, fixtureInput(false), path, sorted)
		fixtureCert(t, good)
		requireCA(t, good)
		bad := fixtureRewrite(t, fixtureInput(false), path, reversed)
		fixtureCert(t, bad)
		requireCAError(t, context.Background(), bad, ErrInvalidMaterial)
	}
}
func TestCAReviewUniversalTagForms(t *testing.T) {
	// The fixed front parser tolerates these unknown-algorithm parameters; the
	// material scanner must reject illegal universal tag forms, not signer policy.
	for _, param := range [][]byte{{0x21, 0}, {0x22, 0}, {0x23, 0}, {0x24, 0}, {0x25, 0}, {0x26, 0}, {0x2c, 0}, {0x10, 0}, {0x11, 0}, {0x0a, 2, 0, 1}} {
		alg := fixtureJoin([][]byte{fixtureHex("06032a0304"), param})
		in := fixtureRewrite(t, fixtureRewrite(t, fixtureInput(false), "innerAlgorithm", alg), "outerAlgorithm", alg)
		fixtureCert(t, in)
		requireCAError(t, context.Background(), in, ErrInvalidMaterial)
	}
}
func TestCAReviewBasicConstraintsConsumption(t *testing.T) {
	for _, v := range []struct {
		hex          string
		want         error
		frontAccepts bool
	}{
		{"3003020101", ErrUnsupportedCAProfile, true},
		{"30050201010500", ErrInvalidMaterial, true},
		{"30020500", ErrInvalidMaterial, true},
		{"30020201", ErrInvalidMaterial, false},
		{"300402020001", ErrInvalidMaterial, false},
	} {
		b := fixtureHex(v.hex)
		c, raw := extensionChange(t, 19, b)
		if _, e := checkCAExtensions(c, raw); !errors.Is(e, v.want) {
			t.Errorf("helper %s got %v want %v", v.hex, e, v.want)
		}
		in := fixtureExt(t, fixtureInput(false), 19, func(e *pkix.Extension) { e.Value = b })
		if v.frontAccepts {
			fixtureCert(t, in)
		}
		requireCAError(t, context.Background(), in, v.want)
	}
}
func TestCAReviewRFCOracle(t *testing.T) {
	// The signature remains the original golden's and is NOT authenticated after
	// replacing its subject key. This checks only the approved partial projection.
	in := fixtureRewrite(t, fixtureInput(false), "spki", fixtureHex(rfcSPKIHex))
	in = fixtureExt(t, in, 14, func(e *pkix.Extension) { e.Value = fixtureTLV(4, fixtureHex(rfcSPKIHash)) })
	in.DeclaredSPKI = fixtureHex(rfcSPKIHex)
	got := requireCA(t, in).Summary()
	want := fixtureDigest(rfcSPKIHash)
	if got.SPKIDERDigest != want || got.SubjectKeyIdentifier != KeyIdentifier(want) || got.Pending != expectedPending() {
		t.Fatal("RFC full-SPKI full32 oracle", got)
	}
}

func TestCAReviewOpaqueSetParameters(t *testing.T) {
	// Unknown signer parameters have no SET/SET OF schema in this material layer.
	// This SET has tag-ordered [0] EXPLICIT NULL, [1] IMPLICIT INTEGER.
	alg := fixtureHex("300e06032a03043107a0020500810101")
	in := fixtureRewrite(t, fixtureRewrite(t, fixtureInput(false), "innerAlgorithm", alg), "outerAlgorithm", alg)
	cert := fixtureCert(t, in)
	if cert.SignatureAlgorithm != x509.UnknownSignatureAlgorithm {
		t.Fatal("fixture must remain an unknown signer algorithm")
	}
	m, e := Parse(context.Background(), in)
	if e != nil || m == nil {
		t.Fatalf("valid opaque SET rejected: %v", e)
	}
	defer m.Clear()
	if m.Summary().Pending != expectedPending() {
		t.Fatal("opaque parameters must not upgrade trust")
	}
}
func TestCAReviewAKINestedConsumption(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name, field string
		want        error
	}{
		{"truncated issuer child", "a101ff", ErrInvalidMaterial},
		{"issuer child nonminimal length", "a10482810161", ErrInvalidMaterial},
		{"valid excluded issuer", "a103820161", ErrUnsupportedCAProfile},
		{"valid excluded serial", "820101", ErrUnsupportedCAProfile},
	} {
		t.Run(v.name, func(t *testing.T) {
			body := append(bytes.Clone(keyID), fixtureHex(v.field)...)
			in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
			fixtureCert(t, in) // The pinned front parser ignores excluded fields here.
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}

func TestCAReviewDeclaredSPKIConsumption(t *testing.T) {
	ed := fixtureHex("302a300506032b6570032100d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a")
	for _, v := range []struct {
		name string
		der  []byte
		want error
	}{
		{"valid unsupported", ed, ErrUnsupportedCAProfile},
		{"nested NULL tail", fixtureTLV(0x30, append(bytes.Clone(ed[2:]), 5, 0)), ErrInvalidInput},
		{"nested tagged tail", fixtureTLV(0x30, append(bytes.Clone(ed[2:]), 0xa0, 0)), ErrInvalidInput},
	} {
		t.Run(v.name, func(t *testing.T) {
			if _, e := x509.ParsePKIXPublicKey(v.der); e != nil {
				t.Fatal("front key parser must accept", e)
			}
			in := fixtureInput(false)
			in.DeclaredSPKI = bytes.Clone(v.der)
			requireCAError(t, context.Background(), in, v.want)
			wireWant := v.want
			if errors.Is(wireWant, ErrInvalidInput) {
				wireWant = ErrInvalidMaterial
			}
			if _, e := canonicalSPKI(v.der, ErrInvalidMaterial); !errors.Is(e, wireWant) {
				t.Fatalf("CA SPKI got %v want %v", e, wireWant)
			}
		})
	}
}
func TestCAReviewUniqueIDConsumption(t *testing.T) {
	base := fixtureInput(false)
	fields := fixtureParts(t, fixtureParts(t, base.DER)[0])
	for _, v := range []struct {
		name, uid, tail string
		want            error
	}{
		{"valid issuer UID", "810100", "", ErrUnsupportedCAProfile},
		{"valid subject UID", "820100", "", ErrUnsupportedCAProfile},
		{"invalid unused count", "810108", "", ErrInvalidMaterial},
		{"empty UID", "8200", "", ErrInvalidMaterial},
		{"invalid padding", "81020103", "", ErrInvalidMaterial},
		{"TBS tail after valid UID", "810100", "0500", ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			parts := append([][]byte{}, fields[:7]...)
			parts = append(parts, fixtureHex(v.uid))
			parts = append(parts, fields[7:]...)
			if v.tail != "" {
				parts = append(parts, fixtureHex(v.tail))
			}
			in := fixtureRewrite(t, base, "tbs", fixtureJoin(parts))
			fixtureCert(t, in)
			if _, e := parseRawCA(in.DER); !errors.Is(e, v.want) {
				t.Errorf("raw scanner got %v want %v", e, v.want)
			}
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}
func TestCAReviewAKISerialConsumption(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name, serial string
		want         error
	}{
		{"valid excluded serial", "820101", ErrUnsupportedCAProfile},
		{"empty serial", "8200", ErrInvalidMaterial},
		{"nonminimal positive serial", "82020001", ErrInvalidMaterial},
		{"nonminimal negative serial", "8202ffff", ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			body := append(bytes.Clone(keyID), fixtureHex(v.serial)...)
			in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
			fixtureCert(t, in)
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}

func TestCAReviewKnownKeyPayloadConsumption(t *testing.T) {
	// DSA is parsed by the pinned toolchain but intentionally not marshalable.
	// Positive, small integers here are encoding controls, not usable key policy.
	dsaAlgorithm := fixtureHex("301406072a8648ce380401300902011702010b020102")
	dsa := fixtureJoin([][]byte{dsaAlgorithm, fixtureTLV(3, fixtureHex("00020108"))})
	dsaParamsTail := fixtureJoin([][]byte{fixtureHex("301606072a8648ce380401300b02011702010b0201020500"), fixtureTLV(3, fixtureHex("00020108"))})
	for _, v := range []struct {
		name string
		der  []byte
		want error
	}{
		{"valid unsupported RSA", fixtureHex("301b300d06092a864886f70d0101010500030a00300702020ca1020111"), ErrUnsupportedCAProfile},
		{"RSA payload sequence tail", fixtureHex("301d300d06092a864886f70d0101010500030c00300902020ca10201110500"), ErrInvalidInput},
		{"RSA payload outer tail", fixtureHex("301d300d06092a864886f70d0101010500030c00300702020ca10201110500"), ErrInvalidInput},
		{"valid unsupported DSA", dsa, ErrUnsupportedCAProfile},
		{"DSA payload tail", fixtureJoin([][]byte{dsaAlgorithm, fixtureTLV(3, fixtureHex("000201080500"))}), ErrInvalidInput},
		{"DSA parameter sequence tail", dsaParamsTail, ErrInvalidInput},
	} {
		t.Run(v.name, func(t *testing.T) {
			if _, e := x509.ParsePKIXPublicKey(v.der); e != nil {
				t.Fatal("front parser must accept", e)
			}
			in := fixtureInput(false)
			in.DeclaredSPKI = bytes.Clone(v.der)
			requireCAError(t, context.Background(), in, v.want)
			wireWant := v.want
			if errors.Is(wireWant, ErrInvalidInput) {
				wireWant = ErrInvalidMaterial
			}
			if _, e := canonicalSPKI(v.der, ErrInvalidMaterial); !errors.Is(e, wireWant) {
				t.Fatalf("CA SPKI got %v want %v", e, wireWant)
			}
		})
	}
}

func TestCAReviewAKIRegisteredIDConsumption(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name, issuer string
		want         error
	}{
		{"valid excluded registeredID", "a10388012a", ErrUnsupportedCAProfile},
		{"nonminimal first arc", "a1048802802a", ErrInvalidMaterial},
		{"nonminimal later arc", "a10588032a8001", ErrInvalidMaterial},
		{"unterminated arc", "a10488022a81", ErrInvalidMaterial},
		{"empty OID", "a1028800", ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			body := append(bytes.Clone(keyID), fixtureHex(v.issuer)...)
			in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
			fixtureCert(t, in)
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}

func TestCAReviewGeneralNamesGrammar(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name, suffix string
		want         error
	}{
		{"valid rfc822Name", "a103810161", ErrUnsupportedCAProfile},
		{"invalid rfc822Name IA5", "a1038101ff", ErrInvalidMaterial},
		{"valid DNS", "a103820161", ErrUnsupportedCAProfile},
		{"invalid DNS IA5", "a1038201ff", ErrInvalidMaterial},
		{"valid URI", "a103860161", ErrUnsupportedCAProfile},
		{"invalid URI IA5", "a1038601ff", ErrInvalidMaterial},
		{"valid directoryName", "a110a40e300c310a300806035504030c0161", ErrUnsupportedCAProfile},
		{"directoryName wrong type", "a104a4020500", ErrInvalidMaterial},
		{"directoryName descending RDN", "a11aa418301631143008060355040b0c0162300806035504030c0161", ErrInvalidMaterial},
		{"directoryName ATV tail", "a112a410300e310c300a06035504030c01610500", ErrInvalidMaterial},
		{"directoryName invalid UTF8", "a110a40e300c310a300806035504030c01ff", ErrInvalidMaterial},
		{"valid otherName", "a109a00706012aa0020500", ErrUnsupportedCAProfile},
		{"otherName missing value", "a107a00506032a0304", ErrInvalidMaterial},
		{"otherName duplicate value", "a10da00b06032a0304a00405000500", ErrInvalidMaterial},
		{"otherName empty value", "a107a00506012aa000", ErrInvalidMaterial},
		{"valid ediPartyName", "a107a505a1030c0161", ErrUnsupportedCAProfile},
		{"valid edi nameAssigner", "a10ca50aa0030c0162a1030c0161", ErrUnsupportedCAProfile},
		{"ediPartyName missing party", "a102a500", ErrInvalidMaterial},
		{"ediPartyName invalid UTF8", "a107a505a1030c01ff", ErrInvalidMaterial},
		{"ediPartyName trailing field", "a109a507a1030c01610500", ErrInvalidMaterial},
		{"valid X400", "a104a3023000", ErrUnsupportedCAProfile},
		{"X400 missing builtins", "a102a300", ErrInvalidMaterial},
		{"X400 tail", "a106a30430000500", ErrInvalidMaterial},
		{"X400 invalid implicit numeric", "a107a3053003800161", ErrInvalidMaterial},
		{"valid X400 numeric", "a107a3053003800131", ErrUnsupportedCAProfile},
		{"valid IP", "a106870401020304", ErrUnsupportedCAProfile},
	} {
		t.Run(v.name, func(t *testing.T) {
			body := append(bytes.Clone(keyID), fixtureHex(v.suffix)...)
			in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
			fixtureCert(t, in)
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}
func TestCAReviewOpaqueOIDRepresentation(t *testing.T) {
	for _, v := range []struct {
		name, parameter string
		want            error
	}{
		{"large minimal component", "06062a8880808000", nil},
		{"nonminimal component", "06032a8001", ErrInvalidMaterial},
		{"unterminated component", "06022a81", ErrInvalidMaterial},
		{"empty OID", "0600", ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			alg := fixtureTLV(0x30, append(fixtureHex("06032a0304"), fixtureHex(v.parameter)...))
			in := fixtureRewrite(t, fixtureRewrite(t, fixtureInput(false), "innerAlgorithm", alg), "outerAlgorithm", alg)
			fixtureCert(t, in)
			if v.want != nil {
				requireCAError(t, context.Background(), in, v.want)
				return
			}
			m, e := Parse(context.Background(), in)
			if e != nil || m == nil {
				t.Fatal("valid unknown parameter rejected", e)
			}
			defer m.Clear()
			s := m.Summary()
			if s.Proof != NodeCAMaterialShapeAndDeclaredKeyBound || s.Pending != expectedPending() {
				t.Fatal("partial proof changed", s)
			}
		})
	}
}

func TestCAReviewX400TypedAttributes(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name  string
		id    byte
		value string
		want  error
	}{
		{"commonName valid", 1, "130161", ErrUnsupportedCAProfile},
		{"commonName wrong type", 1, "0500", ErrInvalidMaterial},
		{"teletex valid", 2, "140161", ErrUnsupportedCAProfile},
		{"teletex wrong type", 2, "0500", ErrInvalidMaterial},
		{"personalName valid", 4, "3103800161", ErrUnsupportedCAProfile},
		{"personalName missing surname", 4, "3100", ErrInvalidMaterial},
		{"unit names valid", 5, "3003140161", ErrUnsupportedCAProfile},
		{"unit wrong type", 5, "30020500", ErrInvalidMaterial},
		{"domain attrs valid", 6, "30083006140161140162", ErrUnsupportedCAProfile},
		{"domain missing value", 6, "30053003140161", ErrInvalidMaterial},
		{"country valid", 8, "13025553", ErrUnsupportedCAProfile},
		{"country wrong type", 8, "0500", ErrInvalidMaterial},
		{"PDS valid", 10, "3103130161", ErrUnsupportedCAProfile},
		{"PDS duplicate Printable", 10, "3106130161130162", ErrInvalidMaterial},
		{"postal valid", 16, "31053003130161", ErrUnsupportedCAProfile},
		{"postal wrong line", 16, "310430020500", ErrInvalidMaterial},
		{"E163 valid", 22, "3003800131", ErrUnsupportedCAProfile},
		{"E163 missing number", 22, "3000", ErrInvalidMaterial},
		{"presentation valid", 22, "a007a3053103040101", ErrUnsupportedCAProfile},
		{"presentation missing addresses", 22, "a000", ErrInvalidMaterial},
		{"terminal valid", 23, "020103", ErrUnsupportedCAProfile},
		{"terminal wrong type", 23, "0500", ErrInvalidMaterial},
		{"unknown attr opaque", 24, "0500", ErrUnsupportedCAProfile},
	} {
		t.Run(v.name, func(t *testing.T) {
			attr := fixtureTLV(0x30, append(fixtureTLV(0x80, []byte{v.id}), fixtureTLV(0xa1, fixtureHex(v.value))...))
			address := fixtureTLV(0xa3, append(fixtureHex("3000"), fixtureTLV(0x31, attr)...))
			body := append(bytes.Clone(keyID), fixtureTLV(0xa1, address)...)
			in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
			fixtureCert(t, in)
			requireCAError(t, context.Background(), in, v.want)
		})
	}
}

func TestCAReviewNameValueTypes(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name, value string
		want        error
	}{
		{"UTF8", "0c0161", ErrUnsupportedCAProfile},
		{"Printable compatibility", "13022a26", ErrUnsupportedCAProfile},
		{"T61 compatibility", "1401ff", ErrUnsupportedCAProfile},
		{"BMP", "1e020061", ErrUnsupportedCAProfile},
		{"IA5", "160161", ErrUnsupportedCAProfile},
		{"Numeric", "120131", ErrUnsupportedCAProfile},
		{"NULL", "0500", ErrInvalidMaterial},
		{"INTEGER", "020101", ErrInvalidMaterial},
		{"OCTET STRING", "0400", ErrInvalidMaterial},
		{"SEQUENCE", "3000", ErrInvalidMaterial},
		{"context value", "8000", ErrInvalidMaterial},
		{"VisibleString not pinned Name", "1a0161", ErrInvalidMaterial},
		{"UniversalString not pinned Name", "1c0400000061", ErrInvalidMaterial},
		{"UTF8 malformed", "0c01ff", ErrInvalidMaterial},
		{"Printable malformed", "130140", ErrInvalidMaterial},
		{"BMP odd", "1e0100", ErrInvalidMaterial},
		{"BMP surrogate", "1e02d800", ErrInvalidMaterial},
		{"IA5 malformed", "1601ff", ErrInvalidMaterial},
		{"Numeric malformed", "120161", ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			// No CN/OID-specific subtype policy: repeat with an unbounded opaque OID.
			for _, oid := range []string{"0603550403", "06062a8880808000"} {
				atv := fixtureTLV(0x30, append(fixtureHex(oid), fixtureHex(v.value)...))
				name := fixtureTLV(0xa4, fixtureTLV(0x30, fixtureTLV(0x31, atv)))
				body := append(bytes.Clone(keyID), fixtureTLV(0xa1, name)...)
				in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
				fixtureCert(t, in)
				requireCAError(t, context.Background(), in, v.want)
			}
		})
	}
}
func TestCAReviewExtensionANYGrammar(t *testing.T) {
	keyID := fixtureHex("80205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3")
	for _, v := range []struct {
		name  string
		value []byte
		want  error
	}{
		{"Visible valid", fixtureHex("1a0161"), ErrUnsupportedCAProfile},
		{"Visible boundaries", fixtureHex("1a02207e"), ErrUnsupportedCAProfile},
		{"Visible high byte", fixtureHex("1a01ff"), ErrInvalidMaterial},
		{"Visible control", fixtureHex("1a011f"), ErrInvalidMaterial},
		{"Visible DEL", fixtureHex("1a017f"), ErrInvalidMaterial},
		{"relative OID valid", fixtureHex("0d0101"), ErrUnsupportedCAProfile},
		{"relative OID large", fixtureHex("0d058880808000"), ErrUnsupportedCAProfile},
		{"relative OID empty", fixtureHex("0d00"), ErrInvalidMaterial},
		{"relative OID unterminated", fixtureHex("0d0180"), ErrInvalidMaterial},
		{"relative OID nonminimal", fixtureHex("0d028001"), ErrInvalidMaterial},
		{"UTC valid", fixtureTLV(23, []byte("260101000000Z")), ErrUnsupportedCAProfile},
		{"UTC empty", fixtureHex("1700"), ErrInvalidMaterial},
		{"UTC bad date", fixtureTLV(23, []byte("260231000000Z")), ErrInvalidMaterial},
		{"UTC missing seconds", fixtureTLV(23, []byte("2601010000Z")), ErrInvalidMaterial},
		{"UTC offset", fixtureTLV(23, []byte("260101000000+0000")), ErrInvalidMaterial},
		{"UTC fraction", fixtureTLV(23, []byte("260101000000.1Z")), ErrInvalidMaterial},
		{"Generalized valid", fixtureTLV(24, []byte("20260101000000Z")), ErrUnsupportedCAProfile},
		{"Generalized fraction", fixtureTLV(24, []byte("20260101000000.123456789123Z")), ErrUnsupportedCAProfile},
		{"Generalized empty", fixtureHex("1800"), ErrInvalidMaterial},
		{"Generalized bad date", fixtureTLV(24, []byte("20260231000000Z")), ErrInvalidMaterial},
		{"Generalized missing seconds", fixtureTLV(24, []byte("202601010000Z")), ErrInvalidMaterial},
		{"Generalized offset", fixtureTLV(24, []byte("20260101000000+0000")), ErrInvalidMaterial},
		{"Generalized fractional zero", fixtureTLV(24, []byte("20260101000000.0Z")), ErrInvalidMaterial},
		{"Generalized fractional trailing zero", fixtureTLV(24, []byte("20260101000000.10Z")), ErrInvalidMaterial},
		{"TIME invalid UTF8", fixtureHex("0e01ff"), ErrInvalidMaterial},
		{"TIME overlong UTF8", fixtureHex("0e02c080"), ErrInvalidMaterial},
		{"TIME surrogate UTF8", fixtureHex("0e03eda080"), ErrInvalidMaterial},
		{"TIME out of range UTF8", fixtureHex("0e04f4908080"), ErrInvalidMaterial},
		{"TIME empty", fixtureHex("0e00"), ErrInvalidMaterial},
		{"TIME valid 0", fixtureTLV(14, []byte("20C")), ErrUnsupportedCAProfile},
		{"TIME valid 1", fixtureTLV(14, []byte("2026")), ErrUnsupportedCAProfile},
		{"TIME valid 2", fixtureTLV(14, []byte("2026-10")), ErrUnsupportedCAProfile},
		{"TIME valid 3", fixtureTLV(14, []byte("2024-02-29")), ErrUnsupportedCAProfile},
		{"TIME valid 4", fixtureTLV(14, []byte("2024-060")), ErrUnsupportedCAProfile},
		{"TIME valid 5", fixtureTLV(14, []byte("2026-W01")), ErrUnsupportedCAProfile},
		{"TIME valid 6", fixtureTLV(14, []byte("2026-W01-1")), ErrUnsupportedCAProfile},
		{"TIME valid 7", fixtureTLV(14, []byte("0000-01-01")), ErrUnsupportedCAProfile},
		{"TIME valid 8", fixtureTLV(14, []byte("-0001-01-01")), ErrUnsupportedCAProfile},
		{"TIME valid 9", fixtureTLV(14, []byte("+10000-01-01")), ErrUnsupportedCAProfile},
		{"TIME valid 10", fixtureTLV(14, []byte("-01C")), ErrUnsupportedCAProfile},
		{"TIME valid 11", fixtureTLV(14, []byte("+100C")), ErrUnsupportedCAProfile},
		{"TIME valid 12", fixtureTLV(14, []byte("12")), ErrUnsupportedCAProfile},
		{"TIME valid 13", fixtureTLV(14, []byte("12:30")), ErrUnsupportedCAProfile},
		{"TIME valid 14", fixtureTLV(14, []byte("12:30:45")), ErrUnsupportedCAProfile},
		{"TIME valid 15", fixtureTLV(14, []byte("12.00")), ErrUnsupportedCAProfile},
		{"TIME valid 16", fixtureTLV(14, []byte("24:00:00")), ErrUnsupportedCAProfile},
		{"TIME valid 17", fixtureTLV(14, []byte("12:30:45.0000000000")), ErrUnsupportedCAProfile},
		{"TIME valid 18", fixtureTLV(14, []byte("12:30Z")), ErrUnsupportedCAProfile},
		{"TIME valid 19", fixtureTLV(14, []byte("12:30+05")), ErrUnsupportedCAProfile},
		{"TIME valid 20", fixtureTLV(14, []byte("12:30-05:30")), ErrUnsupportedCAProfile},
		{"TIME valid 21", fixtureTLV(14, []byte("2016-12-31T23:59:60Z")), ErrUnsupportedCAProfile},
		{"TIME valid 22", fixtureTLV(14, []byte("P1Y2M3DT4H5M6S")), ErrUnsupportedCAProfile},
		{"TIME valid 23", fixtureTLV(14, []byte("P2W")), ErrUnsupportedCAProfile},
		{"TIME valid 24", fixtureTLV(14, []byte("P0D")), ErrUnsupportedCAProfile},
		{"TIME valid 25", fixtureTLV(14, []byte("PT0.00S")), ErrUnsupportedCAProfile},
		{"TIME valid 26", fixtureTLV(14, []byte("P0.00Y")), ErrUnsupportedCAProfile},
		{"TIME valid 27", fixtureTLV(14, []byte("P29MT0S")), ErrUnsupportedCAProfile},
		{"TIME valid 28", fixtureTLV(14, []byte("2026-01-01/2026-01-02")), ErrUnsupportedCAProfile},
		{"TIME valid 29", fixtureTLV(14, []byte("2026-01-01/P1D")), ErrUnsupportedCAProfile},
		{"TIME valid 30", fixtureTLV(14, []byte("P1D/2026-01-02")), ErrUnsupportedCAProfile},
		{"TIME valid 31", fixtureTLV(14, []byte("R/P1D")), ErrUnsupportedCAProfile},
		{"TIME valid 32", fixtureTLV(14, []byte("R123456789012345678901234567890/P1D")), ErrUnsupportedCAProfile},
		{"TIME valid 33", fixtureTLV(14, []byte("2026-01-01T12:00+05/2026-01-02T12:00")), ErrUnsupportedCAProfile},
		{"TIME valid 34", fixtureTLV(14, []byte("12:00+05/13:00+06")), ErrUnsupportedCAProfile},
		{"TIME valid 35", fixtureTLV(14, []byte("2026-01-01T12:00Z/2026-01-02T12:00Z")), ErrUnsupportedCAProfile},
		{"TIME unbounded year", fixtureTLV(14, []byte("+"+strings.Repeat("1", 80)+"-01-01")), ErrUnsupportedCAProfile},
		{"TIME unbounded duration", fixtureTLV(14, []byte("PT"+strings.Repeat("9", 80)+"S")), ErrUnsupportedCAProfile},
		{"TIME long fraction accuracy", fixtureTLV(14, []byte("12:30:45."+strings.Repeat("0", 80)+"Z")), ErrUnsupportedCAProfile},
		{"TIME ordinal leap date-time", fixtureTLV(14, []byte("2024-060T12:30")), ErrUnsupportedCAProfile},
		{"TIME week date-time", fixtureTLV(14, []byte("2026-W01-1T12:30")), ErrUnsupportedCAProfile},
		{"TIME negative leap calendar", fixtureTLV(14, []byte("-0400-02-29")), ErrUnsupportedCAProfile},
		{"TIME expanded leap calendar", fixtureTLV(14, []byte("+10000-02-29")), ErrUnsupportedCAProfile},
		{"TIME end midnight fraction accuracy", fixtureTLV(14, []byte("24:00:00.00")), ErrUnsupportedCAProfile},
		{"TIME UTC interval", fixtureTLV(14, []byte("12:00Z/13:00Z")), ErrUnsupportedCAProfile},
		{"TIME positive offset bound", fixtureTLV(14, []byte("12:00+15")), ErrUnsupportedCAProfile},
		{"TIME negative offset bound", fixtureTLV(14, []byte("12:00-15")), ErrUnsupportedCAProfile},
		{"TIME nonzero offset minute", fixtureTLV(14, []byte("12:00+14:59")), ErrUnsupportedCAProfile},
		{"TIME nonmidnight 24 fraction", fixtureTLV(14, []byte("24:00:00.1")), ErrInvalidMaterial},
		{"TIME nonleap ordinal date", fixtureTLV(14, []byte("2023-366")), ErrInvalidMaterial},
		{"TIME nonexistent week 53", fixtureTLV(14, []byte("2025-W53")), ErrInvalidMaterial},
		{"TIME negative year zero", fixtureTLV(14, []byte("-0000-01-01")), ErrInvalidMaterial},
		{"TIME negative nonleap date", fixtureTLV(14, []byte("-0001-02-29")), ErrInvalidMaterial},
		{"TIME duplicate duration unit", fixtureTLV(14, []byte("PT1H1H")), ErrInvalidMaterial},
		{"TIME reversed duration units", fixtureTLV(14, []byte("P1D1Y")), ErrInvalidMaterial},
		{"TIME mixed endpoint precision", fixtureTLV(14, []byte("12:00/13:00:00")), ErrInvalidMaterial},
		{"TIME mixed date format", fixtureTLV(14, []byte("2026-01-01/2026-002")), ErrInvalidMaterial},
		{"TIME omitted UTC end suffix", fixtureTLV(14, []byte("12:00Z/13:00")), ErrInvalidMaterial},
		{"TIME duplicate separators", fixtureTLV(14, []byte("12:30:45..1")), ErrInvalidMaterial},
		{"TIME malformed 0", fixtureTLV(14, []byte("\"2026\"")), ErrInvalidMaterial},
		{"TIME malformed 1", fixtureTLV(14, []byte("2026  ")), ErrInvalidMaterial},
		{"TIME malformed 2", fixtureTLV(14, []byte("é")), ErrInvalidMaterial},
		{"TIME malformed 3", fixtureTLV(14, []byte("P")), ErrInvalidMaterial},
		{"TIME malformed 4", fixtureTLV(14, []byte("PT")), ErrInvalidMaterial},
		{"TIME malformed 5", fixtureTLV(14, []byte("P1DT")), ErrInvalidMaterial},
		{"TIME malformed 6", fixtureTLV(14, []byte("P01D")), ErrInvalidMaterial},
		{"TIME malformed 7", fixtureTLV(14, []byte("P1W2D")), ErrInvalidMaterial},
		{"TIME malformed 8", fixtureTLV(14, []byte("PT1.2H1M")), ErrInvalidMaterial},
		{"TIME malformed 9", fixtureTLV(14, []byte("PT0H1S")), ErrInvalidMaterial},
		{"TIME malformed 10", fixtureTLV(14, []byte("P0Y1D")), ErrInvalidMaterial},
		{"TIME malformed 11", fixtureTLV(14, []byte("12:30,5Z")), ErrInvalidMaterial},
		{"TIME malformed 12", fixtureTLV(14, []byte("12:30+05:00")), ErrInvalidMaterial},
		{"TIME malformed 13", fixtureTLV(14, []byte("12:30+05/13:30+05")), ErrInvalidMaterial},
		{"TIME malformed 14", fixtureTLV(14, []byte("2026-02-30")), ErrInvalidMaterial},
		{"TIME malformed 15", fixtureTLV(14, []byte("20260101")), ErrInvalidMaterial},
		{"TIME malformed 16", fixtureTLV(14, []byte("25:00")), ErrInvalidMaterial},
		{"TIME malformed 17", fixtureTLV(14, []byte("12:60")), ErrInvalidMaterial},
		{"TIME malformed 18", fixtureTLV(14, []byte("12:30:61")), ErrInvalidMaterial},
		{"TIME malformed 19", fixtureTLV(14, []byte("12:30z")), ErrInvalidMaterial},
		{"TIME malformed 20", fixtureTLV(14, []byte("12:30.")), ErrInvalidMaterial},
		{"TIME malformed 21", fixtureTLV(14, []byte("R1/2026")), ErrInvalidMaterial},
		{"TIME malformed 22", fixtureTLV(14, []byte("2026/P1D/P2D")), ErrInvalidMaterial},
		{"Generalized leap second", fixtureTLV(24, []byte("20161231235960Z")), ErrUnsupportedCAProfile},
		{"Generalized leap fraction", fixtureTLV(24, []byte("20161231235960.123456789123Z")), ErrUnsupportedCAProfile},
		{"Generalized second 61", fixtureTLV(24, []byte("20161231235961Z")), ErrInvalidMaterial},
		{"Generalized leap impossible date", fixtureTLV(24, []byte("20160230235960Z")), ErrInvalidMaterial},
		{"UTC leap still invalid", fixtureTLV(23, []byte("161231235960Z")), ErrInvalidMaterial},
		{"EXTERNAL explicit value", fixtureHex("2804a0020500"), ErrUnsupportedCAProfile},
		{"EXTERNAL octets", fixtureHex("28028100"), ErrUnsupportedCAProfile},
		{"EXTERNAL bits", fixtureHex("2803820100"), ErrUnsupportedCAProfile},
		{"EXTERNAL references", fixtureHex("280b06012a0201010701618100"), ErrUnsupportedCAProfile},
		{"EXTERNAL empty", fixtureHex("2800"), ErrInvalidMaterial},
		{"EXTERNAL reference only", fixtureHex("280306012a"), ErrInvalidMaterial},
		{"EXTERNAL explicit empty", fixtureHex("2802a000"), ErrInvalidMaterial},
		{"EXTERNAL explicit duplicate", fixtureHex("2806a00405000500"), ErrInvalidMaterial},
		{"EXTERNAL wrong choice", fixtureHex("28028000"), ErrInvalidMaterial},
		{"EXTERNAL malformed implicit bits", fixtureHex("2803820108"), ErrInvalidMaterial},
		{"EXTERNAL tail", fixtureHex("280481000500"), ErrInvalidMaterial},
		{"REAL zero", fixtureHex("0900"), ErrUnsupportedCAProfile},
		{"REAL binary", fixtureHex("0903800001"), ErrUnsupportedCAProfile},
		{"REAL unbounded exponent", fixtureHex("090783040100000001"), ErrUnsupportedCAProfile},
		{"REAL decimal", fixtureTLV(9, append([]byte{3}, []byte("1.E+0")...)), ErrUnsupportedCAProfile},
		{"REAL special", fixtureHex("090142"), ErrUnsupportedCAProfile},
		{"REAL reserved base", fixtureHex("0903b00001"), ErrInvalidMaterial},
		{"REAL nonzero scaling", fixtureHex("0903840001"), ErrInvalidMaterial},
		{"REAL nonminimal exponent length", fixtureHex("0906830301000001"), ErrInvalidMaterial},
		{"REAL leading zero mantissa", fixtureHex("090480000001"), ErrInvalidMaterial},
		{"REAL missing mantissa", fixtureHex("09028000"), ErrInvalidMaterial},
		{"REAL nonminimal exponent", fixtureHex("090481000001"), ErrInvalidMaterial},
		{"REAL even mantissa", fixtureHex("0903800002"), ErrInvalidMaterial},
		{"REAL decimal noncanonical", fixtureTLV(9, append([]byte{3}, []byte("1.0E+0")...)), ErrInvalidMaterial},
		{"REAL decimal trailing zero", fixtureTLV(9, append([]byte{3}, []byte("10.E+0")...)), ErrInvalidMaterial},
		{"PDV fixed", fixtureHex("2b06a00285008200"), ErrUnsupportedCAProfile},
		{"PDV syntaxes", fixtureHex("2b0ca008a00680012a81012a8200"), ErrUnsupportedCAProfile},
		{"PDV syntax", fixtureHex("2b07a00381012a8200"), ErrUnsupportedCAProfile},
		{"PDV presentation id", fixtureHex("2b07a0038201018200"), ErrUnsupportedCAProfile},
		{"PDV negotiation", fixtureHex("2b0ca008a30680010181012a8200"), ErrUnsupportedCAProfile},
		{"PDV transfer syntax", fixtureHex("2b07a00384012a8200"), ErrUnsupportedCAProfile},
		{"PDV syntaxes missing second", fixtureHex("2b09a005a00380012a8200"), ErrInvalidMaterial},
		{"PDV syntax nonminimal OID", fixtureHex("2b08a0048102802a8200"), ErrInvalidMaterial},
		{"PDV presentation id nonminimal", fixtureHex("2b08a004820200018200"), ErrInvalidMaterial},
		{"PDV negotiation missing syntax", fixtureHex("2b09a005a3038001018200"), ErrInvalidMaterial},
		{"character fixed", fixtureHex("3d06a00285008200"), ErrUnsupportedCAProfile},
		{"PDV empty", fixtureHex("2b00"), ErrInvalidMaterial},
		{"character empty", fixtureHex("3d00"), ErrInvalidMaterial},
		{"PDV missing data", fixtureHex("2b04a0028500"), ErrInvalidMaterial},
		{"PDV missing identification", fixtureHex("2b028200"), ErrInvalidMaterial},
		{"PDV malformed fixed", fixtureHex("2b07a0038501008200"), ErrInvalidMaterial},
		{"PDV descriptor prohibited", fixtureHex("2b09a00285008101618200"), ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			for _, mode := range []string{"otherName", "nested otherName", "unknown X400"} {
				value := bytes.Clone(v.value)
				if mode == "nested otherName" {
					value = fixtureTLV(0x30, value)
				}
				other := fixtureTLV(0xa0, append(fixtureHex("06012a"), fixtureTLV(0xa0, value)...))
				if mode == "unknown X400" {
					attr := fixtureTLV(0x30, append(fixtureHex("800118"), fixtureTLV(0xa1, value)...))
					other = fixtureTLV(0xa3, append(fixtureHex("3000"), fixtureTLV(0x31, attr)...))
				}
				body := append(bytes.Clone(keyID), fixtureTLV(0xa1, other)...)
				in := fixtureExt(t, fixtureInput(false), 35, func(e *pkix.Extension) { e.Value = fixtureTLV(0x30, body) })
				fixtureCert(t, in)
				beforeDER, beforeSPKI := bytes.Clone(in.DER), bytes.Clone(in.DeclaredSPKI)
				requireCAError(t, context.Background(), in, v.want)
				if !bytes.Equal(in.DER, beforeDER) || !bytes.Equal(in.DeclaredSPKI, beforeSPKI) {
					t.Fatal("rejected ANY mutated caller material")
				}
			}
		})
	}
}

func TestCAReviewCertificateValidityLeapStillInvalid(t *testing.T) {
	for _, v := range []struct {
		tag   byte
		value string
	}{
		{23, "161231235960Z"}, {24, "20161231235960Z"},
	} {
		in := fixtureInput(false)
		fields := fixtureParts(t, fixtureParts(t, in.DER)[0])
		times := fixtureParts(t, fields[4])
		times[0] = fixtureTLV(v.tag, []byte(v.value))
		fields[4] = fixtureJoin(times)
		in = fixtureRewrite(t, in, "tbs", fixtureJoin(fields))
		if _, e := x509.ParseCertificate(in.DER); e == nil {
			t.Fatal("pinned Go unexpectedly accepted certificate validity leap second")
		}
		requireCAError(t, context.Background(), in, ErrInvalidMaterial)
	}
}
