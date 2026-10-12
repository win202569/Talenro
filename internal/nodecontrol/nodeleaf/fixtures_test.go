package nodeleaf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"github.com/google/uuid"
	"math/big"
	"testing"
	"time"
)

func fixtureKey(n int64) *ecdsa.PrivateKey {
	d := big.NewInt(n)
	x, y := elliptic.P256().ScalarBaseMult(d.Bytes())
	return &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, D: d}
}
func fixtureTLV(tag byte, content []byte) []byte {
	b, e := asn1.Marshal(asn1.RawValue{Class: int(tag >> 6), Tag: int(tag & 31), IsCompound: tag&32 != 0, Bytes: content})
	if e != nil {
		panic(e)
	}
	return b
}
func fixtureExtensions() []pkix.Extension {
	ids := make([]byte, 20)
	for i := range ids {
		ids[i] = byte(i + 1)
	}
	aki := make([]byte, 20)
	for i := range aki {
		aki[i] = byte(i + 21)
	}
	eku, e := asn1.Marshal([]asn1.ObjectIdentifier{{1, 3, 6, 1, 5, 5, 7, 3, 2}})
	if e != nil {
		panic(e)
	}
	uri := []byte("spiffe://node.test/node/11111111-2222-4333-8444-555555555555")
	return []pkix.Extension{
		{Id: asn1.ObjectIdentifier{2, 5, 29, 19}, Critical: true, Value: []byte{0x30, 0}},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 15}, Critical: true, Value: []byte{3, 2, 7, 0x80}},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 37}, Value: eku},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 17}, Critical: true, Value: fixtureTLV(0x30, fixtureTLV(0x86, uri))},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 14}, Value: fixtureTLV(4, ids)},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 35}, Value: fixtureTLV(0x30, fixtureTLV(0x80, aki))},
	}
}
func leafFixture(t testing.TB) Input { return leafWithTemplate(t, nil) }
func leafWithTemplate(t testing.TB, edit func(*x509.Certificate)) Input {
	t.Helper()
	issuer, key := fixtureKey(1), fixtureKey(2)
	parent := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "offline test issuer"}, PublicKey: &issuer.PublicKey, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, SubjectKeyId: []byte{1}}
	start := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	c := &x509.Certificate{SerialNumber: big.NewInt(128), NotBefore: start, NotAfter: start.Add(24 * time.Hour), ExtraExtensions: fixtureExtensions()}
	if edit != nil {
		edit(c)
	}
	der, e := x509.CreateCertificate(rand.Reader, c, parent, &key.PublicKey, issuer)
	if e != nil {
		t.Fatal("fixture signing", e)
	}
	spki, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if e != nil {
		t.Fatal(e)
	}
	return Input{DER: der, ExpectedSPKI: spki, NodeID: uuid.MustParse("11111111-2222-4333-8444-555555555555"), TrustDomain: "node.test"}
}
func fixtureParts(t testing.TB, raw []byte) [][]byte {
	t.Helper()
	var v asn1.RawValue
	tail, e := asn1.Unmarshal(raw, &v)
	if e != nil || len(tail) != 0 {
		t.Fatal("fixture outer", e)
	}
	var parts [][]byte
	for b := v.Bytes; len(b) > 0; {
		var child asn1.RawValue
		rest, e := asn1.Unmarshal(b, &child)
		if e != nil {
			t.Fatal("fixture child", e)
		}
		parts = append(parts, bytes.Clone(child.FullBytes))
		b = rest
	}
	return parts
}
func fixtureJoin(parts [][]byte) []byte { return fixtureTLV(0x30, bytes.Join(parts, nil)) }
func rewriteFixture(t testing.TB, in Input, path string, replacement []byte) Input {
	t.Helper()
	in.ExpectedSPKI = bytes.Clone(in.ExpectedSPKI)
	if path == "certificate" {
		in.DER = bytes.Clone(replacement)
		return in
	}
	cert := fixtureParts(t, in.DER)
	if path == "signature" {
		cert[2] = replacement
		in.DER = fixtureJoin(cert)
		return in
	}
	if path == "tbs" {
		cert[0] = replacement
		in.DER = fixtureJoin(cert)
		return in
	}
	fields := fixtureParts(t, cert[0])
	base := 0
	if fields[0][0] == 0xa0 {
		base = 1
	}
	index := -1
	switch path {
	case "version":
		if base == 0 {
			t.Fatal("fixture version absent")
		}
		index = 0
	case "serial":
		index = base
	case "subject":
		index = base + 4
	case "spki":
		index = base + 5
	case "extensions":
		for i, p := range fields {
			if p[0] == 0xa3 {
				index = i
			}
		}
	case "notBefore", "notAfter":
		times := fixtureParts(t, fields[base+3])
		j := 0
		if path == "notAfter" {
			j = 1
		}
		times[j] = replacement
		fields[base+3] = fixtureJoin(times)
	default:
		t.Fatal("unknown fixture path", path)
	}
	if index >= 0 {
		if replacement == nil {
			fields = append(fields[:index], fields[index+1:]...)
		} else {
			fields[index] = replacement
		}
	} else if path != "notBefore" && path != "notAfter" {
		t.Fatal("missing fixture field", path)
	}
	cert[0] = fixtureJoin(fields)
	in.DER = fixtureJoin(cert)
	return in
}
func TestLeafFixtureControls(t *testing.T) {
	in := leafFixture(t)
	c, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	issuer := &x509.Certificate{PublicKey: &fixtureKey(1).PublicKey, PublicKeyAlgorithm: x509.ECDSA, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	if e := c.CheckSignatureFrom(issuer); e != nil {
		t.Fatal("positive fixture signature", e)
	}
	if c.Version != 3 || !bytes.Equal(c.RawSubject, []byte{0x30, 0}) || c.SerialNumber.Int64() != 128 || !c.NotBefore.Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)) || !c.NotAfter.Equal(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)) || !bytes.Equal(c.RawSubjectPublicKeyInfo, in.ExpectedSPKI) {
		t.Fatal("positive fields")
	}
	if len(c.URIs) != 1 || c.URIs[0].String() != "spiffe://node.test/node/11111111-2222-4333-8444-555555555555" || len(c.Extensions) != 6 {
		t.Fatal("positive inventory")
	}
	for _, want := range fixtureExtensions() {
		found := false
		for _, got := range c.Extensions {
			if got.Id.Equal(want.Id) {
				found = true
				if got.Critical != want.Critical || !bytes.Equal(got.Value, want.Value) {
					t.Fatal("positive raw extension", want.Id)
				}
			}
		}
		if !found {
			t.Fatal("missing fixture extension")
		}
	}
}
func fixtureNonP256SPKI(t testing.TB) [][]byte {
	t.Helper()
	x, y := elliptic.P384().ScalarBaseMult([]byte{1})
	var out [][]byte
	for _, key := range []any{&ecdsa.PublicKey{Curve: elliptic.P384(), X: x, Y: y}, &rsa.PublicKey{N: big.NewInt(65537), E: 65537}} {
		b, e := x509.MarshalPKIXPublicKey(key)
		if e != nil {
			t.Fatal(e)
		}
		out = append(out, b)
	}
	return out
}
