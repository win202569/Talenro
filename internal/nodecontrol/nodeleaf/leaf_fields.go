package nodeleaf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"time"
)

func checkSPKIWire(der []byte) error {
	seq, err := singleElement(der, 0x30)
	if err != nil {
		return err
	}
	parts := seq.content
	alg, err := takeElement(&parts, 0x30)
	if err != nil {
		return err
	}
	point, err := takeElement(&parts, 3)
	if err != nil || len(parts) != 0 {
		return ErrInvalidMaterial
	}
	oids := alg.content
	algorithm, err := takeElement(&oids, 6)
	if err != nil {
		return err
	}
	curve, err := takeElement(&oids, 6)
	if err != nil || len(oids) != 0 {
		return ErrInvalidMaterial
	}
	if !bytes.Equal(algorithm.full, []byte{6, 7, 0x2a, 0x86, 0x48, 0xce, 0x3d, 2, 1}) || !bytes.Equal(curve.full, []byte{6, 8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 3, 1, 7}) || len(point.content) != 66 || point.content[0] != 0 || point.content[1] != 4 {
		return ErrUnsupportedLeafProfile
	}
	return nil
}

func p256Key(key any) bool {
	k, ok := key.(*ecdsa.PublicKey)
	return ok && k != nil && k.Curve == elliptic.P256() && k.X != nil && k.Y != nil && k.Curve.IsOnCurve(k.X, k.Y)
}

func checkExpectedSPKI(der []byte) error {
	if checkSPKIWire(der) != nil {
		return ErrInvalidInput
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil || !p256Key(key) {
		return ErrInvalidInput
	}
	return nil
}

func checkLeafFields(cert *x509.Certificate, raw rawLeaf) ([20]byte, uint8, error) {
	var serial [20]byte
	if cert == nil || cert.Version != 3 || !bytes.Equal(raw.subject.full, []byte{0x30, 0}) || !p256Key(cert.PublicKey) {
		return serial, 0, ErrUnsupportedLeafProfile
	}
	if err := checkSPKIWire(raw.spki.full); err != nil {
		return serial, 0, err
	}
	if !bytes.Equal(raw.spki.full, cert.RawSubjectPublicKeyInfo) {
		return serial, 0, ErrInvalidMaterial
	}
	if raw.serial.tag != 2 || !minimalInteger(raw.serial.content) {
		return serial, 0, ErrInvalidMaterial
	}
	b := raw.serial.content
	if b[0]&128 != 0 || len(b) > 20 || (len(b) == 1 && b[0] == 0) || cert.SerialNumber == nil || cert.SerialNumber.Sign() <= 0 {
		return serial, 0, ErrUnsupportedLeafProfile
	}
	if b[0] == 0 {
		b = b[1:]
	}
	if !bytes.Equal(b, cert.SerialNumber.Bytes()) {
		return serial, 0, ErrInvalidMaterial
	}
	if !canonicalTime(raw.notBefore, cert.NotBefore) || !canonicalTime(raw.notAfter, cert.NotAfter) || !cert.NotBefore.Before(cert.NotAfter) || !cert.NotAfter.Equal(cert.NotBefore.Add(24*time.Hour)) {
		return serial, 0, ErrUnsupportedLeafProfile
	}
	copy(serial[:], b)
	return serial, uint8(len(b)), nil
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
		tag, layout = 0x17, "060102150405Z"
	}
	return raw.tag == tag && bytes.Equal(raw.content, []byte(parsed.UTC().Format(layout)))
}
