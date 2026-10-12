package nodeleaf

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

type rawLeaf struct {
	serial, subject, spki derElement
	notBefore, notAfter   derElement
	extensions            []rawExtension
}

// readElement is deliberately limited to the single-byte tags in this profile.
func readElement(raw []byte) (derElement, []byte, error) {
	if len(raw) < 2 || raw[0]&31 == 31 {
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
			// Bound the accumulator by actual remaining input before multiplying.
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
	v, tail, err := readElement(raw)
	if err != nil || v.tag != tag || len(tail) != 0 {
		return derElement{}, ErrInvalidMaterial
	}
	return v, nil
}

func takeElement(raw *[]byte, tag byte) (derElement, error) {
	v, tail, err := readElement(*raw)
	if err != nil || v.tag != tag {
		return derElement{}, ErrInvalidMaterial
	}
	*raw = tail
	return v, nil
}

func parseRawLeaf(der []byte) (rawLeaf, error) {
	var out rawLeaf
	outer, err := singleElement(der, 0x30)
	if err != nil {
		return out, err
	}
	parts := outer.content
	tbs, err := takeElement(&parts, 0x30)
	if err != nil {
		return out, err
	}
	if _, err = takeElement(&parts, 0x30); err != nil {
		return out, err
	}
	if _, err = takeElement(&parts, 3); err != nil || len(parts) != 0 {
		return out, ErrInvalidMaterial
	}
	fields := tbs.content
	if len(fields) == 0 {
		return out, ErrInvalidMaterial
	}
	if fields[0] != 0xa0 {
		if fields[0] == 2 {
			return out, ErrUnsupportedLeafProfile
		}
		return out, ErrInvalidMaterial
	}
	version, err := takeElement(&fields, 0xa0)
	if err != nil {
		return out, err
	}
	v, err := singleElement(version.content, 2)
	if err != nil || !minimalInteger(v.content) {
		return out, ErrInvalidMaterial
	}
	if len(v.content) != 1 || v.content[0] != 2 {
		return out, ErrUnsupportedLeafProfile
	}
	if out.serial, err = takeElement(&fields, 2); err != nil {
		return out, err
	}
	if _, err = takeElement(&fields, 0x30); err != nil {
		return out, err
	} // signature algorithm
	if _, err = takeElement(&fields, 0x30); err != nil {
		return out, err
	} // issuer
	validity, err := takeElement(&fields, 0x30)
	if err != nil {
		return out, err
	}
	times := validity.content
	out.notBefore, times, err = readElement(times)
	if err != nil {
		return out, err
	}
	out.notAfter, times, err = readElement(times)
	if err != nil || len(times) != 0 {
		return out, ErrInvalidMaterial
	}
	if (out.notBefore.tag != 0x17 && out.notBefore.tag != 0x18) || (out.notAfter.tag != 0x17 && out.notAfter.tag != 0x18) {
		return out, ErrInvalidMaterial
	}
	if out.subject, err = takeElement(&fields, 0x30); err != nil {
		return out, err
	}
	if out.spki, err = takeElement(&fields, 0x30); err != nil {
		return out, err
	}
	if len(fields) != 0 && (fields[0] == 0x81 || fields[0] == 0x82) {
		return out, ErrUnsupportedLeafProfile
	}
	ext, err := takeElement(&fields, 0xa3)
	if err != nil || len(fields) != 0 {
		return out, ErrInvalidMaterial
	}
	seq, err := singleElement(ext.content, 0x30)
	if err != nil {
		return out, err
	}
	items := seq.content
	for len(items) != 0 {
		item, err := takeElement(&items, 0x30)
		if err != nil {
			return out, err
		}
		body := item.content
		oid, err := takeElement(&body, 6)
		if err != nil {
			return out, err
		}
		var id asn1.ObjectIdentifier
		tail, err := asn1.Unmarshal(oid.full, &id)
		if err != nil || len(tail) != 0 {
			return out, ErrInvalidMaterial
		}
		critical := false
		if len(body) != 0 && body[0] == 1 {
			flag, err := takeElement(&body, 1)
			// FALSE is DEFAULT and must be absent in DER.
			if err != nil || len(flag.content) != 1 || flag.content[0] != 0xff {
				return out, ErrInvalidMaterial
			}
			critical = true
		}
		value, err := takeElement(&body, 4)
		if err != nil || len(body) != 0 {
			return out, ErrInvalidMaterial
		}
		out.extensions = append(out.extensions, rawExtension{id: id, critical: critical, value: value.content})
	}
	return out, nil
}

func minimalInteger(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	return len(b) == 1 || !((b[0] == 0 && b[1]&128 == 0) || (b[0] == 255 && b[1]&128 != 0))
}
