package nodeleaf

import (
	"bytes"
	"errors"
	"testing"
)

func TestLeafDERGrammar(t *testing.T) {
	for _, raw := range [][]byte{nil, {0x30}, {0x30, 0x80, 0, 0}, {4, 0x81, 1, 1}, {4, 0x82, 0, 0x80}, {4, 2, 1}, {0x1f, 0}} {
		if _, _, e := readElement(raw); !errors.Is(e, ErrInvalidMaterial) {
			t.Fatal("bad TLV", raw, e)
		}
	}
	v, tail, e := readElement([]byte{4, 1, 0x80, 5, 0})
	if e != nil || v.tag != 4 || !bytes.Equal(v.content, []byte{0x80}) || !bytes.Equal(tail, []byte{5, 0}) {
		t.Fatal("TLV range")
	}
	in := leafFixture(t)
	raw, e := parseRawLeaf(in.DER)
	if e != nil || !bytes.Equal(raw.serial.content, []byte{0, 0x80}) || !bytes.Equal(raw.subject.full, []byte{0x30, 0}) || !bytes.Equal(raw.spki.full, in.ExpectedSPKI) || len(raw.extensions) != 6 {
		t.Fatal("raw leaf", e)
	}
	for _, b := range [][]byte{append(bytes.Clone(in.DER), 0), {0x30, 0}, {1, 2, 3}, fixtureTLV(0x30, append(fixtureParts(t, in.DER)[0], 0))} {
		if _, e := parseRawLeaf(b); e == nil {
			t.Fatal("bad certificate accepted")
		}
	}
	cert := fixtureParts(t, in.DER)
	fields := fixtureParts(t, cert[0])
	fields = append(fields[:7], append([][]byte{{0x81, 1, 0}}, fields[7:]...)...)
	cert[0] = fixtureJoin(fields)
	if _, e := parseRawLeaf(fixtureJoin(cert)); !errors.Is(e, ErrUnsupportedLeafProfile) {
		t.Fatal("unique identifier", e)
	}
	cert = fixtureParts(t, in.DER)
	fields = fixtureParts(t, cert[0])
	fields = append(fields, []byte{5, 0})
	cert[0] = fixtureJoin(fields)
	if _, e := parseRawLeaf(fixtureJoin(cert)); e == nil {
		t.Fatal("TBS trailing")
	}
	for _, version := range [][]byte{nil, {0xa0, 3, 2, 1, 1}} {
		variant := rewriteFixture(t, in, "extensions", nil)
		variant = rewriteFixture(t, variant, "version", version)
		if _, e := parseRawLeaf(variant.DER); !errors.Is(e, ErrUnsupportedLeafProfile) {
			t.Fatal("unsupported version", e)
		}
	}
}
func TestLeafLengthOverflow(t *testing.T) {
	for _, b := range [][]byte{{4, 0x88, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, {4, 0x88, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, {4, 0x84, 1, 0, 0, 0}, {4, 0xff}} {
		if _, _, e := readElement(b); !errors.Is(e, ErrInvalidMaterial) {
			t.Fatal("length overflow", e)
		}
	}
}
