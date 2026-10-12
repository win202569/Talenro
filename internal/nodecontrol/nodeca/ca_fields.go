package nodeca

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"time"
)

type fieldSummary struct {
	Serial              [20]byte
	SerialLength        uint8
	NotBefore, NotAfter time.Time
}

func canonicalTime(raw derElement, parsed time.Time) bool {
	y := parsed.Year()
	if y < 1 || y > 9999 || parsed.Nanosecond() != 0 {
		return false
	}
	_, offset := parsed.Zone()
	if offset != 0 {
		return false
	}
	tag, layout := byte(0x18), "20060102150405Z"
	if y >= 1950 && y <= 2049 {
		tag = 0x17
		layout = "060102150405Z"
	}
	return raw.tag == tag && bytes.Equal(raw.content, []byte(parsed.UTC().Format(layout)))
}
func checkTimePair(raw derElement, parsed time.Time) error {
	var decoded time.Time
	tail, e := asn1.Unmarshal(raw.full, &decoded)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	if !decoded.Equal(parsed) {
		return ErrInvalidMaterial
	}
	if !canonicalTime(raw, parsed) {
		return ErrUnsupportedCAProfile
	}
	return nil
}
func checkCAFields(cert *x509.Certificate, raw rawCA) (fieldSummary, error) {
	var out fieldSummary
	if cert == nil {
		return out, ErrInvalidMaterial
	}
	if cert.Version != 3 {
		return out, ErrUnsupportedCAProfile
	}
	if e := checkCASPKI(cert, raw); e != nil {
		return out, e
	}
	if !bytes.Equal(raw.issuer.full, cert.RawIssuer) || !bytes.Equal(raw.subject.full, cert.RawSubject) {
		return out, ErrInvalidMaterial
	}
	if len(raw.issuer.content) == 0 || len(raw.subject.content) == 0 || len(raw.issuer.full) > 4096 || len(raw.subject.full) > 4096 {
		return out, ErrUnsupportedCAProfile
	}
	if raw.serial.tag != 2 || !minimalInteger(raw.serial.content) {
		return out, ErrInvalidMaterial
	}
	serial := raw.serial.content
	if len(serial) > 20 || serial[0]&128 != 0 || (len(serial) == 1 && serial[0] == 0) {
		return out, ErrUnsupportedCAProfile
	}
	if serial[0] == 0 {
		serial = serial[1:]
	}
	if cert.SerialNumber == nil || cert.SerialNumber.Sign() <= 0 || !bytes.Equal(serial, cert.SerialNumber.Bytes()) {
		return out, ErrInvalidMaterial
	}
	if e := checkTimePair(raw.notBefore, cert.NotBefore); e != nil {
		return out, e
	}
	if e := checkTimePair(raw.notAfter, cert.NotAfter); e != nil {
		return out, e
	}
	if !cert.NotBefore.Before(cert.NotAfter) {
		return out, ErrUnsupportedCAProfile
	}
	copy(out.Serial[:], serial)
	out.SerialLength = uint8(len(serial))
	out.NotBefore = cert.NotBefore.UTC()
	out.NotAfter = cert.NotAfter.UTC()
	return out, nil
}
