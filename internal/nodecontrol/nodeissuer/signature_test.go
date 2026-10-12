package nodeissuer

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"math/big"
	"testing"
)

const publicSPKI = "3059301306072a8648ce3d020106082a8648ce3d0301070342000460fed4ba255a9d31c961eb74c6356d68c049b8923b61fa6ce669622e60f29fb67903fe1008b8bc99a41ae9e95628bc64f2f1b20c2d7e9f5177a3c294d4462299"
const publicSignature = "3046022100efd48b2aacb6a8fd1140dd9cd45e81d69d2c877b56aaf991c34d0ea84eaf3716022100f7cb1c942d657c41d436c7a1b6e29f65f3e900dbb9aff4064dc4ab2f843acda8"
const publicHash = "af2bdbe1aa9b6ec1e2ade1d694f41fc71a831d0268e9891562113d8a62add1bf"
const curveOrder = "ffffffff00000000ffffffffffffffffbce6faada7179e84f3b9cac2fc632551"

func profileUnit(t *testing.T, in Input) (*x509.Certificate, signedLeaf) {
	t.Helper()
	cert, e := x509.ParseCertificate(fixtureInput(t).Leaf.DER)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := parseLeafEnvelope(in.Leaf.DER)
	if e != nil {
		t.Fatal(e)
	}
	cert.Raw = in.Leaf.DER
	cert.RawTBSCertificate = raw.tbs.full
	cert.RawIssuer = raw.issuer.full
	cert.RawSubject = raw.subject.full
	cert.RawSubjectPublicKeyInfo = raw.spki.full
	if len(raw.signature.content) > 0 {
		cert.Signature = raw.signature.content[1:]
	} else {
		cert.Signature = nil
	}
	return cert, raw
}

// Catches classifying malformed parameters as merely unsupported, accepting
// opaque algorithms, or trusting a parser enum/raw field without correspondence.
func TestIssuerSignatureProfile(t *testing.T) {
	in := fixtureInput(t)
	cert, raw := profileUnit(t, in)
	if e := checkLeafSignatureProfile(cert, raw); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*x509.Certificate){func(c *x509.Certificate) { c.SignatureAlgorithm = x509.UnknownSignatureAlgorithm }, func(c *x509.Certificate) { c.RawTBSCertificate = []byte{1} }, func(c *x509.Certificate) { c.RawIssuer = []byte{1} }, func(c *x509.Certificate) { c.RawSubject = []byte{1} }, func(c *x509.Certificate) { c.RawSubjectPublicKeyInfo = []byte{1} }, func(c *x509.Certificate) { c.Signature = []byte{1} }} {
		bad := *cert
		change(&bad)
		if checkLeafSignatureProfile(&bad, raw) != ErrInvalidMaterial {
			t.Fatal("parser mismatch")
		}
	}
	if checkLeafSignatureProfile(nil, raw) != ErrInvalidMaterial {
		t.Fatal("nil")
	}
	oid := hexBytes(t, "06082a8648ce3d040302")
	cases := []struct {
		alg  []byte
		want error
	}{
		{fixtureTLV(0x30, append(bytes.Clone(oid), 5, 0)), ErrUnsupportedSignatureProfile},
		{fixtureTLV(0x30, append(bytes.Clone(oid), 5, 1, 0)), ErrInvalidMaterial},
		{fixtureTLV(0x30, append(bytes.Clone(oid), 5, 0, 5, 0)), ErrInvalidMaterial},
		{fixtureTLV(0x30, []byte{5, 0}), ErrInvalidMaterial},
		{fixtureTLV(0x30, []byte{6, 1, 0x80}), ErrInvalidMaterial},
		{fixtureTLV(0x30, []byte{6, 2, 0x80, 0}), ErrInvalidMaterial},
		{fixtureTLV(0x30, append(bytes.Clone(oid), 4, 0x81, 0)), ErrInvalidMaterial},
		{fixtureTLV(0x30, append(bytes.Clone(oid), 4, 2, 0)), ErrInvalidMaterial},
		{fixtureTLV(0x30, append(bytes.Clone(oid), 0x30, 1, 0xff)), ErrUnsupportedSignatureProfile},
	}
	for _, id := range []string{"06072a8648ce3d0401", "06082a8648ce3d040301", "06082a8648ce3d040303", "06082a8648ce3d040304", "06092a864886f70d01010b", "06032b6570", "0609608648016503040302", "06012a"} {
		cases = append(cases, struct {
			alg  []byte
			want error
		}{fixtureTLV(0x30, hexBytes(t, id)), ErrUnsupportedSignatureProfile})
	}
	giant := append(bytes.Repeat([]byte{0x81}, 100), 1)
	cases = append(cases, struct {
		alg  []byte
		want error
	}{fixtureTLV(0x30, fixtureTLV(6, giant)), ErrUnsupportedSignatureProfile})
	if !minimalOID(giant) || minimalOID(nil) || minimalOID([]byte{0x80, 0}) || minimalOID([]byte{0x81}) {
		t.Fatal("OID grammar")
	}
	for i, c := range cases {
		modified := fixtureRewriteLeaf(t, in, "innerAlgorithm", c.alg)
		modified = fixtureRewriteLeaf(t, modified, "outerAlgorithm", c.alg)
		cert, raw := profileUnit(t, modified)
		if got := checkLeafSignatureProfile(cert, raw); got != c.want {
			t.Errorf("algorithm %d: %v want %v", i, got, c.want)
		}
	}
	different := fixtureRewriteLeaf(t, in, "outerAlgorithm", fixtureTLV(0x30, hexBytes(t, "06012a")))
	c, r := profileUnit(t, different)
	if checkLeafSignatureProfile(c, r) != ErrInvalidMaterial {
		t.Fatal("inner/outer")
	}
	for _, sig := range [][]byte{fixtureTLV(3, nil), fixtureTLV(3, []byte{0}), fixtureTLV(3, append([]byte{1}, hexBytes(t, publicSignature)...))} {
		modified := fixtureRewriteLeaf(t, in, "signature", sig)
		c, r := profileUnit(t, modified)
		if checkLeafSignatureProfile(c, r) != ErrInvalidMaterial {
			t.Fatal("unused/empty")
		}
	}
}

