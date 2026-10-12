package nodeleafkeys

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"math/big"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
	"time"
)

func keysCheckPartial(t testing.TB, s Summary) {
	t.Helper()
	want := [6]nodeleaf.Obligation{nodeleaf.LeafSignatureAndIssuerTrust, nodeleaf.KeyIdentifierDerivationAndIssuerBinding, nodeleaf.TrustedCurrentTime, nodeleaf.IdentityAndAuthorityProvenance, nodeleaf.PrivateKeyPossession, nodeleaf.MainStateAndRuntimeBinding}
	if s.Proof != NodeLeafDeclaredKeyIdentifiersBound || s.Leaf.Proof != nodeleaf.NodeLeafMaterialShapeValid || s.Leaf.Pending != want {
		t.Fatal("partial proof/pending")
	}
}
func keysEvalError(t testing.TB, ctx context.Context, in Input, want error) {
	t.Helper()
	m, err := Evaluate(ctx, in)
	if m != nil {
		m.Clear()
		t.Fatal("partial material escaped")
	}
	if err != want {
		t.Fatalf("got %v want %v", err, want)
	}
}
func TestKeysEvaluate(t *testing.T) {
	in := keysFixture(t)
	m, err := Evaluate(context.Background(), in)
	if err != nil || m == nil {
		t.Fatal(err)
	}
	defer m.Clear()
	s := m.Summary()
	keysCheckPartial(t, s)
	ski, aki, der := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI), sha256.Sum256(in.Leaf.DER)
	if s.SubjectKeyIdentifier != KeyIdentifier(ski) || s.AuthorityKeyIdentifier != KeyIdentifier(aki) || s.DeclaredIssuerSPKIDERDigest != contracts.Digest(aki) || s.Leaf.SPKIDERDigest != contracts.Digest(ski) || s.Leaf.LeafDERDigest != contracts.Digest(der) {
		t.Fatal("exact preimages")
	}
	if s.Leaf.NodeID != in.Leaf.NodeID || s.Leaf.TrustDomain != in.Leaf.TrustDomain {
		t.Fatal("scope")
	}
	s.SubjectKeyIdentifier[0] ^= 1
	s.Leaf.Pending[0] = 0
	if m.Summary().SubjectKeyIdentifier == s.SubjectKeyIdentifier || m.Summary().Leaf.Pending[0] == 0 {
		t.Fatal("detached")
	}
}
func TestKeysEvaluateIssuerKinds(t *testing.T) {
	for _, kind := range []string{"p256", "p384", "p521", "rsa", "ed25519"} {
		in := keysFixture(t)
		in.DeclaredIssuerSPKI = keysIssuerSPKI(t, kind)
		ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
		in = keysWithIDs(t, in, ski[:], aki[:])
		m, err := Evaluate(context.Background(), in)
		if err != nil || m == nil {
			t.Fatalf("%s: %v", kind, err)
		}
		keysCheckPartial(t, m.Summary())
		if m.Summary().AuthorityKeyIdentifier != KeyIdentifier(aki) {
			t.Fatal("issuer preimage")
		}
		m.Clear()
	}
}
func TestKeysEvaluateRejections(t *testing.T) {
	good := keysFixture(t)
	for name, raw := range keysSPKIVariants(t) {
		t.Run(name, func(t *testing.T) {
			in := good
			in.DeclaredIssuerSPKI = raw
			keysEvalError(t, context.Background(), in, ErrInvalidInput)
		})
	}
	for name, raw := range keysUnsupportedSPKIs(t) {
		t.Run(name, func(t *testing.T) {
			in := good
			in.DeclaredIssuerSPKI = raw
			keysEvalError(t, context.Background(), in, ErrInvalidInput)
		})
	}
	ski, aki := sha256.Sum256(good.Leaf.ExpectedSPKI), sha256.Sum256(good.DeclaredIssuerSPKI)
	for _, role := range []string{"ski", "aki"} {
		for _, n := range []int{1, 20, 31, 33, 64} {
			a, b := ski[:], aki[:]
			if role == "ski" {
				a = make([]byte, n)
			} else {
				b = make([]byte, n)
			}
			keysEvalError(t, context.Background(), keysWithIDs(t, good, a, b), ErrUnsupportedKeyIdentifierProfile)
		}
	}
	wrongSKI, wrongAKI := ski, aki
	wrongSKI[0] ^= 1
	wrongAKI[0] ^= 1
	for _, in := range []Input{keysWithIDs(t, good, wrongSKI[:], aki[:]), keysWithIDs(t, good, ski[:], wrongAKI[:])} {
		keysEvalError(t, context.Background(), in, ErrBindingMismatch)
	}
	for _, tc := range []struct {
		edit func(*Input)
		want error
	}{
		{func(in *Input) { in.Leaf.DER = []byte{1} }, ErrInvalidMaterial},
		{func(in *Input) { in.Leaf.DER = append(bytes.Clone(in.Leaf.DER), 0) }, ErrInvalidMaterial},
		{func(in *Input) { in.Leaf.ExpectedSPKI = []byte{1} }, ErrInvalidInput},
		{func(in *Input) { in.Leaf.ExpectedSPKI = keysIssuerSPKI(t, "p256") }, ErrBindingMismatch},
		{func(in *Input) { in.Leaf.TrustDomain = "else.test" }, ErrBindingMismatch},
		{func(in *Input) {
			parts := keysParts(t, in.Leaf.DER)
			fields := keysParts(t, parts[0])
			fields[1] = []byte{2, 1, 0}
			parts[0] = keysJoin(fields)
			in.Leaf.DER = keysJoin(parts)
		}, ErrUnsupportedLeafProfile},
	} {
		in := good
		tc.edit(&in)
		keysEvalError(t, context.Background(), in, tc.want)
	}
	// Old arbitrary 20-byte identifiers remain lower-shape-valid, not this profile.
	old := keysWithIDs(t, good, bytes.Repeat([]byte{1}, 20), bytes.Repeat([]byte{2}, 20))
	keysLower(t, old)
	keysEvalError(t, context.Background(), old, ErrUnsupportedKeyIdentifierProfile)
}
func TestKeysPartialProofOnly(t *testing.T) {
	good := keysFixture(t)
	badSig := keysWithSignatureBitFlipped(t, good)
	cert := keysCert(t, badSig)
	var sig struct{ R, S *big.Int }
	tail, err := asn1.Unmarshal(cert.Signature, &sig)
	if err != nil || len(tail) != 0 || sig.R == nil || sig.S == nil || sig.R.Sign() <= 0 || sig.S.Sign() <= 0 {
		t.Fatal("signature syntax control")
	}
	issuer := &x509.Certificate{PublicKey: &keysTestEC(elliptic.P256(), 1).PublicKey, PublicKeyAlgorithm: x509.ECDSA, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	if cert.CheckSignatureFrom(issuer) == nil {
		t.Fatal("bad cryptographic signature control")
	}
	same := good
	same.DeclaredIssuerSPKI = bytes.Clone(good.Leaf.ExpectedSPKI)
	h := sha256.Sum256(same.Leaf.ExpectedSPKI)
	same = keysWithIDs(t, same, h[:], h[:])
	untrusted := good
	untrusted.DeclaredIssuerSPKI = keysIssuerSPKI(t, "ed25519")
	a, b := sha256.Sum256(good.Leaf.ExpectedSPKI), sha256.Sum256(untrusted.DeclaredIssuerSPKI)
	untrusted = keysWithIDs(t, untrusted, a[:], b[:])
	for _, in := range []Input{badSig, keysFixtureAt(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)), keysFixtureAt(t, time.Date(2090, 1, 1, 0, 0, 0, 0, time.UTC)), untrusted, same} {
		m, e := Evaluate(context.Background(), in)
		if e != nil || m == nil {
			t.Fatal("relation only", e)
		}
		keysCheckPartial(t, m.Summary())
		m.Clear()
	}
}
func TestKeysErrorPrecedence(t *testing.T) {
	in := keysFixture(t)
	ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
	ski[0] ^= 1
	bad := in
	bad.Leaf.DER = []byte{1}
	bad.DeclaredIssuerSPKI = []byte{1}
	keysEvalError(t, context.Background(), bad, ErrInvalidMaterial)
	bad = in
	bad.Leaf.TrustDomain = "other.test"
	bad.DeclaredIssuerSPKI = []byte{1}
	keysEvalError(t, context.Background(), bad, ErrBindingMismatch)
	bad = keysWithIDs(t, in, ski[:], aki[:20])
	bad.DeclaredIssuerSPKI = []byte{1}
	keysEvalError(t, context.Background(), bad, ErrInvalidInput)
	bad.DeclaredIssuerSPKI = in.DeclaredIssuerSPKI
	keysEvalError(t, context.Background(), bad, ErrUnsupportedKeyIdentifierProfile)
	keysEvalError(t, nil, Input{DeclaredIssuerSPKI: make([]byte, 8193)}, ErrInvalidInput)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	keysEvalError(t, ctx, Input{DeclaredIssuerSPKI: make([]byte, 8193)}, context.Canceled)
	keysEvalError(t, context.Background(), Input{DeclaredIssuerSPKI: make([]byte, 8193)}, ErrLimitExceeded)
}

