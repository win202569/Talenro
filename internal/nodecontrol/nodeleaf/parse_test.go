package nodeleaf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"math/big"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func wantPending() [6]Obligation {
	return [6]Obligation{LeafSignatureAndIssuerTrust, KeyIdentifierDerivationAndIssuerBinding, TrustedCurrentTime, IdentityAndAuthorityProvenance, PrivateKeyPossession, MainStateAndRuntimeBinding}
}

func TestLeafParse(t *testing.T) {
	in := leafFixture(t)
	m, e := Parse(context.Background(), in)
	if e != nil || m == nil {
		t.Fatal(e)
	}
	defer m.Clear()
	s := m.Summary()
	if s.Proof != NodeLeafMaterialShapeValid || s.Pending != wantPending() || s.NodeID != in.NodeID || s.TrustDomain != "node.test" || s.SerialLength != 1 || s.Serial[0] != 128 || !bytes.Equal(s.Serial[1:], make([]byte, 19)) || !s.NotBefore.Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)) || !s.NotAfter.Equal(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("summary")
	}
	if s.LeafDERDigest != contracts.Digest(sha256.Sum256(in.DER)) || s.SPKIDERDigest != contracts.Digest(sha256.Sum256(in.ExpectedSPKI)) || s.LeafDERDigest == s.SPKIDERDigest {
		t.Fatal("digest preimages")
	}
	out := m.DER()
	if !bytes.Equal(out, in.DER) {
		t.Fatal("exact DER")
	}
	out[0] ^= 1
	if !bytes.Equal(m.DER(), in.DER) {
		t.Fatal("DER escaped")
	}
}

func TestLeafURIBinding(t *testing.T) {
	for _, uri := range []string{
		"spiffe://NODE.test/node/11111111-2222-4333-8444-555555555555",
		"SPIFFE://node.test/node/11111111-2222-4333-8444-555555555555",
		"spiffe://node.test/node/%311111111-2222-4333-8444-555555555555",
		"spiffe://node.test:443/node/11111111-2222-4333-8444-555555555555",
		"spiffe://a@node.test/node/11111111-2222-4333-8444-555555555555",
		"spiffe://node.test/node/11111111-2222-4333-8444-555555555555?q=1",
		"spiffe://node.test/node/11111111-2222-4333-8444-555555555555#f",
		"spiffe://node.test/node/11111111-2222-4333-8444-555555555555/extra",
		"spiffe://node.test/node/11111111-2222-4333-8444-555555555555/",
	} {
		in := fixtureWithExtension(t, 3, fixtureTLV(0x30, fixtureTLV(0x86, []byte(uri))))
		_, parseErr := x509.ParseCertificate(in.DER)
		m, e := Parse(context.Background(), in)
		want := ErrBindingMismatch
		if parseErr != nil {
			want = ErrInvalidMaterial
		}
		if m != nil || !errors.Is(e, want) {
			t.Fatal("URI equivalence accepted", e)
		}
	}
}

func TestLeafSPKIBinding(t *testing.T) {
	in := leafFixture(t)
	other, e := x509.MarshalPKIXPublicKey(&fixtureKey(3).PublicKey)
	if e != nil {
		t.Fatal(e)
	}
	in.ExpectedSPKI = other
	m, e := Parse(context.Background(), in)
	if m != nil || !errors.Is(e, ErrBindingMismatch) {
		t.Fatal("different valid SPKI", e)
	}
	in.ExpectedSPKI = append(other, 0)
	m, e = Parse(context.Background(), in)
	if m != nil || !errors.Is(e, ErrInvalidInput) {
		t.Fatal("expected trailing", e)
	}
	in = leafFixture(t)
	pointSame := append([]byte{0x30, 0x81, in.ExpectedSPKI[1]}, in.ExpectedSPKI[2:]...)
	in = rewriteFixture(t, in, "spki", pointSame)
	if m, e := Parse(context.Background(), in); m != nil || e == nil {
		t.Fatal("same point nonminimal wire")
	}
}