// Catches malformed/nonminimal DER hiding behind a value-range rejection.
func TestIssuerSignatureDER(t *testing.T) {
	n := hexBytes(t, curveOrder)
	nm := bytes.Clone(n)
	nm[len(nm)-1]--
	np := bytes.Clone(n)
	np[len(np)-1]++
	integer := func(b []byte) []byte { return fixtureTLV(2, b) }
	signature := func(r, s []byte) []byte { return fixtureJoin([][]byte{integer(r), integer(s)}) }
	for _, sig := range [][]byte{signature([]byte{1}, []byte{1}), signature(append([]byte{0}, nm...), append([]byte{0}, nm...)), hexBytes(t, publicSignature)} {
		if e := checkSignatureDER(sig); e != nil {
			t.Fatal("valid representation", e)
		}
	}
	for _, value := range [][]byte{{0}, {0xff}, {0x80}, append([]byte{0}, n...), append([]byte{0}, np...), append([]byte{1}, make([]byte, 33)...)} {
		for _, sig := range [][]byte{signature(value, []byte{1}), signature([]byte{1}, value)} {
			if e := checkSignatureDER(sig); e != ErrUnsupportedSignatureProfile {
				t.Fatal("canonical range", e)
			}
		}
	}
	for _, sig := range [][]byte{nil, bytes.Repeat([]byte{1}, 64), {0x30, 0}, signature([]byte{0, 1}, []byte{1}), signature([]byte{0xff, 0x80}, []byte{1}), signature([]byte{1}, []byte{0, 1}), fixtureJoin([][]byte{integer([]byte{1}), integer([]byte{1}), integer([]byte{1})}), append(signature([]byte{1}, []byte{1}), 0), fixtureJoin([][]byte{integer([]byte{0}), integer([]byte{1}), []byte{5, 0}}), fixtureJoin([][]byte{integer(append([]byte{0}, n...)), integer([]byte{1}), []byte{5, 0}}), {0x30, 6, 2, 1, 1, 2, 2, 1}, {0x30, 6, 4, 1, 1, 2, 1, 1}} {
		if e := checkSignatureDER(sig); e != ErrInvalidMaterial {
			t.Fatal("malformed signature", sig, e)
		}
	}
}

