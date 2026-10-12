package nodeleaf

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
)

func checkExtensions(raw []rawExtension, cert *x509.Certificate) ([]byte, error) {
	ids := [6]asn1.ObjectIdentifier{{2, 5, 29, 19}, {2, 5, 29, 15}, {2, 5, 29, 37}, {2, 5, 29, 17}, {2, 5, 29, 14}, {2, 5, 29, 35}}
	critical := [6]bool{true, true, false, true, false, false}
	var slots [6]*rawExtension
	if len(raw) != 6 || cert == nil {
		return nil, ErrUnsupportedLeafProfile
	}
	for i := range raw {
		index := -1
		for j := range ids {
			if raw[i].id.Equal(ids[j]) {
				index = j
				break
			}
		}
		if index < 0 || slots[index] != nil || raw[i].critical != critical[index] {
			return nil, ErrUnsupportedLeafProfile
		}
		slots[index] = &raw[i]
	}
	if len(cert.Extensions) != 6 {
		return nil, ErrInvalidMaterial
	}
	for i := range slots {
		matches := 0
		for _, ext := range cert.Extensions {
			if ext.Id.Equal(ids[i]) {
				matches++
				if ext.Critical != critical[i] || !bytes.Equal(ext.Value, slots[i].value) {
					return nil, ErrInvalidMaterial
				}
			}
		}
		if matches != 1 {
			return nil, ErrInvalidMaterial
		}
	}
	// The profile checks below are intentionally ordered, not map-dependent.
	if _, err := singleElement(slots[0].value, 0x30); err != nil {
		return nil, err
	}
	if !bytes.Equal(slots[0].value, []byte{0x30, 0}) || !cert.BasicConstraintsValid || cert.IsCA {
		return nil, ErrUnsupportedLeafProfile
	}
	if _, err := singleElement(slots[1].value, 3); err != nil {
		return nil, err
	}
	if !bytes.Equal(slots[1].value, []byte{3, 2, 7, 0x80}) || cert.KeyUsage != x509.KeyUsageDigitalSignature {
		return nil, ErrUnsupportedLeafProfile
	}
	if _, err := singleElement(slots[2].value, 0x30); err != nil {
		return nil, err
	}
	if !bytes.Equal(slots[2].value, []byte{0x30, 10, 6, 8, 0x2b, 6, 1, 5, 5, 7, 3, 2}) || len(cert.ExtKeyUsage) != 1 || cert.ExtKeyUsage[0] != x509.ExtKeyUsageClientAuth || len(cert.UnknownExtKeyUsage) != 0 {
		return nil, ErrUnsupportedLeafProfile
	}
	san, err := singleElement(slots[3].value, 0x30)
	if err != nil {
		return nil, err
	}
	uri, tail, err := readElement(san.content)
	if err != nil {
		return nil, err
	}
	if uri.tag != 0x86 || len(tail) != 0 || len(uri.content) == 0 || len(uri.content) > 304 || len(cert.URIs) != 1 || cert.URIs[0] == nil || len(cert.DNSNames) != 0 || len(cert.EmailAddresses) != 0 || len(cert.IPAddresses) != 0 {
		return nil, ErrUnsupportedLeafProfile
	}
	for _, b := range uri.content {
		if b > 127 {
			return nil, ErrUnsupportedLeafProfile
		}
	}
	ski, err := singleElement(slots[4].value, 4)
	if err != nil {
		return nil, err
	}
	if len(ski.content) == 0 || len(ski.content) > 64 {
		return nil, ErrUnsupportedLeafProfile
	}
	if !bytes.Equal(ski.content, cert.SubjectKeyId) {
		return nil, ErrInvalidMaterial
	}
	aki, err := singleElement(slots[5].value, 0x30)
	if err != nil {
		return nil, err
	}
	keyID, tail, err := readElement(aki.content)
	if err != nil {
		return nil, err
	}
	if keyID.tag != 0x80 || len(tail) != 0 || len(keyID.content) == 0 || len(keyID.content) > 64 {
		return nil, ErrUnsupportedLeafProfile
	}
	if !bytes.Equal(keyID.content, cert.AuthorityKeyId) {
		return nil, ErrInvalidMaterial
	}
	return uri.content, nil
}
