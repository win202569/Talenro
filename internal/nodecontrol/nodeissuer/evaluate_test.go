package nodeissuer

import (
	"bytes"
	"context"
	"crypto/x509"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	"testing"
	"time"
)

func assertIssuerPartial(t testing.TB, s Summary) {
	t.Helper()
	if s.Proof != NodeLeafDirectIssuerSignatureBound || s.Leaf.Proof != nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound || s.Leaf.Leaf.Proof != nodeleaf.NodeLeafMaterialShapeValid || s.CA.Proof != nodeca.NodeCAMaterialShapeAndDeclaredKeyBound {
		t.Fatal("partial proof")
	}
	if s.Leaf.Leaf.Pending != ([6]nodeleaf.Obligation{nodeleaf.LeafSignatureAndIssuerTrust, nodeleaf.KeyIdentifierDerivationAndIssuerBinding, nodeleaf.TrustedCurrentTime, nodeleaf.IdentityAndAuthorityProvenance, nodeleaf.PrivateKeyPossession, nodeleaf.MainStateAndRuntimeBinding}) {
		t.Fatal("leaf pending")
	}
	if s.CA.Pending != ([6]nodeca.Obligation{nodeca.CertificateSignatureAndIssuerChain, nodeca.AuthorityIdentifierAndParentBinding, nodeca.TrustedCurrentTime, nodeca.DeploymentAuthorizationAndPurpose, nodeca.BundleStatusAndRollback, nodeca.LeafIssuerRelationAndRuntimeUse}) {
		t.Fatal("CA pending")
	}
}

// Catches wrong derived descriptions, incomplete obligations and S normalization.
func TestIssuerEvaluateGolden(t *testing.T) {
	for name, g := range fixtureGoldens(t) {
		t.Run(name, func(t *testing.T) {
			in := fixtureVariant(t, name)
			m, e := Evaluate(context.Background(), in)
			if e != nil {
				t.Fatal(e)
			}
			defer m.Clear()
			s := m.Summary()
			assertIssuerPartial(t, s)
			for field, b := range map[string][]byte{"CADERDigest": s.CA.CADERDigest[:], "CASPKIDigest": s.CA.SPKIDERDigest[:], "CANameDigest": s.CA.SubjectDERDigest[:], "CAIssuerDigest": s.CA.IssuerDERDigest[:], "LeafDERDigest": s.Leaf.Leaf.LeafDERDigest[:], "LeafSPKIDigest": s.Leaf.Leaf.SPKIDERDigest[:], "DeclaredSPKIDigest": s.Leaf.DeclaredIssuerSPKIDERDigest[:], "LeafTBSDigest": s.LeafTBSDERDigest[:], "LeafIssuerDigest": s.LeafIssuerDERDigest[:]} {
				if !bytes.Equal(b, hexBytes(t, g[field])) {
					t.Fatal("digest", field)
				}
			}
			for field, b := range map[string][]byte{"CASKI": s.CA.SubjectKeyIdentifier[:], "CAAKI": s.CA.AuthorityKeyIdentifier[:], "LeafSKI": s.Leaf.SubjectKeyIdentifier[:], "LeafAKI": s.Leaf.AuthorityKeyIdentifier[:]} {
				if !bytes.Equal(b, hexBytes(t, g[field])) {
					t.Fatal("identifier", field)
				}
			}
			caSerial := byte(128)
			if name == "sameCAKeyAndNameOtherCert" {
				caSerial = 130
			}
			var cs, ls [20]byte
			cs[0] = caSerial
			ls[0] = 129
			if s.CA.Serial != cs || s.CA.SerialLength != 1 || s.Leaf.Leaf.Serial != ls || s.Leaf.Leaf.SerialLength != 1 {
				t.Fatal("serial")
			}
			cb, ce, ly := 2026, 2036, 2026
			if name == "expired" {
				cb, ce, ly = 2010, 2011, 2010
			}
			if name == "future" {
				cb, ce, ly = 2040, 2050, 2041
			}
			if name == "crossCAWindow" {
				cb, ce = 2020, 2021
			}
			lb := time.Date(ly, 10, 12, 0, 0, 0, 0, time.UTC)
			if s.CA.NotBefore != time.Date(cb, 1, 1, 0, 0, 0, 0, time.UTC) || s.CA.NotAfter != time.Date(ce, 1, 1, 0, 0, 0, 0, time.UTC) || s.Leaf.Leaf.NotBefore != lb || s.Leaf.Leaf.NotAfter != lb.Add(24*time.Hour) {
				t.Fatal("fixed times")
			}
			if s.CA.HasPathLen != (name == "caPathZero") || s.CA.PathLen != 0 || s.CA.CRLSignAllowed != (name == "caCRL") || s.Leaf.Leaf.NodeID != in.Leaf.NodeID || s.Leaf.Leaf.TrustDomain != "node.test" {
				t.Fatal("metadata")
			}
			if !bytes.Equal(m.owned.leafDER, in.Leaf.DER) || !bytes.Equal(m.owned.caDER, in.CA.DER) {
				t.Fatal("exact bytes altered")
			}
			buffers := [][]byte{m.owned.leafDER, m.owned.expectedSPKI, m.owned.caDER, m.owned.declaredSPKI}
			m.Clear()
			if m.Summary() != (Summary{}) {
				t.Fatal("clear proof")
			}
			assertIssuerPartial(t, s)
			for _, b := range buffers {
				for _, v := range b {
					if v != 0 {
						t.Fatal("reachable buffer")
					}
				}
			}
		})
	}
}
func assertIssuerFailure(t *testing.T, in Input, want error) {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if m != nil {
		m.Clear()
		t.Fatal("failure returned material")
	}
	if e != want {
		t.Fatalf("error %v want %v", e, want)
	}
}
func wrongNameInput(t *testing.T) Input {
	return fixtureRewriteCA(t, fixtureInput(t), "subject", hexBytes(t, fixtureGoldens(t)["alternateDN"]["CAName"]))
}

