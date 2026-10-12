package nodeissuer

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	"testing"
)

func FuzzIssuerEnvelope(f *testing.F) {
	f.Add(fixtureInput(f).Leaf.DER)
	f.Add(hexBytes(f, publicSignature))
	f.Add([]byte{0x30, 0x81, 0})
	f.Add(fixtureRewriteLeaf(f, fixtureInput(f), "signature", fixtureTLV(3, []byte{1, 0x30, 0})).Leaf.DER)
	f.Add([]byte{0x30})
	f.Fuzz(func(t *testing.T, der []byte) {
		if len(der) > 65536 {
			t.Skip()
		}
		raw, e := parseLeafEnvelope(der)
		if e != nil && e != ErrInvalidMaterial {
			t.Fatal("envelope error identity")
		}
		if e == nil {
			if !bytes.Equal(fixtureJoin([][]byte{raw.tbs.full, raw.outerAlgorithm.full, raw.signature.full}), der) {
				t.Fatal("envelope range")
			}
		}
		e = checkSignatureDER(der)
		if e != nil && e != ErrInvalidMaterial && e != ErrUnsupportedSignatureProfile {
			t.Fatal("signature error identity")
		}
	})
}
func FuzzIssuerPair(f *testing.F) {
	base := fixtureInput(f)
	for _, in := range []Input{base, fixtureRewriteLeaf(f, fixtureInput(f), "signature", fixtureTLV(3, append([]byte{0}, hexBytes(f, publicSignature)...))), fixtureRewriteCA(f, fixtureInput(f), "subject", hexBytes(f, fixtureGoldens(f)["alternateDN"]["CAName"])), fixtureSetKeyID(f, fixtureInput(f), "leafAKI", make([]byte, 32))} {
		f.Add(in.Leaf.DER, in.CA.DER, in.Leaf.ExpectedSPKI, in.CA.DeclaredSPKI)
	}
	f.Fuzz(func(t *testing.T, leafDER, caDER, expectedSPKI, declaredSPKI []byte) {
		if len(leafDER) > 65536 || len(caDER) > 65536 || len(expectedSPKI) > 512 || len(declaredSPKI) > 8192 {
			t.Skip()
		}
		leaf := base.Leaf
		leaf.DER = bytes.Clone(leafDER)
		leaf.ExpectedSPKI = bytes.Clone(expectedSPKI)
		in := Input{Leaf: leaf, CA: nodeca.Input{DER: bytes.Clone(caDER), DeclaredSPKI: bytes.Clone(declaredSPKI)}}
		m, e := Evaluate(context.Background(), in)
		if e != nil {
			if m != nil {
				m.Clear()
				t.Fatal("failed result")
			}
			known := false
			for _, allowed := range []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidMaterial, ErrUnsupportedSignatureProfile, ErrIssuerRelationMismatch, ErrInvalidSignature, nodeleaf.ErrInvalidInput, nodeleaf.ErrLimitExceeded, nodeleaf.ErrInvalidMaterial, nodeleaf.ErrUnsupportedLeafProfile, nodeleaf.ErrBindingMismatch, nodeleafkeys.ErrInvalidInput, nodeleafkeys.ErrLimitExceeded, nodeleafkeys.ErrInvalidMaterial, nodeleafkeys.ErrUnsupportedKeyIdentifierProfile, nodeleafkeys.ErrBindingMismatch, nodeca.ErrInvalidInput, nodeca.ErrLimitExceeded, nodeca.ErrInvalidMaterial, nodeca.ErrUnsupportedCAProfile, nodeca.ErrBindingMismatch} {
				if e == allowed {
					known = true
				}
			}
			if !known {
				t.Fatal("unexpected error identity", e)
			}
			return
		}
		if m == nil {
			t.Fatal("nil success")
		}
		assertIssuerPartial(t, m.Summary())
		m.Clear()
		if m.Summary() != (Summary{}) {
			t.Fatal("uncleared")
		}
		if !bytes.Equal(in.Leaf.DER, leafDER) || !bytes.Equal(in.CA.DER, caDER) || !bytes.Equal(in.Leaf.ExpectedSPKI, expectedSPKI) || !bytes.Equal(in.CA.DeclaredSPKI, declaredSPKI) {
			t.Fatal("caller bytes changed")
		}
	})
}
