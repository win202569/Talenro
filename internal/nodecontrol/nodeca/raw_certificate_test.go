package nodeca

import (
	"bytes"
	"errors"
	"testing"
)

func TestCARawCertificate(t *testing.T) {
	for _, parent := range []bool{false, true} {
		in := fixtureInput(parent)
		c := fixtureCert(t, in)
		r, e := parseRawCA(in.DER)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(r.tbs.full, c.RawTBSCertificate) || !bytes.Equal(r.spki.full, c.RawSubjectPublicKeyInfo) || !bytes.Equal(r.subject.full, c.RawSubject) || !bytes.Equal(r.issuer.full, c.RawIssuer) || len(r.extensions) != 4 || !bytes.Equal(r.innerSignatureAlgorithm.full, r.outerSignatureAlgorithm.full) {
			t.Fatal("raw projection")
		}
	}
	in := fixtureInput(false)
	cert := fixtureParts(t, in.DER)
	fields := fixtureParts(t, cert[0])
	cases := []struct {
		name string
		in   Input
		want error
	}{
		{"tail", fixtureRewrite(t, in, "certificate", append(bytes.Clone(in.DER), 0)), ErrInvalidMaterial},
		{"concat", fixtureRewrite(t, in, "certificate", append(bytes.Clone(in.DER), in.DER...)), ErrInvalidMaterial},
		{"truncated", fixtureRewrite(t, in, "certificate", in.DER[:len(in.DER)-1]), ErrInvalidMaterial},
		{"outer algorithm", fixtureRewrite(t, in, "outerAlgorithm", fixtureHex("300906072a8648ce3d0403")), ErrInvalidMaterial},
		{"empty sig", fixtureRewrite(t, in, "signature", []byte{3, 1, 0}), ErrUnsupportedCAProfile},
		{"unused sig", fixtureRewrite(t, in, "signature", []byte{3, 2, 1, 2}), ErrUnsupportedCAProfile},
		{"bad unused", fixtureRewrite(t, in, "signature", []byte{3, 2, 8, 0}), ErrInvalidMaterial},
		{"bad last bits", fixtureRewrite(t, in, "signature", []byte{3, 2, 1, 3}), ErrInvalidMaterial},
		{"unique ID", fixtureRewrite(t, in, "tbs", fixtureJoin(append(append([][]byte{}, fields[:7]...), append([][]byte{{0x81, 1, 0}}, fields[7:]...)...))), ErrUnsupportedCAProfile},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, e := parseRawCA(c.in.DER); !errors.Is(e, c.want) {
				t.Fatalf("got %v want %v", e, c.want)
			}
		})
	}
}
func TestCARawNestedTails(t *testing.T) {
	in := fixtureInput(false)
	fields := fixtureParts(t, fixtureParts(t, in.DER)[0])
	name := fixtureParts(t, fields[5])
	rdn := fixtureParts(t, name[0])
	atv := fixtureParts(t, rdn[0])
	extWrap := fixtureParts(t, fields[7])[0]
	exts := fixtureParts(t, extWrap)
	ext := fixtureParts(t, exts[0])
	times := fixtureParts(t, fields[4])
	cases := []struct {
		name, path string
		b          []byte
	}{
		{"empty RDN", "subject", fixtureTLV(0x30, []byte{0x31, 0})},
		{"ATV tail", "subject", fixtureTLV(0x30, fixtureTLV(0x31, fixtureJoin(append(atv, []byte{5, 0}))))},
		{"RDN wrong tag", "subject", fixtureJoin([][]byte{rdn[0]})},
		{"valid empty Name", "subject", []byte{0x30, 0}},
		{"validity tail", "tbs", fixtureJoin(append(append(append([][]byte{}, fields[:4]...), fixtureJoin(append(times, []byte{5, 0}))), fields[5:]...))},
		{"TBS tail", "tbs", fixtureJoin(append(fields, []byte{5, 0}))},
		{"extension tail", "extensions", fixtureTLV(0xa3, fixtureJoin(append([][]byte{fixtureJoin(append(ext, []byte{5, 0}))}, exts[1:]...)))},
		{"explicit false", "extensions", fixtureTLV(0xa3, fixtureJoin(append([][]byte{fixtureJoin([][]byte{ext[0], {1, 1, 0}, ext[len(ext)-1]})}, exts[1:]...)))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, e := parseRawCA(fixtureRewrite(t, in, c.path, c.b).DER)
			if c.name == "valid empty Name" {
				if e != nil {
					t.Fatal(e)
				}
				return
			}
			if !errors.Is(e, ErrInvalidMaterial) {
				t.Fatal("accepted nested tail", e)
			}
		})
	}
}