// Catches mathematical invalidity, exact Name differences and lower remapping.
func TestIssuerEvaluateRejects(t *testing.T) {
	in := fixtureInput(t)
	assertIssuerFailure(t, fixtureRewriteLeaf(t, in, "serial", []byte{2, 2, 0, 0x82}), ErrInvalidSignature)
	assertIssuerFailure(t, fixtureRewriteLeaf(t, in, "signature", fixtureTLV(3, append([]byte{0}, hexBytes(t, publicSignature)...))), ErrInvalidSignature)
	assertIssuerFailure(t, wrongNameInput(t), ErrIssuerRelationMismatch)
	name := hexBytes(t, fixtureGoldens(t)["base"]["CAName"])
	at := bytes.Index(name, []byte{19, byte(len("nodeissuer offline CA"))})
	if at < 0 {
		t.Fatal("PrintableString fixture")
	}
	name[at] = 12
	changed := fixtureRewriteCA(t, in, "subject", name)
	baseCA, e := x509.ParseCertificate(in.CA.DER)
	if e != nil {
		t.Fatal(e)
	}
	otherCA, e := x509.ParseCertificate(changed.CA.DER)
	if e != nil {
		t.Fatal(e)
	}
	if baseCA.Subject.String() != otherCA.Subject.String() {
		t.Fatal("same display control")
	}
	assertIssuerFailure(t, changed, ErrIssuerRelationMismatch)
	assertIssuerFailure(t, fixtureSetKeyID(t, in, "leafAKI", make([]byte, 32)), nodeleafkeys.ErrBindingMismatch)
	assertIssuerFailure(t, fixtureSetKeyID(t, in, "caSKI", make([]byte, 32)), nodeca.ErrBindingMismatch)
	wrong := in
	wrong.Leaf.ExpectedSPKI = in.CA.DeclaredSPKI
	assertIssuerFailure(t, wrong, nodeleaf.ErrBindingMismatch)
	wrong = in
	wrong.CA.DeclaredSPKI = in.Leaf.ExpectedSPKI
	assertIssuerFailure(t, wrong, nodeca.ErrBindingMismatch)
	wrong = in
	wrong.Leaf.DER = append(bytes.Clone(in.Leaf.DER), 0)
	assertIssuerFailure(t, wrong, nodeleaf.ErrInvalidMaterial)
}