func leafRejections(t testing.TB) []Input {
	t.Helper()
	in := leafFixture(t)
	cases := []Input{rewriteFixture(t, in, "certificate", nil), rewriteFixture(t, in, "certificate", []byte("-----BEGIN CERTIFICATE-----")), rewriteFixture(t, in, "certificate", append(bytes.Clone(in.DER), 0)), rewriteFixture(t, in, "certificate", append(bytes.Clone(in.DER), in.DER...))}
	for _, serial := range [][]byte{{2, 1, 0}, {2, 1, 255}, {2, 2, 0, 1}, fixtureTLV(2, append([]byte{0}, bytes.Repeat([]byte{0x80}, 20)...))} {
		cases = append(cases, rewriteFixture(t, in, "serial", serial))
	}
	for _, subject := range [][]byte{{0x30, 2, 0x31, 0}, {0x31, 0}} {
		cases = append(cases, rewriteFixture(t, in, "subject", subject))
	}
	cases = append(cases, leafWithTemplate(t, func(c *x509.Certificate) { c.Subject = pkix.Name{CommonName: "not empty"} }))
	for _, spki := range fixtureNonP256SPKI(t) {
		cases = append(cases, rewriteFixture(t, in, "spki", spki))
	}
	point := bytes.Clone(in.ExpectedSPKI)
	clear(point[len(point)-64:])
	cases = append(cases, rewriteFixture(t, in, "spki", point))
	for _, version := range [][]byte{nil, {0xa0, 3, 2, 1, 1}, {0xa0, 3, 2, 1, 3}, {0xa0, 3, 4, 1, 2}} {
		v := rewriteFixture(t, in, "extensions", nil)
		cases = append(cases, rewriteFixture(t, v, "version", version))
	}
	for _, raw := range [][]byte{fixtureTLV(0x17, []byte("261011000000+0000")), fixtureTLV(0x18, []byte("20261011000000.1Z")), fixtureTLV(0x18, []byte("20261011000000Z")), fixtureTLV(0x18, []byte("00001011000000Z")), fixtureTLV(0x16, []byte("261011000000Z"))} {
		cases = append(cases, rewriteFixture(t, in, "notBefore", raw))
	}
	for _, delta := range []time.Duration{0, -time.Second, 24*time.Hour - time.Second, 24*time.Hour + time.Second} {
		d := delta
		cases = append(cases, leafWithTemplate(t, func(c *x509.Certificate) { c.NotAfter = c.NotBefore.Add(d) }))
	}
	cert := fixtureParts(t, in.DER)
	fields := fixtureParts(t, cert[0])
	fields = append(fields[:7], append([][]byte{{0x81, 1, 0}}, fields[7:]...)...)
	cases = append(cases, rewriteFixture(t, in, "tbs", fixtureJoin(fields)))
	cert = fixtureParts(t, in.DER)
	fields = fixtureParts(t, cert[0])
	cases = append(cases, rewriteFixture(t, in, "tbs", fixtureJoin(append(fields, []byte{5, 0}))))
	for _, tc := range leafExtensionCases() {
		cases = append(cases, fixtureWithExtension(t, tc.index, tc.value))
	}
	ext := fixtureExtensions()
	for i := range ext {
		missing := append([]pkix.Extension(nil), ext[:i]...)
		missing = append(missing, ext[i+1:]...)
		duplicate := append(append([]pkix.Extension(nil), ext...), ext[i])
		wrong := append([]pkix.Extension(nil), ext...)
		wrong[i].Critical = !wrong[i].Critical
		for _, items := range [][]pkix.Extension{missing, duplicate, wrong} {
			cases = append(cases, rewriteFixture(t, in, "extensions", fixtureExtensionBlock(items, false)))
		}
	}
	for _, critical := range []bool{false, true} {
		extra := append(append([]pkix.Extension(nil), ext...), pkix.Extension{Id: asn1.ObjectIdentifier{1, 2, 3, 4}, Critical: critical, Value: []byte{5, 0}})
		cases = append(cases, rewriteFixture(t, in, "extensions", fixtureExtensionBlock(extra, false)))
	}
	cases = append(cases, rewriteFixture(t, in, "extensions", fixtureExtensionBlock(ext, true)))
	for _, arm := range [][]byte{{0xa0, 7, 6, 2, 0x2a, 3, 0xa0, 1, 0}, {0x88, 2, 0x2a, 3}, fixtureTLV(0x82, []byte("node.test")), []byte{0x87, 4, 127, 0, 0, 1}, fixtureTLV(0x81, []byte("a@node.test"))} {
		uri := fixtureParts(t, ext[3].Value)[0]
		cases = append(cases, fixtureWithExtension(t, 3, fixtureTLV(0x30, append(uri, arm...))))
	}
	return cases
}

func TestLeafParseRejections(t *testing.T) {
	for i, in := range leafRejections(t) {
		m, e := Parse(context.Background(), in)
		if m != nil || e == nil {
			t.Fatal("rejection accepted", i)
		}
		if !errors.Is(e, ErrInvalidInput) && !errors.Is(e, ErrInvalidMaterial) && !errors.Is(e, ErrUnsupportedLeafProfile) {
			t.Fatal("unexpected category", i, e)
		}
		if e != ErrInvalidInput && e != ErrInvalidMaterial && e != ErrUnsupportedLeafProfile {
			t.Fatal("parser detail escaped")
		}
	}
}