// RFC6979 facts supply the validity oracle; this is not a certificate result.
func TestIssuerSignaturePrimitive(t *testing.T) {
	pub, e := x509.ParsePKIXPublicKey(hexBytes(t, publicSPKI))
	if e != nil {
		t.Fatal(e)
	}
	key := pub.(*ecdsa.PublicKey)
	sig := hexBytes(t, publicSignature)
	if !ecdsa.VerifyASN1(key, hexBytes(t, publicHash), sig) {
		t.Fatal("public oracle")
	}
	var rs struct{ R, S *big.Int }
	if _, e = asn1.Unmarshal(sig, &rs); e != nil {
		t.Fatal(e)
	}
	other := new(big.Int).SetBytes(hexBytes(t, "0834e36ad29a83bf2bc9385e491d6099c8fdf9d1ed67aa7ea5f51f93782857a9"))
	otherDER, e := asn1.Marshal(struct{ R, S *big.Int }{rs.R, other})
	if e != nil {
		t.Fatal(e)
	}
	for _, signature := range [][]byte{sig, otherDER} {
		raw := signedLeaf{tbs: derElement{full: []byte("sample")}, signature: derElement{content: append([]byte{0}, signature...)}}
		if e = verifyLeafSignature(&x509.Certificate{PublicKey: key}, raw); e != nil {
			t.Fatal("primitive", e)
		}
	}
	raw := signedLeaf{tbs: derElement{full: []byte("sample")}, signature: derElement{content: append([]byte{0}, sig...)}}
	flip := raw
	flip.signature.content = bytes.Clone(raw.signature.content)
	flip.signature.content[len(flip.signature.content)-1] ^= 1
	if verifyLeafSignature(&x509.Certificate{PublicKey: key}, flip) != ErrInvalidSignature {
		t.Fatal("bit flip")
	}
	wrong := raw
	wrong.tbs.full = []byte("samplex")
	if verifyLeafSignature(&x509.Certificate{PublicKey: key}, wrong) != ErrInvalidSignature {
		t.Fatal("preimage")
	}
	ca, e := x509.ParseCertificate(fixtureInput(t).CA.DER)
	if e != nil {
		t.Fatal(e)
	}
	if verifyLeafSignature(ca, raw) != ErrInvalidSignature {
		t.Fatal("wrong key")
	}
	for _, c := range []*x509.Certificate{nil, {}, {PublicKey: (*ecdsa.PublicKey)(nil)}, {PublicKey: &ecdsa.PublicKey{Curve: elliptic.P256(), X: big.NewInt(0), Y: big.NewInt(0)}}} {
		if verifyLeafSignature(c, raw) != ErrInvalidMaterial {
			t.Fatal("invalid actual key")
		}
	}
}

// Catches hashing whole certificate or TBS content instead of the exact TBS DER.
func TestIssuerSignaturePreimage(t *testing.T) {
	in := fixtureInput(t)
	s, e := evaluateLower(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	c, e := parseSnapshots(context.Background(), in, s)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := parseLeafEnvelope(in.Leaf.DER)
	if e != nil {
		t.Fatal(e)
	}
	if e = checkLeafSignatureProfile(c.leaf, raw); e != nil {
		t.Fatal(e)
	}
	if e = verifyLeafSignature(c.ca, raw); e != nil {
		t.Fatal(e)
	}
	for _, preimage := range [][]byte{in.Leaf.DER, raw.tbs.content, raw.signature.full, in.CA.DER} {
		wrong := raw
		wrong.tbs.full = preimage
		if verifyLeafSignature(c.ca, wrong) != ErrInvalidSignature {
			t.Fatal("wrong preimage")
		}
	}
}