type keysTraceContext struct {
	context.Context
	calls, trigger int
	failure        error
	panicAt        bool
}

func (c *keysTraceContext) Err() error {
	c.calls++
	if c.trigger > 0 && c.calls >= c.trigger {
		if c.panicAt {
			panic("keys context checkpoint")
		}
		return c.failure
	}
	return nil
}
func keysAssertPanic(t testing.TB, ctx *keysTraceContext, in Input) {
	t.Helper()
	panicked := false
	func() {
		defer func() {
			if v := recover(); v != nil {
				if v != "keys context checkpoint" {
					t.Fatalf("unexpected panic: %v", v)
				}
				panicked = true
			}
		}()
		m, _ := Evaluate(ctx, in)
		if m != nil {
			m.Clear()
			t.Error("material escaped panic")
		}
	}()
	if !panicked {
		t.Fatal("panic swallowed")
	}
}
func TestKeysCancellation(t *testing.T) {
	in := keysFixture(t)
	trace := &keysTraceContext{Context: context.Background()}
	m, err := Evaluate(trace, in)
	if err != nil || m == nil || trace.calls == 0 {
		t.Fatal("trace", err)
	}
	m.Clear()
	for point := 1; point <= trace.calls; point++ {
		for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
			ctx := &keysTraceContext{Context: context.Background(), trigger: point, failure: failure}
			keysEvalError(t, ctx, in, failure)
		}
		keysAssertPanic(t, &keysTraceContext{Context: context.Background(), trigger: point, panicAt: true}, in)
	}
}
func TestKeysLateCancellation(t *testing.T) {
	good := keysFixture(t)
	ski, aki := sha256.Sum256(good.Leaf.ExpectedSPKI), sha256.Sum256(good.DeclaredIssuerSPKI)
	badLower := good
	badLower.Leaf.DER = []byte{1}
	badIssuer := good
	badIssuer.DeclaredIssuerSPKI = []byte{1}
	short := keysWithIDs(t, good, ski[:], aki[:20])
	ski[0] ^= 1
	wrong := keysWithIDs(t, good, ski[:], aki[:])
	for _, tc := range []struct {
		in   Input
		want error
	}{{good, nil}, {badLower, ErrInvalidMaterial}, {badIssuer, ErrInvalidInput}, {short, ErrUnsupportedKeyIdentifierProfile}, {wrong, ErrBindingMismatch}} {
		trace := &keysTraceContext{Context: context.Background()}
		m, err := Evaluate(trace, tc.in)
		if err != tc.want {
			t.Fatal("ordinary error control", err)
		}
		if m != nil {
			m.Clear()
		}
		if trace.calls == 0 {
			t.Fatal("no checkpoints")
		}
		for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
			keysEvalError(t, &keysTraceContext{Context: context.Background(), trigger: trace.calls, failure: failure}, tc.in, failure)
		}
		keysAssertPanic(t, &keysTraceContext{Context: context.Background(), trigger: trace.calls, panicAt: true}, tc.in)
	}
}
func TestKeysEvaluateOwnership(t *testing.T) {
	in := keysFixture(t)
	n := len(in.Leaf.DER)
	backing := append(bytes.Clone(in.Leaf.DER), in.DeclaredIssuerSPKI...)
	offset := bytes.Index(backing[:n], in.Leaf.ExpectedSPKI)
	if offset < 0 {
		t.Fatal("fixture SPKI subslice")
	}
	in.Leaf.DER = backing[:n]
	in.Leaf.ExpectedSPKI = backing[offset : offset+len(in.Leaf.ExpectedSPKI)]
	in.DeclaredIssuerSPKI = backing[n:]
	caller := bytes.Clone(backing)
	m, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	detached := m.Summary()
	copied := *m
	owned := m.owned
	buffers := [][]byte{owned.der, owned.expectedSPKI, owned.declaredIssuerSPKI}
	originals := [][]byte{bytes.Clone(buffers[0]), bytes.Clone(buffers[1]), bytes.Clone(buffers[2])}
	clear(backing)
	if m.Summary() != detached {
		t.Fatal("caller mutation summary")
	}
	for i, b := range buffers {
		if !bytes.Equal(b, originals[i]) {
			t.Fatal("caller mutation owner")
		}
	}
	copy(backing, caller)
	m.Clear()
	copied.Clear()
	if !bytes.Equal(backing, caller) || copied.Summary() != (Summary{}) || owned.summary != (Summary{}) {
		t.Fatal("Clear alias/caller")
	}
	for _, b := range buffers {
		if !bytes.Equal(b, make([]byte, len(b))) {
			t.Fatal("reachable owned buffer not cleared")
		}
	}
	keysCheckPartial(t, detached)
}
