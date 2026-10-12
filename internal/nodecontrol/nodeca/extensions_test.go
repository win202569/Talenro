package nodeca

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"testing"
)

func extensionPair(t testing.TB) (*x509.Certificate, []rawExtension) {
	t.Helper()
	in := fixtureInput(false)
	c := fixtureCert(t, in)
	r, e := parseRawCA(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	return c, r.extensions
}
func extensionChange(t testing.TB, last int, value []byte) (*x509.Certificate, []rawExtension) {
	t.Helper()
	c, r := extensionPair(t)
	for i := range r {
		if r[i].id.Equal(asn1.ObjectIdentifier{2, 5, 29, last}) {
			r[i].value = value
		}
	}
	c.Extensions = nil
	for _, v := range r {
		c.Extensions = append(c.Extensions, pkix.Extension{Id: v.id, Critical: v.critical, Value: v.value})
	}
	return c, r
}

func TestCAExtensions(t *testing.T) {
	c, r := extensionPair(t)
	got, e := checkCAExtensions(c, r)
	if e != nil || got.HasPathLen || got.PathLen != 0 || got.CRLSignAllowed || !bytes.Equal(got.SubjectKeyIdentifier[:], fixtureHex(subjectSPKIHash)) || got.SubjectKeyIdentifier != got.AuthorityKeyIdentifier {
		t.Fatal("projection", got, e)
	}
	for i := range r {
		for _, kind := range []string{"missing", "duplicate", "criticality"} {
			cp, rr := extensionPair(t)
			switch kind {
			case "missing":
				rr = append(rr[:i], rr[i+1:]...)
			case "duplicate":
				rr = append(rr, rr[i])
			case "criticality":
				rr[i].critical = !rr[i].critical
			}
			if _, e := checkCAExtensions(cp, rr); !errors.Is(e, ErrUnsupportedCAProfile) {
				t.Fatalf("%d/%s %v", i, kind, e)
			}
		}
	}
	for _, id := range []asn1.ObjectIdentifier{{2, 5, 29, 17}, {2, 5, 29, 37}, {2, 5, 29, 30}, {2, 5, 29, 32}, {1, 3, 6, 1, 5, 5, 7, 1, 1}, {2, 5, 29, 31}, {1, 2, 3, 4}} {
		for _, critical := range []bool{false, true} {
			cp, rr := extensionPair(t)
			rr = append(rr, rawExtension{id: id, critical: critical, value: []byte{5, 0}})
			if _, e := checkCAExtensions(cp, rr); !errors.Is(e, ErrUnsupportedCAProfile) {
				t.Fatal("extra extension", e)
			}
		}
	}
	// All 24 permutations must preserve the same fixed projection.
	var visit func([]rawExtension, int)
	visit = func(v []rawExtension, n int) {
		if n == len(v) {
			cp := *c
			cp.Extensions = nil
			for _, x := range v {
				cp.Extensions = append(cp.Extensions, pkix.Extension{Id: x.id, Critical: x.critical, Value: x.value})
			}
			g, e := checkCAExtensions(&cp, v)
			if e != nil || g != got {
				t.Fatal("permutation", e)
			}
			return
		}
		for i := n; i < len(v); i++ {
			v[n], v[i] = v[i], v[n]
			visit(v, n+1)
			v[n], v[i] = v[i], v[n]
		}
	}
	visit(append([]rawExtension(nil), r...), 0)
	mismatch := *c
	mismatch.SubjectKeyId = []byte{1}
	if _, e = checkCAExtensions(&mismatch, r); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("parsed SKI mismatch", e)
	}
	mismatch = *c
	mismatch.KeyUsage = x509.KeyUsageDigitalSignature
	if _, e = checkCAExtensions(&mismatch, r); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("parsed KU mismatch", e)
	}
}
func TestCAPathLenPresence(t *testing.T) {
	for _, v := range []struct {
		name string
		b    []byte
		has  bool
		path uint8
		want error
	}{
		{"absent", fixtureHex("30030101ff"), false, 0, nil}, {"zero", fixtureHex("30060101ff020100"), true, 0, nil}, {"one", fixtureHex("30060101ff020101"), true, 1, nil}, {"255", fixtureHex("30070101ff020200ff"), true, 255, nil},
		{"negative", fixtureHex("30060101ff0201ff"), false, 0, ErrUnsupportedCAProfile}, {"256", fixtureHex("30070101ff02020100"), false, 0, ErrUnsupportedCAProfile}, {"nonminimal", fixtureHex("30070101ff02020001"), false, 0, ErrInvalidMaterial},
		{"CAfalse", fixtureHex("3000"), false, 0, ErrUnsupportedCAProfile}, {"explicitfalse", fixtureHex("3003010100"), false, 0, ErrInvalidMaterial},
	} {
		t.Run(v.name, func(t *testing.T) {
			c, r := extensionChange(t, 19, v.b)
			if v.want == nil {
				c.BasicConstraintsValid = true
				c.IsCA = true
				c.MaxPathLen = -1
				c.MaxPathLenZero = false
				if v.has {
					c.MaxPathLen = int(v.path)
					c.MaxPathLenZero = v.path == 0
				}
			}
			got, e := checkCAExtensions(c, r)
			if !errors.Is(e, v.want) || (e == nil && (got.HasPathLen != v.has || got.PathLen != v.path)) {
				t.Fatalf("got %+v %v", got, e)
			}
		})
	}
}
func TestCAExtensionCanonicality(t *testing.T) {
	for _, v := range []struct {
		last int
		b    []byte
		want error
	}{
		{15, fixtureHex("03020204"), nil}, {15, fixtureHex("03020106"), nil}, {15, fixtureHex("03020780"), ErrUnsupportedCAProfile}, {15, fixtureHex("03020186"), ErrUnsupportedCAProfile},
		{15, fixtureHex("0303020400"), ErrInvalidMaterial}, {15, fixtureHex("030202040000"), ErrInvalidMaterial}, {15, fixtureHex("03020804"), ErrInvalidMaterial},
		{19, fixtureHex("30050101ff0500"), ErrInvalidMaterial}, {19, fixtureHex("300301017f"), ErrInvalidMaterial},
		{14, append(fixtureTLV(4, make([]byte, 32)), 0), ErrInvalidMaterial}, {35, append(fixtureTLV(0x30, fixtureTLV(0x80, make([]byte, 32))), 0), ErrInvalidMaterial},
		{35, fixtureTLV(0x30, fixtureTLV(0xa0, make([]byte, 32))), ErrInvalidMaterial},
		{35, fixtureTLV(0x30, fixtureTLV(0xa0, fixtureTLV(4, make([]byte, 32)))), ErrUnsupportedCAProfile},
		{35, fixtureTLV(0x30, append(fixtureTLV(0x80, make([]byte, 32)), []byte{0x82, 1, 1}...)), ErrUnsupportedCAProfile},
		{35, fixtureTLV(0x30, fixtureTLV(0xa1, []byte{})), ErrUnsupportedCAProfile},
	} {
		c, r := extensionChange(t, v.last, v.b)
		if v.want == nil && v.last == 15 {
			c.KeyUsage = x509.KeyUsageCertSign
			if bytes.Equal(v.b, fixtureHex("03020106")) {
				c.KeyUsage |= x509.KeyUsageCRLSign
			}
		}
		got, e := checkCAExtensions(c, r)
		if !errors.Is(e, v.want) {
			t.Fatalf("%d/%x got %v want %v", v.last, v.b, e, v.want)
		}
		if e == nil && got.CRLSignAllowed != (c.KeyUsage&x509.KeyUsageCRLSign != 0) {
			t.Fatal("CRL projection")
		}
	}
	for _, last := range []int{14, 35} {
		for _, n := range []int{1, 20, 31, 32, 33, 64} {
			b := fixtureTLV(4, make([]byte, n))
			if last == 35 {
				b = fixtureTLV(0x30, fixtureTLV(0x80, make([]byte, n)))
			}
			c, r := extensionChange(t, last, b)
			want := error(ErrUnsupportedCAProfile)
			if n == 32 {
				want = nil
				if last == 14 {
					c.SubjectKeyId = make([]byte, 32)
				} else {
					c.AuthorityKeyId = make([]byte, 32)
				}
			}
			if _, e := checkCAExtensions(c, r); !errors.Is(e, want) {
				t.Fatalf("ID %d/%d got %v want %v", last, n, e, want)
			}
		}
	}
}
