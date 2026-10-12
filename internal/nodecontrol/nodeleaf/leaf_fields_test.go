package nodeleaf

import (
	"bytes"
	"crypto/x509"
	"errors"
	"math/big"
	"testing"
	"time"
)

func TestLeafExpectedSPKI(t *testing.T) {
	in := leafFixture(t)
	if e := checkExpectedSPKI(in.ExpectedSPKI); e != nil {
		t.Fatal("valid expected", e)
	}
	bad := fixtureNonP256SPKI(t)
	bad = append(bad, nil, []byte{1}, append(bytes.Clone(in.ExpectedSPKI), 0), append([]byte{0x30, 0x81, in.ExpectedSPKI[1]}, in.ExpectedSPKI[2:]...))
	zeroPoint := bytes.Clone(in.ExpectedSPKI)
	clear(zeroPoint[len(zeroPoint)-64:])
	bad = append(bad, zeroPoint)
	for _, b := range bad {
		if !errors.Is(checkExpectedSPKI(b), ErrInvalidInput) {
			t.Fatal("unsupported expected SPKI")
		}
	}
}
func TestLeafSerialPadding(t *testing.T) {
	in := leafFixture(t)
	c, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	r, e := parseRawLeaf(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		raw, mag []byte
		n        uint8
		want     error
	}{
		{[]byte{0, 0x80}, []byte{0x80}, 1, nil},
		{bytes.Repeat([]byte{0x7f}, 20), bytes.Repeat([]byte{0x7f}, 20), 20, nil},
		{append([]byte{0}, bytes.Repeat([]byte{0x80}, 20)...), bytes.Repeat([]byte{0x80}, 20), 0, ErrUnsupportedLeafProfile},
		{[]byte{0}, nil, 0, ErrUnsupportedLeafProfile}, {[]byte{0xff}, []byte{1}, 0, ErrUnsupportedLeafProfile},
		{[]byte{0, 1}, []byte{1}, 0, ErrInvalidMaterial}, {[]byte{0, 0, 0x80}, []byte{0x80}, 0, ErrInvalidMaterial},
	} {
		cc := *c
		cc.SerialNumber = new(big.Int).SetBytes(tc.mag)
		rr := r
		rr.serial = derElement{tag: 2, content: tc.raw, full: fixtureTLV(2, tc.raw)}
		serial, n, e := checkLeafFields(&cc, rr)
		if !errors.Is(e, tc.want) {
			t.Fatal("serial error", tc.raw, e, tc.want)
		}
		if e == nil && (n != tc.n || !bytes.Equal(serial[:n], tc.mag) || !bytes.Equal(serial[n:], make([]byte, 20-int(n)))) {
			t.Fatal("serial magnitude")
		}
	}
}
func TestLeafFieldWindow(t *testing.T) {
	in := leafFixture(t)
	c, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	r, e := parseRawLeaf(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e := checkLeafFields(c, r); e != nil {
		t.Fatal(e)
	}
	for _, y := range []int{1, 1949, 1950, 2049, 2050, 9999} {
		cc := *c
		cc.NotBefore = time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		cc.NotAfter = cc.NotBefore.Add(24 * time.Hour)
		rr := r
		tag := byte(0x18)
		layout := "20060102150405Z"
		if y >= 1950 && y <= 2049 {
			tag = 0x17
			layout = "060102150405Z"
		}
		rr.notBefore = derElement{tag: tag, content: []byte(cc.NotBefore.Format(layout))}
		rr.notAfter = derElement{tag: tag, content: []byte(cc.NotAfter.Format(layout))}
		if _, _, e := checkLeafFields(&cc, rr); e != nil {
			t.Fatal("year boundary", y, e)
		}
	}
	for _, end := range []time.Time{c.NotBefore, c.NotBefore.Add(-time.Second), c.NotAfter.Add(-time.Second), c.NotAfter.Add(time.Second)} {
		cc := *c
		cc.NotAfter = end
		if _, _, e := checkLeafFields(&cc, r); e == nil {
			t.Fatal("window accepted")
		}
	}
	for _, raw := range []derElement{{tag: 0x17, content: []byte("261011000000+0000")}, {tag: 0x18, content: []byte("20261011000000.1Z")}, {tag: 0x18, content: []byte("20261011000000Z")}, {tag: 0x18, content: []byte("00001011000000Z")}, {tag: 0x16, content: []byte("261011000000Z")}} {
		rr := r
		rr.notBefore = raw
		if _, _, e := checkLeafFields(c, rr); e == nil {
			t.Fatal("noncanonical time accepted")
		}
	}
	cc := *c
	cc.NotBefore = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, _, e := checkLeafFields(&cc, r); e == nil {
		t.Fatal("year zero")
	}
	cc = *c
	cc.NotAfter = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, _, e := checkLeafFields(&cc, r); e == nil {
		t.Fatal("year overflow")
	}
	for _, subject := range [][]byte{{0x30, 2, 0x31, 0}, {0x31, 0}} {
		rr := r
		rr.subject = derElement{tag: subject[0], full: subject}
		if _, _, e := checkLeafFields(c, rr); e == nil {
			t.Fatal("subject accepted")
		}
	}
}
