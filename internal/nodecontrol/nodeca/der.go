package nodeca

import "encoding/asn1"

type derElement struct {
	tag           byte
	full, content []byte
}
type rawExtension struct {
	id       asn1.ObjectIdentifier
	critical bool
	value    []byte
}
type rawCA struct {
	tbs, serial, issuer, subject, spki, notBefore, notAfter     derElement
	innerSignatureAlgorithm, outerSignatureAlgorithm, signature derElement
	extensions                                                  []rawExtension
}

func readElement(raw []byte) (derElement, []byte, error) {
	if len(raw) < 2 || raw[0]&31 == 31 || raw[0] == 0 {
		return derElement{}, nil, ErrInvalidMaterial
	}
	header, size := 2, int(raw[1])
	if raw[1]&128 != 0 {
		n := int(raw[1] & 127)
		if n == 0 || n > 8 || n > len(raw)-2 || raw[2] == 0 {
			return derElement{}, nil, ErrInvalidMaterial
		}
		header += n
		size = 0
		remaining := len(raw) - header
		for _, b := range raw[2:header] {
			if int(b) > remaining || size > (remaining-int(b))/256 {
				return derElement{}, nil, ErrInvalidMaterial
			}
			size = size*256 + int(b)
		}
		if size < 128 {
			return derElement{}, nil, ErrInvalidMaterial
		}
	}
	if size > len(raw)-header {
		return derElement{}, nil, ErrInvalidMaterial
	}
	end := header + size
	return derElement{tag: raw[0], full: raw[:end], content: raw[header:end]}, raw[end:], nil
}
func singleElement(raw []byte, tag byte) (derElement, error) {
	v, tail, e := readElement(raw)
	if e != nil || v.tag != tag || len(tail) != 0 {
		return derElement{}, ErrInvalidMaterial
	}
	return v, nil
}
func takeElement(raw *[]byte, tag byte) (derElement, error) {
	v, tail, e := readElement(*raw)
	if e != nil || v.tag != tag {
		return derElement{}, ErrInvalidMaterial
	}
	*raw = tail
	return v, nil
}
func minimalInteger(b []byte) bool {
	return len(b) > 0 && (len(b) == 1 || !((b[0] == 0 && b[1]&128 == 0) || (b[0] == 255 && b[1]&128 != 0)))
}

// Opaque OID encodings have no component-size policy. Actual field OIDs that
// are decoded into Go identifiers continue to use readOID below.
func minimalOIDContent(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	start := true
	for _, v := range b {
		if start && v == 0x80 {
			return false
		}
		start = v&0x80 == 0
	}
	return start
}
func readOID(raw []byte) (asn1.ObjectIdentifier, error) {
	var id asn1.ObjectIdentifier
	tail, e := asn1.Unmarshal(raw, &id)
	if e != nil || len(tail) != 0 {
		return nil, ErrInvalidMaterial
	}
	return id, nil
}
func bitStringShape(b []byte) bool {
	return len(b) > 0 && b[0] <= 7 && (len(b) > 1 || b[0] == 0) && (len(b) == 1 || b[len(b)-1]&byte((1<<b[0])-1) == 0)
}

// DER SET OF ordering compares complete encodings without normalizing them.
func derLess(a, b []byte) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}

// An iterative walk avoids stack growth without imposing an unapproved depth
// limit on otherwise legal unknown signer parameters. Input budgets still apply.
func checkDERValue(v derElement) error {
	pending := []derElement{v}
	for len(pending) > 0 {
		v = pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if v.tag&0xc0 == 0 {
			tag := v.tag & 31
			constructed := tag == 8 || tag == 11 || tag == 16 || tag == 17 || tag == 29
			if tag == 0 || tag == 15 || (v.tag&32 != 0) != constructed {
				return ErrInvalidMaterial
			}
		}
		if v.tag&32 != 0 {
			// Opaque signer parameters have no SET versus SET OF schema here.
			// RDN SET OF ordering is checked separately in checkName.
			for rest := v.content; len(rest) > 0; {
				child, tail, e := readElement(rest)
				if e != nil {
					return e
				}
				pending = append(pending, child)
				rest = tail
			}
			continue
		}
		switch v.tag {
		case 1:
			if len(v.content) != 1 || (v.content[0] != 0 && v.content[0] != 255) {
				return ErrInvalidMaterial
			}
		case 2, 10:
			if !minimalInteger(v.content) {
				return ErrInvalidMaterial
			}
		case 3:
			if !bitStringShape(v.content) {
				return ErrInvalidMaterial
			}
		case 5:
			if len(v.content) != 0 {
				return ErrInvalidMaterial
			}
		case 6:
			if !minimalOIDContent(v.content) {
				return ErrInvalidMaterial
			}
		}
	}
	return nil
}
