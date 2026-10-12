package nodeleafkeys

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
)

func keysLower(t testing.TB, in Input) nodeleaf.Summary {
	t.Helper()
	m, err := nodeleaf.Parse(context.Background(), in.Leaf)
	if err != nil {
		t.Fatal("lower fixture", err)
	}
	defer m.Clear()
	return m.Summary()
}
func keysCert(t testing.TB, in Input) *x509.Certificate {
	t.Helper()
	c, err := x509.ParseCertificate(in.Leaf.DER)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestKeysLowerSummary(t *testing.T) {
	good := keysLower(t, keysFixture(t))
	if checkLowerSummary(good) != nil {
		t.Fatal("good lower")
	}
	variants := []nodeleaf.Summary{{}}
	v := good
	v.Proof = nodeleaf.None
	variants = append(variants, v)
	v = good
	v.Proof = 99
	variants = append(variants, v)
	for i := range good.Pending {
		v = good
		v.Pending[i] = 0
		variants = append(variants, v)
		v = good
		v.Pending[i] = 99
		variants = append(variants, v)
	}
	v = good
	v.Pending[0], v.Pending[1] = v.Pending[1], v.Pending[0]
	variants = append(variants, v)
	for _, s := range variants {
		if checkLowerSummary(s) != ErrInvalidMaterial {
			t.Fatal("contradictory lower")
		}
	}
}
func TestKeysLeafSnapshot(t *testing.T) {
	in := keysFixture(t)
	s := keysLower(t, in)
	c, err := checkLeafSnapshot(in.Leaf.DER, in.Leaf.ExpectedSPKI, s)
	if err != nil || c == nil || !bytes.Equal(c.Raw, in.Leaf.DER) || !bytes.Equal(c.RawSubjectPublicKeyInfo, in.Leaf.ExpectedSPKI) {
		t.Fatal("snapshot", err)
	}
	for _, tc := range []struct {
		der, spki []byte
		summary   nodeleaf.Summary
	}{
		{append(bytes.Clone(in.Leaf.DER), 0), in.Leaf.ExpectedSPKI, s},
		{[]byte{0}, in.Leaf.ExpectedSPKI, s},
		{in.Leaf.DER, in.DeclaredIssuerSPKI, s},
		{keysWithSignatureBitFlipped(t, in).Leaf.DER, in.Leaf.ExpectedSPKI, s},
	} {
		if c, e := checkLeafSnapshot(tc.der, tc.spki, tc.summary); c != nil || e != ErrInvalidMaterial {
			t.Fatal("snapshot mismatch", e)
		}
	}
	for _, which := range []string{"der", "spki"} {
		bad := s
		if which == "der" {
			bad.LeafDERDigest[0] ^= 1
		} else {
			bad.SPKIDERDigest[0] ^= 1
		}
		if c, e := checkLeafSnapshot(in.Leaf.DER, in.Leaf.ExpectedSPKI, bad); c != nil || e != ErrInvalidMaterial {
			t.Fatal("digest mismatch")
		}
	}
}
func TestKeysRFC7093(t *testing.T) {
	got := keyIdentifierForSPKI(rfcSPKI(t))
	if hex.EncodeToString(got[:]) != "6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405" {
		t.Fatal("full SPKI oracle")
	}
}
func keysAssertIDsError(t testing.TB, in Input, want error) {
	t.Helper()
	a, b, d, err := checkKeyIdentifiers(keysCert(t, in), in.DeclaredIssuerSPKI)
	if err != want || a != (KeyIdentifier{}) || b != (KeyIdentifier{}) || d != (contracts.Digest{}) {
		t.Fatalf("ID error/zero result: %v want %v", err, want)
	}
}
func TestKeysIdentifierLengths(t *testing.T) {
	in := keysFixture(t)
	ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
	for _, role := range []string{"ski", "aki"} {
		for _, n := range []int{1, 20, 31, 32, 33, 64} {
			a, b := bytes.Clone(ski[:]), bytes.Clone(aki[:])
			if role == "ski" {
				a = make([]byte, n)
				copy(a, ski[:])
			} else {
				b = make([]byte, n)
				copy(b, aki[:])
			}
			v := keysWithIDs(t, in, a, b)
			// Every tested length remains valid under the old 1..64 shape contract.
			keysLower(t, v)
			if n != 32 {
				keysAssertIDsError(t, v, ErrUnsupportedKeyIdentifierProfile)
				continue
			}
			x, y, d, e := checkKeyIdentifiers(keysCert(t, v), v.DeclaredIssuerSPKI)
			if e != nil || x != KeyIdentifier(ski) || y != KeyIdentifier(aki) || d != contracts.Digest(aki) {
				t.Fatal("32 true IDs", e)
			}
		}
	}
}
func keysRawBytes(t testing.TB, b []byte) []byte {
	t.Helper()
	var v asn1.RawValue
	rest, e := asn1.Unmarshal(b, &v)
	if e != nil || len(rest) != 0 {
		t.Fatal(e)
	}
	return v.Bytes
}
func TestKeysIdentifierBinding(t *testing.T) {
	in := keysFixture(t)
	ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
	for _, role := range []string{"ski", "aki"} {
		raw := in.Leaf.ExpectedSPKI
		if role == "aki" {
			raw = in.DeclaredIssuerSPKI
		}
		parts := keysParts(t, raw)
		var bitContent []byte
		// Independent raw TLV decoding: the bit-string payload begins after unused count.
		var bs = keysRawBytes(t, parts[1])
		bitContent = bytes.Clone(bs[1:])
		short := sha1.Sum(bitContent)
		bitHash := sha256.Sum256(parts[1])
		pointHash := sha256.Sum256(bitContent)
		leafHash := sha256.Sum256(in.Leaf.DER)
		prefixed := sha256.Sum256(append([]byte("node-key-id:"), raw...))
		trueID := ski
		if role == "aki" {
			trueID = aki
		}
		wrong := trueID
		wrong[0] ^= 1
		swapped := aki
		if role == "aki" {
			swapped = ski
		}
		variants := [][]byte{wrong[:], swapped[:], short[:], trueID[:20], bitHash[:], pointHash[:], leafHash[:], prefixed[:], []byte(hex.EncodeToString(trueID[:])), []byte(base64.StdEncoding.EncodeToString(trueID[:]))}
		for _, bad := range variants {
			a, b := ski[:], aki[:]
			if role == "ski" {
				a = bad
			} else {
				b = bad
			}
			v := keysWithIDs(t, in, a, b)
			want := ErrBindingMismatch
			if len(bad) != 32 {
				want = ErrUnsupportedKeyIdentifierProfile
			}
			keysAssertIDsError(t, v, want)
		}
	}
}
func TestKeysSameSPKI(t *testing.T) {
	in := keysFixture(t)
	in.DeclaredIssuerSPKI = bytes.Clone(in.Leaf.ExpectedSPKI)
	h := sha256.Sum256(in.Leaf.ExpectedSPKI)
	in = keysWithIDs(t, in, h[:], h[:])
	a, b, d, e := checkKeyIdentifiers(keysCert(t, in), in.DeclaredIssuerSPKI)
	if e != nil || a != b || a != KeyIdentifier(h) || d != contracts.Digest(h) {
		t.Fatal("equal roles valid relation", e)
	}
}
func TestKeysIdentifierPrecedence(t *testing.T) {
	in := keysFixture(t)
	ski, aki := sha256.Sum256(in.Leaf.ExpectedSPKI), sha256.Sum256(in.DeclaredIssuerSPKI)
	ski[0] ^= 1
	keysAssertIDsError(t, keysWithIDs(t, in, ski[:], aki[:20]), ErrUnsupportedKeyIdentifierProfile)
	aki[0] ^= 1
	keysAssertIDsError(t, keysWithIDs(t, in, ski[:], aki[:]), ErrBindingMismatch)
	a, b, d, e := checkKeyIdentifiers(nil, in.DeclaredIssuerSPKI)
	if e != ErrInvalidMaterial || a != (KeyIdentifier{}) || b != (KeyIdentifier{}) || d != (contracts.Digest{}) {
		t.Fatal("nil certificate")
	}
}