// Catches incorrect public stage precedence, without pretending same-sentinel
// key/Name/ID helper controls dynamically reveal their internal comparison order.
func TestIssuerCombinedFailurePriority(t *testing.T) {
	both := fixtureInput(t)
	both.CA.DER = []byte{0x30, 0}
	both.Leaf.DER = []byte{0x30, 0}
	assertIssuerFailure(t, both, nodeca.ErrInvalidMaterial)
	sig := fixtureTLV(3, []byte{0, 0x30, 6, 2, 1, 0, 2, 1, 1})
	badLower := fixtureRewriteLeaf(t, fixtureInput(t), "signature", sig)
	badLower = fixtureSetKeyID(t, badLower, "leafAKI", make([]byte, 32))
	assertIssuerFailure(t, badLower, nodeleafkeys.ErrBindingMismatch)
	malformed := fixtureRewriteLeaf(t, wrongNameInput(t), "signature", fixtureTLV(3, []byte{0, 0x30, 6, 2, 1, 1, 2, 1, 1, 0}))
	assertIssuerFailure(t, malformed, ErrInvalidMaterial)
	outOfRange := fixtureRewriteLeaf(t, wrongNameInput(t), "signature", sig)
	assertIssuerFailure(t, outOfRange, ErrUnsupportedSignatureProfile)
	badMath := fixtureRewriteLeaf(t, wrongNameInput(t), "signature", fixtureTLV(3, append([]byte{0}, hexBytes(t, publicSignature)...)))
	assertIssuerFailure(t, badMath, ErrIssuerRelationMismatch)
}
func TestIssuerCertificateAmbiguity(t *testing.T) {
	a := fixtureInput(t)
	b := a
	b.CA = fixtureVariant(t, "sameCAKeyAndNameOtherCert").CA
	m, e := Evaluate(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	n, e := Evaluate(context.Background(), b)
	if e != nil {
		t.Fatal(e)
	}
	defer n.Clear()
	sa, sb := m.Summary(), n.Summary()
	assertIssuerPartial(t, sa)
	assertIssuerPartial(t, sb)
	if sa.CA.CADERDigest == sb.CA.CADERDigest || sa.Leaf.Leaf.LeafDERDigest != sb.Leaf.Leaf.LeafDERDigest || sa.LeafTBSDERDigest != sb.LeafTBSDERDigest {
		t.Fatal("exact CA ambiguity")
	}
}
func TestIssuerUntrustedFactsRetained(t *testing.T) {
	cases := []Input{}
	for _, name := range []string{"expired", "future", "crossCAWindow", "sameKeyRoles", "sameCAKeyAndNameOtherCert"} {
		cases = append(cases, fixtureVariant(t, name))
	}
	caSignature := fixtureInput(t)
	caSignature.CA.DER[len(caSignature.CA.DER)-1] ^= 1
	cases = append(cases, caSignature)
	cases = append(cases, fixtureSetKeyID(t, fixtureInput(t), "caAKI", bytes.Repeat([]byte{0x99}, 32)))
	for _, in := range cases {
		m, e := Evaluate(context.Background(), in)
		if e != nil {
			t.Fatal("untrusted fact was incorrectly promoted to policy", e)
		}
		assertIssuerPartial(t, m.Summary())
		m.Clear()
	}
}
func TestIssuerIndependentLifecycle(t *testing.T) {
	in := fixtureInput(t)
	original := [][]byte{bytes.Clone(in.Leaf.DER), bytes.Clone(in.Leaf.ExpectedSPKI), bytes.Clone(in.CA.DER), bytes.Clone(in.CA.DeclaredSPKI)}
	m, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	n, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer n.Clear()
	preserved := n.Summary()
	copyM := *m
	detached := m.Summary()
	m.Clear()
	copyM.Clear()
	if n.Summary() != preserved || m.Summary() != (Summary{}) || copyM.Summary() != (Summary{}) {
		t.Fatal("owner isolation")
	}
	assertIssuerPartial(t, detached)
	for i, b := range [][]byte{in.Leaf.DER, in.Leaf.ExpectedSPKI, in.CA.DER, in.CA.DeclaredSPKI} {
		if !bytes.Equal(b, original[i]) {
			t.Fatal("caller modified", i)
		}
	}
	in.Leaf.DER[0] ^= 1
	if n.Summary() != preserved {
		t.Fatal("borrowed input")
	}
}
func TestIssuerCancellation(t *testing.T) {
	in := fixtureInput(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	for _, c := range []struct {
		ctx  context.Context
		want error
	}{{nil, ErrInvalidInput}, {ctx, context.Canceled}, {expired, context.DeadlineExceeded}} {
		m, e := Evaluate(c.ctx, in)
		if m != nil || e != c.want {
			t.Fatal("pre-cancel", e)
		}
	}
	observed, unobserved := 0, 0
	for i := 0; i < 32; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		timer := time.AfterFunc(50*time.Microsecond, cancel)
		m, e := Evaluate(ctx, in)
		if e == context.Canceled {
			observed++
			if m != nil {
				t.Fatal("canceled partial result")
			}
		} else if e == nil {
			unobserved++
			if m == nil {
				t.Fatal("nil success")
			}
			assertIssuerPartial(t, m.Summary())
			m.Clear()
		} else {
			t.Fatal(e)
		}
		timer.Stop()
		cancel()
	}
	t.Logf("cancellation observed %d; completed before observation %d (not a hard-real-time assertion)", observed, unobserved)
}