func TestLeafPartialProofOnly(t *testing.T) {
	cases := []Input{}
	for _, year := range []int{2000, 2090} {
		y := year
		cases = append(cases, leafWithTemplate(t, func(c *x509.Certificate) {
			c.NotBefore = time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
			c.NotAfter = c.NotBefore.Add(24 * time.Hour)
		}))
	}
	in := leafFixture(t)
	parts := fixtureParts(t, in.DER)
	sig := bytes.Clone(parts[2])
	sig[len(sig)-1] ^= 1
	bad := rewriteFixture(t, in, "signature", sig)
	c, e := x509.ParseCertificate(bad.DER)
	if e != nil {
		t.Fatal("signature syntax", e)
	}
	issuer := &x509.Certificate{PublicKey: &fixtureKey(1).PublicKey, PublicKeyAlgorithm: x509.ECDSA, BasicConstraintsValid: true, IsCA: true, KeyUsage: x509.KeyUsageCertSign}
	if e := c.CheckSignatureFrom(issuer); e == nil {
		t.Fatal("bad-signature control")
	}
	cases = append(cases, bad)
	cases = append(cases, leafWithTemplate(t, func(c *x509.Certificate) {
		c.ExtraExtensions[4].Value = fixtureTLV(4, []byte{99})
		c.ExtraExtensions[5].Value = fixtureTLV(0x30, fixtureTLV(0x80, []byte{100}))
	}))
	for _, in := range cases {
		m, e := Parse(context.Background(), in)
		if e != nil || m == nil {
			t.Fatal("shape-only case rejected", e)
		}
		s := m.Summary()
		m.Clear()
		if s.Proof != NodeLeafMaterialShapeValid || s.Pending != wantPending() {
			t.Fatal("trust enlarged")
		}
	}
}

func TestLeafNegativeSerialCompatibility(t *testing.T) {
	in := rewriteFixture(t, leafFixture(t), "serial", []byte{2, 1, 255})
	for _, setting := range []string{"x509negativeserial=0", "x509negativeserial=1"} {
		t.Run(setting, func(t *testing.T) {
			t.Setenv("GODEBUG", setting)
			m, e := Parse(context.Background(), in)
			if m != nil || e == nil {
				t.Fatal("negative accepted")
			}
			if e != ErrInvalidMaterial && e != ErrUnsupportedLeafProfile {
				t.Fatal(e)
			}
		})
	}
	// Parsed compatibility cannot remove our explicit positive-number condition.
	in = leafWithTemplate(t, func(c *x509.Certificate) { c.SerialNumber = big.NewInt(1) })
	if m, e := Parse(context.Background(), in); e != nil {
		t.Fatal(e)
	} else {
		m.Clear()
	}
}

type leafTraceContext struct {
	context.Context
	calls, fail int
	failure     error
	panicAt     bool
}

func (c *leafTraceContext) Err() error {
	c.calls++
	if c.calls == c.fail {
		if c.panicAt {
			panic("test context panic")
		}
		return c.failure
	}
	return nil
}

func TestLeafCancellation(t *testing.T) {
	in := leafFixture(t)
	if m, e := Parse(nil, in); m != nil || e != ErrInvalidInput {
		t.Fatal("nil context")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if m, e := Parse(ctx, in); m != nil || e != context.Canceled {
		t.Fatal("pre-cancel")
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if m, e := Parse(ctx, in); m != nil || e != context.DeadlineExceeded {
		t.Fatal("pre-deadline")
	}
}

func TestLeafLateCancellation(t *testing.T) {
	in := leafFixture(t)
	trace := &leafTraceContext{Context: context.Background()}
	m, e := Parse(trace, in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	if trace.calls < 4 {
		t.Fatal("missing phase checkpoints")
	}
	for point := 2; point <= trace.calls; point++ {
		for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
			ctx := &leafTraceContext{Context: context.Background(), fail: point, failure: failure}
			m, e := Parse(ctx, in)
			if m != nil || e != failure {
				t.Fatal("late cancellation", point, e)
			}
		}
		func() {
			ctx := &leafTraceContext{Context: context.Background(), fail: point, panicAt: true}
			seen := false
			defer func() {
				if recover() != nil {
					seen = true
				}
				if !seen {
					t.Fatal("panic swallowed", point)
				}
			}()
			m, e := Parse(ctx, in)
			if m != nil {
				m.Clear()
			}
			t.Fatalf("panic did not propagate at %d: %v", point, e)
		}()
	}
}

func TestLeafParseOwnership(t *testing.T) {
	in := leafFixture(t)
	original := bytes.Clone(in.DER)
	spkiOriginal := bytes.Clone(in.ExpectedSPKI)
	shared := append(bytes.Clone(in.DER), in.ExpectedSPKI...)
	in.DER = shared[:len(original)]
	in.ExpectedSPKI = shared[len(original):]
	m, e := Parse(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	derBacking, spkiBacking := m.owned.der, m.owned.expectedSPKI
	copyWrapper := *m
	detached := m.DER()
	s := m.Summary()
	clear(shared)
	if !bytes.Equal(m.DER(), original) || !bytes.Equal(m.owned.expectedSPKI, spkiOriginal) {
		t.Fatal("caller owns result")
	}
	s.Serial[0] = 0
	s.Pending[0] = UnknownObligation
	if m.Summary().Serial[0] != 128 || m.Summary().Pending != wantPending() {
		t.Fatal("summary alias")
	}
	copyWrapper.Clear()
	if m.DER() != nil || m.Summary() != (Summary{}) || !bytes.Equal(derBacking, make([]byte, len(derBacking))) || !bytes.Equal(spkiBacking, make([]byte, len(spkiBacking))) || !bytes.Equal(detached, original) {
		t.Fatal("shared clear")
	}
	m.Clear()
}
