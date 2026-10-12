package nodeleafkeys

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"github.com/google/uuid"
	"math/big"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
	"time"
)

func keysTestEC(curve elliptic.Curve, n int64) *ecdsa.PrivateKey {
	d := big.NewInt(n)
	x, y := curve.ScalarBaseMult(d.Bytes())
	return &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: curve, X: x, Y: y}, D: d}
}
func keysTLV(tag byte, content []byte) []byte {
	b, err := asn1.Marshal(asn1.RawValue{Class: int(tag >> 6), Tag: int(tag & 31), IsCompound: tag&32 != 0, Bytes: content})
	if err != nil {
		panic(err)
	}
	return b
}
func keysParts(t testing.TB, b []byte) [][]byte {
	t.Helper()
	var v asn1.RawValue
	rest, err := asn1.Unmarshal(b, &v)
	if err != nil || len(rest) != 0 {
		t.Fatal("fixture outer", err)
	}
	var out [][]byte
	for data := v.Bytes; len(data) > 0; {
		var child asn1.RawValue
		tail, e := asn1.Unmarshal(data, &child)
		if e != nil || len(tail) >= len(data) {
			t.Fatal("fixture child", e)
		}
		out = append(out, bytes.Clone(child.FullBytes))
		data = tail
	}
	return out
}
func keysJoin(parts [][]byte) []byte { return keysTLV(0x30, bytes.Join(parts, nil)) }
func keysIssuerSPKI(t testing.TB, kind string) []byte {
	t.Helper()
	var key any
	switch kind {
	case "p256":
		key = &keysTestEC(elliptic.P256(), 1).PublicKey
	case "p384":
		key = &keysTestEC(elliptic.P384(), 1).PublicKey
	case "p521":
		key = &keysTestEC(elliptic.P521(), 1).PublicKey
	case "rsa":
		key = &rsa.PublicKey{N: big.NewInt(65537), E: 65537}
	case "ed25519":
		key = ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize)).Public()
	default:
		t.Fatal("unknown fixture key kind", kind)
	}
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func keysExtensions(ski, aki []byte) []pkix.Extension {
	eku, err := asn1.Marshal([]asn1.ObjectIdentifier{{1, 3, 6, 1, 5, 5, 7, 3, 2}})
	if err != nil {
		panic(err)
	}
	uri := []byte("spiffe://node.test/node/11111111-2222-4333-8444-555555555555")
	return []pkix.Extension{
		{Id: asn1.ObjectIdentifier{2, 5, 29, 19}, Critical: true, Value: []byte{0x30, 0}},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 15}, Critical: true, Value: []byte{3, 2, 7, 0x80}},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 37}, Value: eku},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 17}, Critical: true, Value: keysTLV(0x30, keysTLV(0x86, uri))},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 14}, Value: keysTLV(4, ski)},
		{Id: asn1.ObjectIdentifier{2, 5, 29, 35}, Value: keysTLV(0x30, keysTLV(0x80, aki))},
	}
}
func keysFixture(t testing.TB) Input {
	return keysFixtureAt(t, time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC))
}
func keysFixtureAt(t testing.TB, start time.Time) Input {
	t.Helper()
	subject, issuer := keysTestEC(elliptic.P256(), 2), keysTestEC(elliptic.P256(), 1)
	subjectSPKI, err := x509.MarshalPKIXPublicKey(&subject.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	issuerSPKI := keysIssuerSPKI(t, "p256")
	ski, aki := sha256.Sum256(subjectSPKI), sha256.Sum256(issuerSPKI)
	parent := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "offline test issuer"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, SubjectKeyId: bytes.Clone(aki[:])}
	leaf := &x509.Certificate{SerialNumber: big.NewInt(128), NotBefore: start, NotAfter: start.Add(24 * time.Hour), ExtraExtensions: keysExtensions(ski[:], aki[:])}
	der, err := x509.CreateCertificate(rand.Reader, leaf, parent, &subject.PublicKey, issuer)
	if err != nil {
		t.Fatal(err)
	}
	return Input{Leaf: nodeleaf.Input{DER: der, ExpectedSPKI: subjectSPKI, NodeID: uuid.MustParse("11111111-2222-4333-8444-555555555555"), TrustDomain: "node.test"}, DeclaredIssuerSPKI: issuerSPKI}
}
func keysWithIDs(t testing.TB, in Input, ski, aki []byte) Input {
	t.Helper()
	cert := keysParts(t, in.Leaf.DER)
	fields := keysParts(t, cert[0])
	found := false
	for i, field := range fields {
		if field[0] == 0xa3 {
			exts := keysParts(t, keysParts(t, field)[0])
			for j, ext := range exts {
				p := keysParts(t, ext)
				var oid asn1.ObjectIdentifier
				rest, err := asn1.Unmarshal(p[0], &oid)
				if err != nil || len(rest) != 0 {
					t.Fatal("fixture OID")
				}
				if oid.Equal(asn1.ObjectIdentifier{2, 5, 29, 14}) {
					p[len(p)-1] = keysTLV(4, keysTLV(4, ski))
				}
				if oid.Equal(asn1.ObjectIdentifier{2, 5, 29, 35}) {
					p[len(p)-1] = keysTLV(4, keysTLV(0x30, keysTLV(0x80, aki)))
				}
				exts[j] = keysJoin(p)
			}
			fields[i] = keysTLV(0xa3, keysJoin(exts))
			found = true
		}
	}
	if !found {
		t.Fatal("fixture extensions")
	}
	cert[0] = keysJoin(fields)
	in.Leaf.DER = keysJoin(cert)
	in.Leaf.ExpectedSPKI = bytes.Clone(in.Leaf.ExpectedSPKI)
	in.DeclaredIssuerSPKI = bytes.Clone(in.DeclaredIssuerSPKI)
	return in
}
func keysWithSignatureBitFlipped(t testing.TB, in Input) Input {
	t.Helper()
	parts := keysParts(t, in.Leaf.DER)
	parts[2][len(parts[2])-1] ^= 1
	in.Leaf.DER = keysJoin(parts)
	return in
}
func rfcSPKI(t testing.TB) []byte {
	t.Helper()
	b, err := hex.DecodeString("3059301306072a8648ce3d020106082a8648ce3d030107034200047f7f35a79794c950060b8029fc8f363a28f11159692d9d34e6ac948190434735f833b1a66652dc514337aff7f5c9c75d670c019d95a5d639b72744c64a9128bb")
	if err != nil || len(b) != 91 {
		t.Fatal("RFC bytes", err)
	}
	return b
}
func TestKeysFixtureControls(t *testing.T) {
	in := keysFixture(t)
	c, err := x509.ParseCertificate(in.Leaf.DER)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &x509.Certificate{PublicKey: &keysTestEC(elliptic.P256(), 1).PublicKey, PublicKeyAlgorithm: x509.ECDSA, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	if err := c.CheckSignatureFrom(issuer); err != nil {
		t.Fatal("positive signature", err)
	}
	ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
	if !bytes.Equal(c.Raw, in.Leaf.DER) || !bytes.Equal(c.RawSubject, []byte{0x30, 0}) || !bytes.Equal(c.RawSubjectPublicKeyInfo, in.Leaf.ExpectedSPKI) || c.Version != 3 || c.SerialNumber.Int64() != 128 || !c.NotBefore.Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)) || !c.NotAfter.Equal(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("fixture fields")
	}
	if len(c.Extensions) != 6 || len(c.URIs) != 1 || c.URIs[0].String() != "spiffe://node.test/node/11111111-2222-4333-8444-555555555555" || !bytes.Equal(c.SubjectKeyId, ski[:]) || !bytes.Equal(c.AuthorityKeyId, aki[:]) {
		t.Fatal("fixture IDs/inventory")
	}
	for _, want := range keysExtensions(ski[:], aki[:]) {
		found := false
		for _, got := range c.Extensions {
			if got.Id.Equal(want.Id) {
				found = true
				if got.Critical != want.Critical || !bytes.Equal(got.Value, want.Value) {
					t.Fatal("raw extension", want.Id)
				}
			}
		}
		if !found {
			t.Fatal("missing extension")
		}
	}
	oracle := sha256.Sum256(rfcSPKI(t))
	if hex.EncodeToString(oracle[:]) != "6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405" {
		t.Fatal("independent RFC oracle")
	}
}
